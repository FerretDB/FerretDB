// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backends

import (
	"cmp"
	"context"
	"slices"
	"time"

	"go.opentelemetry.io/otel"
	otelcodes "go.opentelemetry.io/otel/codes"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// DefaultIndexName is a name of the index that is created when a collection is created.
// This index defines document's primary key.
const DefaultIndexName = "_id_"

// Collection is a generic interface for all backends for accessing collection.
//
// Collection object should be stateless and temporary;
// all state should be in the Backend that created Database instance that created this Collection instance.
// Handler can create and destroy Collection objects on the fly.
// Creating a Collection object does not imply the creation of the database or collection.
//
// Collection methods should be thread-safe.
//
// See collectionContract and its methods for additional details.
type Collection interface {
	Query(context.Context, *QueryParams) (*QueryResult, error)
	Explain(context.Context, *ExplainParams) (*ExplainResult, error)
	InsertAll(context.Context, *InsertAllParams) (*InsertAllResult, error)
	UpdateAll(context.Context, *UpdateAllParams) (*UpdateAllResult, error)
	DeleteAll(context.Context, *DeleteAllParams) (*DeleteAllResult, error)

	Stats(context.Context, *CollectionStatsParams) (*CollectionStatsResult, error)
	Compact(context.Context, *CompactParams) (*CompactResult, error)

	ListIndexes(context.Context, *ListIndexesParams) (*ListIndexesResult, error)
	CreateIndexes(context.Context, *CreateIndexesParams) (*CreateIndexesResult, error)
	DropIndexes(context.Context, *DropIndexesParams) (*DropIndexesResult, error)
}

// collectionContract implements Collection interface.
type collectionContract struct {
	c Collection
}

// CollectionContract wraps Collection and enforces its contract.
//
// All backend implementations should use that function when they create new Collection instances.
// The handler should not use that function.
//
// See collectionContract and its methods for additional details.
func CollectionContract(c Collection) Collection {
	return &collectionContract{
		c: c,
	}
}

// QueryParams represents the parameters of Collection.Query method.
type QueryParams struct {
	Filter *types.Document
	Sort   *types.Document
	Limit  int64

	OnlyRecordIDs bool
	Comment       string
}

// QueryResult represents the results of Collection.Query method.
type QueryResult struct {
	Iter types.DocumentsIterator
}

// Query executes a query against the collection.
//
// If database or collection does not exist it returns empty iterator.
//
// The passed context should be used for canceling the initial query.
// It also can be used to close the returned iterator and free underlying resources,
// but doing so is not necessary - the handler will do that anyway.
//
// Filter may be ignored, or safely applied partially or entirely.
// Extra documents will be filtered out by the handler.
//
// Sort should have one of the following forms: nil, {}, {"$natural": int64(1)} or {"$natural": int64(-1)}.
// Other field names are not supported.
// If non-empty, it should be applied.
//
// Limit, if non-zero, should be applied.
func (cc *collectionContract) Query(ctx context.Context, params *QueryParams) (*QueryResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Query")
	defer span.End()

	if params == nil {
		params = new(QueryParams)
	}

	if params.Sort.Len() != 0 {
		must.BeTrue(params.Sort.Len() == 1)
		sortValue := params.Sort.Map()["$natural"].(int64)

		if sortValue != -1 && sortValue != 1 {
			panic("sort value must be 1 (for ascending) or -1 (for descending)")
		}
	}

	res, err := cc.c.Query(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err)

	return res, err
}

// ExplainParams represents the parameters of Collection.Explain method.
type ExplainParams struct {
	Filter *types.Document
	Sort   *types.Document
	Limit  int64
}

// ExplainResult represents the results of Collection.Explain method.
type ExplainResult struct {
	QueryPlanner   *types.Document
	FilterPushdown bool
	SortPushdown   bool
	LimitPushdown  bool
}

// Explain return a backend-specific execution plan for the given query.
//
// Database or collection may not exist; that's not an error, it still
// returns the ExplainResult with QueryPlanner.
//
// The ExplainResult's FilterPushdown field is set to true if the backend could have applied the requested filtering
// partially or completely (but safely in any case).
// If it wasn't possible to apply it safely at least partially, that field should be set to false.
//
// The ExplainResult's SortPushdown field is set to true if the backend could have applied the whole requested sorting.
// If it was possible to apply it only partially or not at all, that field should be set to false.
func (cc *collectionContract) Explain(ctx context.Context, params *ExplainParams) (*ExplainResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Explain")
	defer span.End()

	if params == nil {
		params = new(ExplainParams)
	}

	if params.Sort.Len() != 0 {
		must.BeTrue(params.Sort.Len() == 1)
		sortValue := params.Sort.Map()["$natural"].(int64)

		if sortValue != -1 && sortValue != 1 {
			panic("sort value must be 1 (for ascending) or -1 (for descending)")
		}
	}

	res, err := cc.c.Explain(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err)

	return res, err
}

// InsertAllParams represents the parameters of Collection.InsertAll method.
type InsertAllParams struct {
	Docs []*types.Document
}

// InsertAllResult represents the results of Collection.InsertAll method.
type InsertAllResult struct{}

// InsertAll inserts documents into the collection.
//
// The operation should be atomic.
// If some documents cannot be inserted, the operation should be rolled back,
// and the first encountered error should be returned.
//
// All documents are expected to be valid and include _id fields.
// They will be frozen.
//
// Both database and collection may or may not exist; they should be created automatically if needed.
func (cc *collectionContract) InsertAll(ctx context.Context, params *InsertAllParams) (*InsertAllResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "InsertAll")
	defer span.End()

	now := time.Now()
	for _, doc := range params.Docs {
		doc.SetRecordID(types.NextTimestamp(now).Signed())
		doc.Freeze()
	}

	res, err := cc.c.InsertAll(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err, ErrorCodeInsertDuplicateID)

	return res, err
}

