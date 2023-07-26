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

	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// Database is a generic interface for all backends for accessing databases.
//
// Database object is expected to be mostly stateless and temporary;
// all state should be in the Backend that created this Database instance.
// Handler can create and destroy Database objects on the fly (but it should Close() them).
// Creating a Database object does not imply the creating of the database itself.
//
// Database methods should be thread-safe.
//
// See databaseContract and its methods for additional details.
type Database interface {
	// TODO remove?
	Close()

	Collection(string) (Collection, error)
	ListCollections(context.Context, *ListCollectionsParams) (*ListCollectionsResult, error)
	CreateCollection(context.Context, *CreateCollectionParams) error
	DropCollection(context.Context, *DropCollectionParams) error
}

// databaseContract implements Database interface.
type databaseContract struct {
	db    Database
	token *resource.Token
}

// DatabaseContract wraps Database and enforces its contract.
//
// All backend implementations should use that function when they create new Database instances.
// The handler should not use that function.
//
// See databaseContract and its methods for additional details.
func DatabaseContract(db Database) Database {
	dbc := &databaseContract{
		db:    db,
		token: resource.NewToken(),
	}
	resource.Track(dbc, dbc.token)

	return dbc
}

// Close marks this Database instance as not being used anymore.
// The implementation may close an associated database connection, decrease a reference counter, etc.
func (dbc *databaseContract) Close() {
	dbc.db.Close()

	resource.Untrack(dbc, dbc.token)
}

// Collection returns a Collection instance for the given valid name.
//
// The collection (or database) does not need to exist.
func (dbc *databaseContract) Collection(name string) (Collection, error) {
	var res Collection
	err := validateCollectionName(name)
	if err == nil {
		res, err = dbc.db.Collection(name)
	}

	checkError(err, ErrorCodeCollectionNameIsInvalid)

	return res, err
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
// Database may not exist; that's not an error.
func (dbc *databaseContract) ListCollections(ctx context.Context, params *ListCollectionsParams) (*ListCollectionsResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := dbc.db.ListCollections(ctx, params)
	checkError(err)

	return res, err
}

// CreateCollectionParams represents the parameters of Database.CreateCollection method.
type CreateCollectionParams struct {
	Name string
}

// CreateCollection creates a new collection with valid name in the database; it should not already exist.
//
// Database may or may not exist; it should be created automatically if needed.
func (dbc *databaseContract) CreateCollection(ctx context.Context, params *CreateCollectionParams) error {
	defer observability.FuncCall(ctx)()

	err := validateCollectionName(params.Name)
	if err == nil {
		err = dbc.db.CreateCollection(ctx, params)
	}

	checkError(err, ErrorCodeCollectionNameIsInvalid, ErrorCodeCollectionAlreadyExists)

	return err
}

// DropCollectionParams represents the parameters of Database.DropCollection method.
type DropCollectionParams struct {
	Name string
}

// DropCollection drops existing collection with valid name in the database.
//
// The errors for non-existing database and non-existing collection are the same (TODO?).
func (dbc *databaseContract) DropCollection(ctx context.Context, params *DropCollectionParams) error {
	defer observability.FuncCall(ctx)()

	err := validateCollectionName(params.Name)
	if err == nil {
		err = dbc.db.DropCollection(ctx, params)
	}

	checkError(err, ErrorCodeCollectionNameIsInvalid, ErrorCodeCollectionDoesNotExist) // TODO: ErrorCodeDatabaseDoesNotExist ?

	return err
}

// check interfaces
var (
	_ Database = (*databaseContract)(nil)
)
