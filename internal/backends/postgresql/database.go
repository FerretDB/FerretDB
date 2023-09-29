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

package postgresql

import (
	"context"
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"
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
	for i, c := range list {
		res[i] = backends.CollectionInfo{
			Name: c.Name,
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
	c, err := db.r.CollectionGet(ctx, db.name, params.OldName)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if c == nil {
		return backends.NewError(
			backends.ErrorCodeCollectionDoesNotExist,
			lazyerrors.Errorf("old database %q or collection %q does not exist", db.name, params.OldName),
		)
	}

	c, err = db.r.CollectionGet(ctx, db.name, params.NewName)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if c != nil {
		return backends.NewError(
			backends.ErrorCodeCollectionAlreadyExists,
			lazyerrors.Errorf("new database %q and collection %q already exists", db.name, params.NewName),
		)
	}

	renamed, err := db.r.CollectionRename(ctx, db.name, params.OldName, params.NewName)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if !renamed {
		return backends.NewError(backends.ErrorCodeCollectionDoesNotExist, err)
	}

	return nil
}

// Stats implements backends.Database interface.
func (db *database) Stats(ctx context.Context, params *backends.DatabaseStatsParams) (*backends.DatabaseStatsResult, error) {
	var res backends.DatabaseStatsResult

	p, err := db.r.DatabaseGetExisting(ctx, db.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if p == nil {
		return nil, backends.NewError(backends.ErrorCodeDatabaseDoesNotExist, lazyerrors.Errorf("no database %s", db.name))
	}

	list, err := db.r.CollectionList(ctx, db.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	res.CountCollections = int64(len(list))

	// Call ANALYZE to update statistics, the actual statistics are needed to estimate the number of rows in all tables,
	// see https://wiki.postgresql.org/wiki/Count_estimate.
	q := `ANALYZE`
	if _, err := p.Exec(ctx, q); err != nil {
		return nil, lazyerrors.Error(err)
	}

	// Total size is the disk space used by all the relations in the given schema, including tables, indexes and TOAST data.
	// It also includes the size of FerretDB metadata relations.
	//  See also https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-DBOBJECT
	q = `
		SELECT
		    SUM(pg_total_relation_size(quote_ident(schemaname) || '.' || quote_ident(tablename)))
		FROM pg_tables
		WHERE schemaname = $1`
	args := []any{db.name}
	row := p.QueryRow(ctx, q, args...)

	var schemaSize *int64
	if err := row.Scan(&schemaSize); err != nil {
		return nil, lazyerrors.Error(err)
	}

	// If the query gave nil, it means the schema does not exist or empty, no need to check other stats.
	if schemaSize == nil {
		return &res, nil
	}

	res.SizeTotal = *schemaSize

	var placeholder metadata.Placeholder
	placeholders := make([]string, len(list))
	args = []any{db.name}

	placeholder.Next()

	for i, c := range list {
		placeholders[i] = placeholder.Next()
		args = append(args, c.TableName)
	}

	// In this query we select all the tables in the given schema, but we exclude FerretDB metadata table (by reserved prefix).
	q = fmt.Sprintf(`
		SELECT
			COUNT(i.indexname)                       AS CountIndexes,
			COALESCE(SUM(c.reltuples), 0)            AS CountRows,
			COALESCE(SUM(pg_table_size(c.oid)), 0) 	 AS SizeTables,
			COALESCE(SUM(pg_indexes_size(c.oid)), 0) AS SizeIndexes
		FROM pg_tables AS t
			LEFT JOIN pg_class AS c ON c.relname = t.tablename AND c.relnamespace = quote_ident(t.schemaname)::regnamespace
			LEFT JOIN pg_indexes AS i ON i.schemaname = t.schemaname AND i.tablename = t.tablename
		WHERE t.schemaname = $1 AND t.tablename IN (%s)`,
		strings.Join(placeholders, ", "),
	)

	row = p.QueryRow(ctx, q, args...)
	if err := row.Scan(
		&res.CountIndexes, &res.CountObjects, &res.SizeCollections, &res.SizeIndexes,
	); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &res, nil
}

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
