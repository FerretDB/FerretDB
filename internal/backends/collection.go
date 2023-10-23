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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
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
	InsertAll(context.Context, *InsertAllParams) (*InsertAllResult, error)
	UpdateAll(context.Context, *UpdateAllParams) (*UpdateAllResult, error)
	DeleteAll(context.Context, *DeleteAllParams) (*DeleteAllResult, error)
	Explain(context.Context, *ExplainParams) (*ExplainResult, error)

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

// SortField consists of a field name and a sort order that are used in queries.
type SortField struct {
	Key        string
	Descending bool
}

// QueryParams represents the parameters of Collection.Query method.
type QueryParams struct {
	// TODO https://github.com/FerretDB/FerretDB/issues/3235
	Filter        *types.Document
	Sort          *SortField
	Limit         int64  // if 0 no limit pushdown is applied
	OnlyRecordIDs bool   // TODO https://github.com/FerretDB/FerretDB/issues/3490
	Comment       string // TODO https://github.com/FerretDB/FerretDB/issues/3573
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
func (cc *collectionContract) Query(ctx context.Context, params *QueryParams) (*QueryResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := cc.c.Query(ctx, params)
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
	defer observability.FuncCall(ctx)()

	now := time.Now()
	for _, doc := range params.Docs {
		doc.SetRecordID(types.NextTimestamp(now))
		doc.Freeze()
	}

	res, err := cc.c.InsertAll(ctx, params)
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
	defer observability.FuncCall(ctx)()

	for _, doc := range params.Docs {
		doc.Freeze()
	}

	res, err := cc.c.UpdateAll(ctx, params)
	checkError(err)

	return res, err
}

// DeleteAllParams represents the parameters of Collection.Delete method.
type DeleteAllParams struct {
	IDs       []any
	RecordIDs []types.Timestamp
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
	defer observability.FuncCall(ctx)()

	must.BeTrue((params.IDs == nil) != (params.RecordIDs == nil))

	res, err := cc.c.DeleteAll(ctx, params)
	checkError(err)

	return res, err
}

// ExplainParams represents the parameters of Collection.Explain method.
type ExplainParams struct {
	// TODO https://github.com/FerretDB/FerretDB/issues/3235
	Filter *types.Document
	Sort   *SortField
	Limit  int64 // if 0 no limit pushdown is applied
}

// ExplainResult represents the results of Collection.Explain method.
type ExplainResult struct {
	QueryPlanner *types.Document
	// TODO https://github.com/FerretDB/FerretDB/issues/3235
	QueryPushdown bool
	SortPushdown  bool
	LimitPushdown bool
}

// Explain return a backend-specific execution plan for the given query.
//
// Database or collection may not exist; that's not an error, it still
// returns the ExplainResult with QueryPlanner.
func (cc *collectionContract) Explain(ctx context.Context, params *ExplainParams) (*ExplainResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := cc.c.Explain(ctx, params)
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
	defer observability.FuncCall(ctx)()

	res, err := cc.c.Stats(ctx, params)
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
	defer observability.FuncCall(ctx)()

	res, err := cc.c.Compact(ctx, params)
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
	defer observability.FuncCall(ctx)()

	res, err := cc.c.ListIndexes(ctx, params)
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
	defer observability.FuncCall(ctx)()

	res, err := cc.c.CreateIndexes(ctx, params)
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
	defer observability.FuncCall(ctx)()

	res, err := cc.c.DropIndexes(ctx, params)
	checkError(err)

	return res, err
}

// check interfaces
var (
	_ Collection = (*collectionContract)(nil)
)
