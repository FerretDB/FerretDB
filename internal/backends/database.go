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

import "context"

// Database is a generic interface for all backends for accessing databases.
//
// Database object is expected to be stateless and temporary;
// all state should be in the Backend that created this Database instance.
// Handler can create and destroy Database objects on the fly.
// Creating a Database object does not imply the creating of the database itself.
//
// Database methods should be thread-safe.
//
// See databaseContract and its methods for additional details.
type Database interface {
	Collection(string) Collection
	ListCollections(context.Context, *ListCollectionsParams) (*ListCollectionsResult, error)
	CreateCollection(context.Context, *CreateCollectionParams) error
	DropCollection(context.Context, *DropCollectionParams) error
}

// DatabaseContract wraps Database and enforces its contract.
//
// All backend implementations should use that function when they create new Database instances.
// The handler should not use that function.
//
// See databaseContract and its methods for additional details.
func DatabaseContract(db Database) Database {
	return &databaseContract{
		db: db,
	}
}

// databaseContract implements Database interface.
type databaseContract struct {
	db Database
}

// Collection returns a Collection instance for the given name.
//
// The collection (or database) does not need to exist; even parameters like name could be invalid.
func (dbc *databaseContract) Collection(name string) Collection {
	return dbc.db.Collection(name)
}

// ListCollectionsParams represents the parameters of Database.ListCollections method.
type ListCollectionsParams struct{}

// ListCollectionsResult represents the results of Database.ListCollections method.
type ListCollectionsResult struct {
	Collections []CollectionInfo
}

// CollectionInfo represents information about a single collection.
type CollectionInfo struct {
	Name string
}

// ListCollections returns information about collections in the database.
//
// Database doesn't have to exist; that's not an error.
//
//nolint:lll // for readability
func (dbc *databaseContract) ListCollections(ctx context.Context, params *ListCollectionsParams) (res *ListCollectionsResult, err error) {
	defer checkError(err)
	res, err = dbc.db.ListCollections(ctx, params)

	return
}

// CreateCollectionParams represents the parameters of Database.CreateCollection method.
type CreateCollectionParams struct {
	Name string
}

// CreateCollection creates a new collection in the database; it should not already exist.
//
// Database may or may not exist; it should be created automatically if needed.
func (dbc *databaseContract) CreateCollection(ctx context.Context, params *CreateCollectionParams) (err error) {
	defer checkError(err, ErrorCodeCollectionAlreadyExists, ErrorCodeCollectionNameIsInvalid)
	err = dbc.db.CreateCollection(ctx, params)

	return
}

// DropCollectionParams represents the parameters of Database.DropCollection method.
type DropCollectionParams struct {
	Name string
}

// DropCollection drops existing collection in the database.
func (dbc *databaseContract) DropCollection(ctx context.Context, params *DropCollectionParams) (err error) {
	defer checkError(err, ErrorCodeCollectionDoesNotExist)
	err = dbc.db.DropCollection(ctx, params)

	return
}

// check interfaces
var (
	_ Database = (*databaseContract)(nil)
)
