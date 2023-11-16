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
	"database/sql"
	"fmt"
	"math"
	"sort"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// database implements backends.Database.
type database struct {
	hdb    *sql.DB
	schema string
}

// newDatabase creates a new Database.
func newDatabase(hdb *sql.DB, name string) backends.Database {
	return backends.DatabaseContract(&database{
		hdb: hdb,
	})
}

// Collection implements backends.Database interface.
func (db *database) Collection(name string) (backends.Collection, error) {
	return newCollection(db.hdb, db.schema, name), nil
}

// ListCollections implements backends.Database interface.
func (db *database) ListCollections(ctx context.Context, params *backends.ListCollectionsParams) (*backends.ListCollectionsResult, error) {
	sqlStmt := "SELECT TABLE_NAME FROM M_TABLES" +
		" WHERE SCHEMA_NAME = $1 AND TABLE_TYPE = 'COLLECTION'"
	rows, err := db.hdb.QueryContext(ctx, sqlStmt, db.schema)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	// TODO: add propper limits for collection sizes.
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
	sql := fmt.Sprintf("CREATE COLLECTION %q.%q", db.schema, params.Name)

	_, err := db.hdb.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}

// DropCollection implements backends.Database interface.
func (db *database) DropCollection(ctx context.Context, params *backends.DropCollectionParams) error {
	sql := fmt.Sprintf("DROP COLLECTION %q.%q", db.schema, params.Name)

	_, err := db.hdb.ExecContext(ctx, sql)

	return getHanaErrorIfExists(err)
}

// RenameCollection implements backends.Database interface.
func (db *database) RenameCollection(ctx context.Context, params *backends.RenameCollectionParams) error {
	// Todo check if collection exists
	sqlStmt := fmt.Sprintf("RENAME COLLECTION %s.%s to %s", db.schema, params.OldName, params.NewName)
	_, err := db.hdb.ExecContext(ctx, sqlStmt)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// Stats implements backends.Database interface.
func (db *database) Stats(ctx context.Context, params *backends.DatabaseStatsParams) (*backends.DatabaseStatsResult, error) {
	if params == nil {
		params = new(backends.DatabaseStatsParams)
	}

	// Todo: should we load unloaded schemas?

	queryCountDocuments := "SELECT COALESCE(SUM(RECORD_COUNT),0) FROM M_TABLES " +
		"WHERE TABLE_TYPE = 'COLLECTION' AND SCHEMA_NAME = $1"

	rowCount := db.hdb.QueryRowContext(ctx, queryCountDocuments, db.schema)
	var countDocuments int64
	if err := rowCount.Scan(&countDocuments); err != nil {
		return nil, lazyerrors.Error(err)
	}

	querySizeTotal := "SELECT COALESCE(SUM(TABLE_SIZE),0) FROM M_TABLES " +
		"WHERE TABLE_TYPE = 'COLLECTION' AND SCHEMA_NAME = $1"

	rowSizeTotal := db.hdb.QueryRowContext(ctx, querySizeTotal, db.schema)

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
