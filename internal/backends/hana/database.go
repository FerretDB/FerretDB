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

package hana

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// database implements backends.Database.
type database struct {
	hdb    *fsql.DB
	schema string
}

// newDatabase creates a new Database.
func newDatabase(hdb *fsql.DB, name string) backends.Database {
	return backends.DatabaseContract(&database{
		hdb:    hdb,
		schema: name,
	})
}

// Collection implements backends.Database interface.
func (db *database) Collection(name string) (backends.Collection, error) {
	return newCollection(db.hdb, db.schema, name), nil
}

// ListCollections implements backends.Database interface.
func (db *database) ListCollections(
	ctx context.Context,
	params *backends.ListCollectionsParams,
) (*backends.ListCollectionsResult, error) {
	sqlStmt := fmt.Sprintf("SELECT TABLE_NAME FROM M_TABLES"+
		" WHERE SCHEMA_NAME = '%s' AND TABLE_TYPE = 'COLLECTION'", db.schema,
	)

	rows, err := db.hdb.QueryContext(ctx, sqlStmt)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer rows.Close()

	// HANATODO add proper limits for collection sizes.
	var res []backends.CollectionInfo

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, lazyerrors.Error(err)
		}
		ci := backends.CollectionInfo{
			Name:            name,
			CappedSize:      math.MaxInt64,
			CappedDocuments: math.MaxInt64,
		}

		res = append(res, ci)
	}

	if err = rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })

	return &backends.ListCollectionsResult{
		Collections: res,
	}, nil
}

// CreateCollection implements backends.Database interface.
func (db *database) CreateCollection(ctx context.Context, params *backends.CreateCollectionParams) error {
	exists, err := collectionExists(ctx, db.hdb, db.schema, params.Name)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if exists {
		return backends.NewError(backends.ErrorCodeCollectionAlreadyExists, err)
	}

	err = CreateCollection(ctx, db.hdb, db.schema, params.Name)

	return getHanaErrorIfExists(err)
}

// DropCollection implements backends.Database interface.
func (db *database) DropCollection(ctx context.Context, params *backends.DropCollectionParams) error {
	err := dropCollection(ctx, db.hdb, db.schema, params.Name)
	return getHanaErrorIfExists(err)
}

// RenameCollection implements backends.Database interface.
func (db *database) RenameCollection(ctx context.Context, params *backends.RenameCollectionParams) error {
	exists, err := collectionExists(ctx, db.hdb, db.schema, params.OldName)
	if err != nil {
		return getHanaErrorIfExists(err)
	}

	if !exists {
		return lazyerrors.Errorf("old database %q or collection %q does not exist", db.schema, params.OldName)
	}

	sqlStmt := fmt.Sprintf("RENAME COLLECTION %q.%q to %q", db.schema, params.OldName, params.NewName)

	_, err = db.hdb.ExecContext(ctx, sqlStmt)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// Stats implements backends.Database interface.
func (db *database) Stats(ctx context.Context, params *backends.DatabaseStatsParams) (*backends.DatabaseStatsResult, error) {
	// Todo: should we load unloaded schemas?

	queryCountDocuments := "SELECT COALESCE(SUM(RECORD_COUNT),0) FROM M_TABLES " +
		"WHERE TABLE_TYPE = 'COLLECTION' AND SCHEMA_NAME = '%s'"

	queryCountDocuments = fmt.Sprintf(queryCountDocuments, db.schema)

	rowCount := db.hdb.QueryRowContext(ctx, queryCountDocuments)

	var countDocuments int64
	if err := rowCount.Scan(&countDocuments); err != nil {
		return nil, lazyerrors.Error(err)
	}

	querySizeTotal := "SELECT COALESCE(SUM(TABLE_SIZE),0) FROM M_TABLES " +
		"WHERE TABLE_TYPE = 'COLLECTION' AND SCHEMA_NAME = '%s'"
	querySizeTotal = fmt.Sprintf(querySizeTotal, db.schema)

	rowSizeTotal := db.hdb.QueryRowContext(ctx, querySizeTotal)

	var sizeTotal int64
	if err := rowSizeTotal.Scan(&sizeTotal); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &backends.DatabaseStatsResult{
		CountDocuments: countDocuments,
		SizeTotal:      sizeTotal,
	}, nil
}

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
