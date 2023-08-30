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

package sqlite

import (
	"context"
	"database/sql"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// database implements backends.Database interface.
type database struct {
	r    *metadata.Registry
	name string
}

// newDatabase creates a new Database.
func newDatabase(r *metadata.Registry, name string) backends.Database {
	return backends.DatabaseContract(&database{
		r:    r,
		name: name,
	})
}

// Close implements backends.Database interface.
func (db *database) Close() {
	// nothing
}

// Collection implements backends.Database interface.
func (db *database) Collection(name string) (backends.Collection, error) {
	return newCollection(db.r, db.name, name), nil
}

// ListCollections implements backends.Database interface.
//
//nolint:lll // for readability
func (db *database) ListCollections(ctx context.Context, params *backends.ListCollectionsParams) (*backends.ListCollectionsResult, error) {
	list, err := db.r.CollectionList(ctx, db.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := make([]backends.CollectionInfo, len(list))
	for i, name := range list {
		res[i] = backends.CollectionInfo{
			Name: name,
		}
	}

	return &backends.ListCollectionsResult{
		Collections: res,
	}, nil
}

// CreateCollection implements backends.Database interface.
func (db *database) CreateCollection(ctx context.Context, params *backends.CreateCollectionParams) error {
	created, err := db.r.CollectionCreate(ctx, db.name, params.Name)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if !created {
		return backends.NewError(backends.ErrorCodeCollectionAlreadyExists, err)
	}

	return nil
}

// DropCollection implements backends.Database interface.
func (db *database) DropCollection(ctx context.Context, params *backends.DropCollectionParams) error {
	dropped, err := db.r.CollectionDrop(ctx, db.name, params.Name)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if !dropped {
		return backends.NewError(backends.ErrorCodeCollectionDoesNotExist, err)
	}

	return nil
}

// RenameCollection implements backends.Database interface.
func (db *database) RenameCollection(ctx context.Context, params *backends.RenameCollectionParams) error {
	// TODO https://github.com/FerretDB/FerretDB/issues/2760
	panic("not implemented")
}

// Stats implements backends.Database interface.
//
// If the database does not exist, it returns *backends.DBStatsResult filled with zeros for all the fields.
func (db *database) Stats(ctx context.Context, params *backends.StatsParams) (*backends.StatsResult, error) {
	stats := new(backends.StatsResult)

	d := db.r.DatabaseGetExisting(ctx, db.name)
	if d == nil {
		return stats, nil
	}

	// Call ANALYZE to update statistics of tables and indexes,
	// see https://www.sqlite.org/lang_analyze.html.
	q := `ANALYZE`
	if _, err := d.ExecContext(ctx, q); err != nil {
		return nil, lazyerrors.Error(err)
	}

	// Total size is the disk space used by the database,
	// see https://www.sqlite.org/dbstat.html.
	q = `
		SELECT
			SUM(pgsize)
		FROM dbstat`

	err := d.QueryRowContext(ctx, q).Scan(&stats.SizeTotal)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// Count, size and cells of tables exclude sqlite internal tables and FerretDB meta table.
	// It uses number of cells of btree pages to approximate row count, it returns 0 upon overflow pages.
	// See https://www.sqlite.org/schematab.html and
	// also https://www.sqlite.org/dbstat.html.
	q = `
		SELECT
		    COUNT(s.name)             AS CountTables,
			COALESCE(SUM(d.pgsize),0) AS SizeTables,
			COALESCE(SUM(d.ncell),0)  AS CountCells
		FROM sqlite_schema AS s
			LEFT JOIN dbstat AS d ON d.name = s.tbl_name
		WHERE s.type = 'table' AND s.name NOT LIKE :reserved AND s.name <> :metadata`

	args := []any{
		sql.Named("reserved", metadata.ReservedTablePrefix+"%"),
		sql.Named("metadata", metadata.MetadataTableName),
	}
	if err = d.QueryRowContext(ctx, q, args...).Scan(
		&stats.CountCollections,
		&stats.SizeCollections,
		&stats.CountObjects,
	); err != nil {
		return nil, lazyerrors.Error(err)
	}

	// Count and size of indexes exclude sqlite internal indexes.
	q = `
		SELECT
			COUNT(s.name)             AS CountIndexes,
			COALESCE(SUM(d.pgsize),0) AS SizeIndexes
		FROM sqlite_schema AS s
			LEFT JOIN dbstat AS d ON d.name = s.tbl_name
		WHERE s.type = 'index' AND s.name NOT LIKE :reserved`

	args = []any{
		sql.Named("reserved", metadata.ReservedTablePrefix+"%"),
	}
	if err = d.QueryRowContext(ctx, q, args...).Scan(
		&stats.CountIndexes,
		&stats.SizeIndexes,
	); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return stats, nil
}

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