// UpdateAllParams represents the parameters of Collection.Update method.
type UpdateAllParams struct {
	Docs []*types.Document
}

// UpdateAllResult represents the results of Collection.Update method.
type UpdateAllResult struct {
	Updated int32
}

// UpdateAll updates documents in collection.
//
// The operation should be atomic.
// If some documents cannot be updated, the operation should be rolled back,
// and the first encountered error should be returned.
//
// All documents are expected to be valid and include _id fields.
// They will be frozen.
//
// Database or collection may not exist; that's not an error.
func (cc *collectionContract) UpdateAll(ctx context.Context, params *UpdateAllParams) (*UpdateAllResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "UpdateAll")
	defer span.End()

	for _, doc := range params.Docs {
		doc.Freeze()
	}

	res, err := cc.c.UpdateAll(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err)

	return res, err
}

// DeleteAllParams represents the parameters of Collection.Delete method.
type DeleteAllParams struct {
	IDs       []any
	RecordIDs []int64
}

// DeleteAllResult represents the results of Collection.Delete method.
type DeleteAllResult struct {
	Deleted int32
}

// DeleteAll deletes documents in collection.
//
// Passed IDs may contain duplicates or point to non-existing documents.
//
// The operation should be atomic.
// If some documents cannot be deleted, the operation should be rolled back,
// and the first encountered error should be returned.
//
// Database or collection may not exist; that's not an error.
func (cc *collectionContract) DeleteAll(ctx context.Context, params *DeleteAllParams) (*DeleteAllResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "DeleteAll")
	defer span.End()

	must.BeTrue((params.IDs == nil) != (params.RecordIDs == nil))

	res, err := cc.c.DeleteAll(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err)

	return res, err
}

// CollectionStatsParams represents the parameters of Collection.Stats method.
type CollectionStatsParams struct {
	Refresh bool
}

// CollectionStatsResult represents the results of Collection.Stats method.
type CollectionStatsResult struct {
	CountDocuments  int64
	SizeTotal       int64
	SizeIndexes     int64
	SizeCollection  int64
	SizeFreeStorage int64
	IndexSizes      []IndexSize
}

