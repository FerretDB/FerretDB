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

	"go.opentelemetry.io/otel"
	otelcodes "go.opentelemetry.io/otel/codes"

	"github.com/FerretDB/FerretDB/internal/util/must"
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
type ListCollectionsParams struct {
	Name string
}

// ListCollectionsResult represents the results of Database.ListCollections method.
type ListCollectionsResult struct {
	Collections []CollectionInfo
}

// CollectionInfo represents information about a single collection.
type CollectionInfo struct {
	Name            string
	UUID            string
	CappedSize      int64
	CappedDocuments int64
	_               struct{} // prevent unkeyed literals
}

// Capped returns true if collection is capped.
func (ci *CollectionInfo) Capped() bool {
	return ci.CappedSize > 0 // TODO https://github.com/FerretDB/FerretDB/issues/3631
}

// ListCollections returns a list collections in the database sorted by name.
//
// If ListCollectionsParams' Name is not empty, then only the collection with that name should be returned (or an empty list).
//
// Database may not exist; that's not an error.
func (dbc *databaseContract) ListCollections(ctx context.Context, params *ListCollectionsParams) (*ListCollectionsResult, error) {
	ctx, span := otel.Tracer("").Start(ctx, "ListCollections")
	defer span.End()

	res, err := dbc.db.ListCollections(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err)

	if res != nil && len(res.Collections) > 0 {
		must.BeTrue(slices.IsSortedFunc(res.Collections, func(a, b CollectionInfo) int {
			return cmp.Compare(a.Name, b.Name)
		}))

		if params != nil && params.Name != "" {
			must.BeTrue(len(res.Collections) == 1)
			must.BeTrue(res.Collections[0].Name == params.Name)
		}
	}

	return res, err
}

// CreateCollectionParams represents the parameters of Database.CreateCollection method.
type CreateCollectionParams struct {
	Name            string
	CappedSize      int64
	CappedDocuments int64
	_               struct{} // prevent unkeyed literals
}

// Capped returns true if capped collection creation is requested.
func (ccp *CreateCollectionParams) Capped() bool {
	return ccp.CappedSize > 0 // TODO https://github.com/FerretDB/FerretDB/issues/3631
}

// CreateCollection creates a new collection with valid name in the database; it should not already exist.
//
// Database may or may not exist; it should be created automatically if needed.
func (dbc *databaseContract) CreateCollection(ctx context.Context, params *CreateCollectionParams) error {
	ctx, span := otel.Tracer("").Start(ctx, "CreateCollection")
	defer span.End()

	must.BeTrue(params.CappedSize >= 0)
	must.BeTrue(params.CappedDocuments >= 0)

	err := validateCollectionName(params.Name)
	if err == nil {
		err = dbc.db.CreateCollection(ctx, params)
	}

	if err != nil {
		span.SetStatus(otelcodes.Error, "")
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
	ctx, span := otel.Tracer("").Start(ctx, "DropCollection")
	defer span.End()

	err := validateCollectionName(params.Name)
	if err == nil {
		err = dbc.db.DropCollection(ctx, params)
	}

	if err != nil {
		span.SetStatus(otelcodes.Error, "")
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
	ctx, span := otel.Tracer("").Start(ctx, "RenameCollection")
	defer span.End()

	err := validateCollectionName(params.OldName)

	if err == nil {
		err = validateCollectionName(params.NewName)
	}

	if err == nil {
		err = dbc.db.RenameCollection(ctx, params)
	}

	if err != nil {
		span.SetStatus(otelcodes.Error, "")
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
	ctx, span := otel.Tracer("").Start(ctx, "Stats")
	defer span.End()

	res, err := dbc.db.Stats(ctx, params)
	if err != nil {
		span.SetStatus(otelcodes.Error, "")
	}

	checkError(err, ErrorCodeDatabaseDoesNotExist)

	return res, err
}

// check interfaces
var (
	_ Database = (*databaseContract)(nil)
)
