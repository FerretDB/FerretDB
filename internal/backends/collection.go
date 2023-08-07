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
	"context"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/observability"
)

// Collection is a generic interface for all backends for accessing collection.
//
// Collection object is expected to be stateless and temporary;
// all state should be in the Backend that created Database instance that created this Collection instance.
// Handler can create and destroy Collection objects on the fly.
// Creating a Collection object does not imply the creation of the database or collection.
//
// Collection methods should be thread-safe.
//
// See collectionContract and its methods for additional details.
type Collection interface {
	Query(context.Context, *QueryParams) (*QueryResult, error)
	Insert(context.Context, *InsertParams) (*InsertResult, error)
	Update(context.Context, *UpdateParams) (*UpdateResult, error)
	Delete(context.Context, *DeleteParams) (*DeleteResult, error)
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
	// nothing for now - no pushdowns yet
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

// InsertParams represents the parameters of Collection.Insert method.
type InsertParams struct {
	// TODO https://github.com/FerretDB/FerretDB/issues/2750
	// that should be types.DocumentIterator
	Iter iterator.Interface[int, any]
}

// InsertResult represents the results of Collection.Insert method.
type InsertResult struct {
	Inserted int64
}

// Insert inserts documents into the collection.
//
// Both database and collection may or may not exist; they should be created automatically if needed.
// TODO https://github.com/FerretDB/FerretDB/issues/3069
func (cc *collectionContract) Insert(ctx context.Context, params *InsertParams) (*InsertResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := cc.c.Insert(ctx, params)
	checkError(err)

	return res, err
}

// UpdateParams represents the parameters of Collection.Update method.
type UpdateParams struct {
	// that should be types.DocumentIterator
	// TODO https://github.com/FerretDB/FerretDB/issues/3079
	Docs *types.Array
}

// UpdateResult represents the results of Collection.Update method.
type UpdateResult struct {
	Updated int64
}

// Update updates documents in collection.
//
// Database or collection may not exist; that's not an error.
func (cc *collectionContract) Update(ctx context.Context, params *UpdateParams) (*UpdateResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := cc.c.Update(ctx, params)
	checkError(err)

	return res, err
}

// DeleteParams represents the parameters of Collection.Delete method.
type DeleteParams struct {
	// TODO https://github.com/FerretDB/FerretDB/issues/3085
	IDs []any
}

// DeleteResult represents the results of Collection.Delete method.
type DeleteResult struct {
	Deleted int64
}

// Delete deletes documents in collection.
//
// Database or collection may not exist; that's not an error.
func (cc *collectionContract) Delete(ctx context.Context, params *DeleteParams) (*DeleteResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := cc.c.Delete(ctx, params)
	checkError(err)

	return res, err
}

// check interfaces
var (
	_ Collection = (*collectionContract)(nil)
)
