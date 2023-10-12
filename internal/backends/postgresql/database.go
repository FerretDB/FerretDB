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

	"github.com/AlekSi/pointer"

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

	stats, err := collectionsStats(ctx, p, db.name, list)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// Total size is the disk space used by all the relations in the given schema,
	// including tables, indexes, TOAST data also includes FerretDB metadata relations.
	// See https://www.postgresql.org/docs/15/functions-admin.html#FUNCTIONS-ADMIN-DBOBJECT.
	q := `
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

	if schemaSize == nil {
		schemaSize = pointer.ToInt64(0)
	}

	return &backends.DatabaseStatsResult{
		CountCollections: int64(len(list)),
		CountObjects:     stats.countRows,
		CountIndexes:     stats.countIndexes,
		SizeTotal:        *schemaSize,
		SizeIndexes:      stats.sizeIndexes,
		SizeCollections:  stats.sizeTables,
	}, nil
}

// check interfaces
var (
	_ backends.Database = (*database)(nil)
)