// IndexSize represents the name and the size of an index.
type IndexSize struct {
	Name string
	Size int64
}

// Stats returns statistic estimations about the collection.
// All returned values are not exact, but might be more accurate when Stats is called with `Refresh: true`.
//
// The errors for non-existing database and non-existing collection are the same.
func (cc *collectionContract) Stats(ctx context.Context, params *CollectionStatsParams) (*CollectionStatsResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "CollectionStats")
	defer span.End()

	res, err := cc.c.Stats(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err, ErrorCodeCollectionDoesNotExist)

	return res, err
}

// CompactParams represents the parameters of Collection.Compact method.
type CompactParams struct {
	Full bool
}

// CompactResult represents the results of Collection.Compact method.
type CompactResult struct{}

// Compact reduces the disk space collection takes (by defragmenting, removing dead rows, etc)
// and refreshes its statistics.
//
// If full is true, the operation should try to reduce the disk space as much as possible,
// even if collection or the whole database will be locked for some time.
func (cc *collectionContract) Compact(ctx context.Context, params *CompactParams) (*CompactResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Compact")
	defer span.End()

	res, err := cc.c.Compact(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err, ErrorCodeDatabaseDoesNotExist, ErrorCodeCollectionDoesNotExist)

	return res, err
}

// ListIndexesParams represents the parameters of Collection.ListIndexes method.
type ListIndexesParams struct{}

// ListIndexesResult represents the results of Collection.ListIndexes method.
type ListIndexesResult struct {
	Indexes []IndexInfo
}

// IndexInfo represents information about a single index.
type IndexInfo struct {
	Name   string
	Key    []IndexKeyPair
	Unique bool
}

// IndexKeyPair consists of a field name and a sort order that are part of the index.
type IndexKeyPair struct {
	Field      string
	Descending bool
}

// ListIndexes returns a list of collection indexes.
//
// The errors for non-existing database and non-existing collection are the same.
func (cc *collectionContract) ListIndexes(ctx context.Context, params *ListIndexesParams) (*ListIndexesResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "ListIndexes")
	defer span.End()

	res, err := cc.c.ListIndexes(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err, ErrorCodeCollectionDoesNotExist)

	if res != nil && len(res.Indexes) > 0 {
		must.BeTrue(slices.IsSortedFunc(res.Indexes, func(a, b IndexInfo) int {
			return cmp.Compare(a.Name, b.Name)
		}))
	}

	return res, err
}

// CreateIndexesParams represents the parameters of Collection.CreateIndexes method.
type CreateIndexesParams struct {
	Indexes []IndexInfo
}

// CreateIndexesResult represents the results of Collection.CreateIndexes method.
type CreateIndexesResult struct{}

// CreateIndexes creates indexes for the collection.
//
// The operation should be atomic.
// If some indexes cannot be created, the operation should be rolled back,
// and the first encountered error should be returned.
//
// Database or collection may not exist; that's not an error.
func (cc *collectionContract) CreateIndexes(ctx context.Context, params *CreateIndexesParams) (*CreateIndexesResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "CreateIndexes")
	defer span.End()

	res, err := cc.c.CreateIndexes(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err)

	return res, err
}

// DropIndexesParams represents the parameters of Collection.DropIndexes method.
type DropIndexesParams struct {
	Indexes []string
}

// DropIndexesResult represents the results of Collection.DropIndexes method.
type DropIndexesResult struct{}

// DropIndexes drops indexes for the collection.
//
// The operation should be atomic.
// If some indexes cannot be dropped, the operation should be rolled back,
// and the first encountered error should be returned.
//
// Database or collection may not exist; that's not an error.
func (cc *collectionContract) DropIndexes(ctx context.Context, params *DropIndexesParams) (*DropIndexesResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "DropIndexes")
	defer span.End()

	res, err := cc.c.DropIndexes(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err)

	return res, err
}

// check interfaces
var (
	_ Collection = (*collectionContract)(nil)
)
