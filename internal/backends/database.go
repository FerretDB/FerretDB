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

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/observability"
)

// Database is a generic interface for all backends for accessing databases.
//
// Database object should be stateless and temporary;
// all state should be in the Backend that created this Database instance.
// Handler can create and destroy Database objects on the fly.
// Creating a Database object does not imply the creation of the database.
//
// Database methods should be thread-safe.
//
// See databaseContract and its methods for additional details.
type Database interface {
	Collection(string) (Collection, error)
	ListCollections(context.Context, *ListCollectionsParams) (*ListCollectionsResult, error)
	CreateCollection(context.Context, *CreateCollectionParams) error
	DropCollection(context.Context, *DropCollectionParams) error
	RenameCollection(context.Context, *RenameCollectionParams) error

	Stats(context.Context, *DatabaseStatsParams) (*DatabaseStatsResult, error)
}

// databaseContract implements Database interface.
type databaseContract struct {
	db Database
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
	Name            string
	CappedSize      int64 // TODO https://github.com/FerretDB/FerretDB/issues/3458
	CappedDocuments int64 // TODO https://github.com/FerretDB/FerretDB/issues/3458
}

// Capped returns true if collection is capped.
func (ci *CollectionInfo) Capped() bool {
	return ci.CappedSize > 0 || ci.CappedDocuments > 0
}

// ListCollections returns a list collections in the database sorted by name.
//
// Database may not exist; that's not an error.
//
// Contract ensures that returned list is sorted by name.
func (dbc *databaseContract) ListCollections(ctx context.Context, params *ListCollectionsParams) (*ListCollectionsResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := dbc.db.ListCollections(ctx, params)
	checkError(err)

	if res != nil && len(res.Collections) > 0 {
		must.BeTrue(slices.IsSortedFunc(res.Collections, func(a, b CollectionInfo) int {
			return cmp.Compare(a.Name, b.Name)
		}))
	}

	return res, err
}

// CreateCollectionParams represents the parameters of Database.CreateCollection method.
type CreateCollectionParams struct {
	Name            string
	CappedSize      int64 // TODO https://github.com/FerretDB/FerretDB/issues/3458
	CappedDocuments int64 // TODO https://github.com/FerretDB/FerretDB/issues/3458
}

// Capped returns true if capped collection creation is requested.
func (ccp *CreateCollectionParams) Capped() bool {
	return ccp.CappedSize > 0 || ccp.CappedDocuments > 0
}

// CreateCollection creates a new collection with valid name in the database; it should not already exist.
//
// Database may or may not exist; it should be created automatically if needed.
func (dbc *databaseContract) CreateCollection(ctx context.Context, params *CreateCollectionParams) error {
	defer observability.FuncCall(ctx)()

	must.BeTrue(params.CappedSize >= 0)
	must.BeTrue(params.CappedSize%256 == 0)
	must.BeTrue(params.CappedDocuments >= 0)

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
// The errors for non-existing database and non-existing collection are the same.
func (dbc *databaseContract) DropCollection(ctx context.Context, params *DropCollectionParams) error {
	defer observability.FuncCall(ctx)()

	err := validateCollectionName(params.Name)
	if err == nil {
		err = dbc.db.DropCollection(ctx, params)
	}

	checkError(err, ErrorCodeCollectionNameIsInvalid, ErrorCodeCollectionDoesNotExist)

	return err
}

// RenameCollectionParams represents the parameters of Database.RenameCollection method.
type RenameCollectionParams struct {
	OldName string
	NewName string
}

// RenameCollection renames existing collection in the database.
// Both old and new names should be valid.
//
// The errors for non-existing database and non-existing collection are the same.
func (dbc *databaseContract) RenameCollection(ctx context.Context, params *RenameCollectionParams) error {
	defer observability.FuncCall(ctx)()

	err := validateCollectionName(params.OldName)

	if err == nil {
		err = validateCollectionName(params.NewName)
	}

	if err == nil {
		err = dbc.db.RenameCollection(ctx, params)
	}

	checkError(err, ErrorCodeCollectionNameIsInvalid, ErrorCodeCollectionDoesNotExist, ErrorCodeCollectionAlreadyExists)
	return err
}

// DatabaseStatsParams represents the parameters of Database.Stats method.
type DatabaseStatsParams struct {
	Refresh bool
}

// DatabaseStatsResult represents the results of Database.Stats method.
type DatabaseStatsResult struct {
	CountDocuments  int64
	SizeTotal       int64
	SizeIndexes     int64
	SizeCollections int64
	SizeFreeStorage int64
}

// Stats returns statistic estimations about the database.
// All returned values are not exact, but might be more accurate when Stats is called with `Refresh: true`.
func (dbc *databaseContract) Stats(ctx context.Context, params *DatabaseStatsParams) (*DatabaseStatsResult, error) {
	defer observability.FuncCall(ctx)()

	res, err := dbc.db.Stats(ctx, params)
	checkError(err, ErrorCodeDatabaseDoesNotExist)

	return res, err
}

// check interfaces
var (
	_ Database = (*databaseContract)(nil)
)
