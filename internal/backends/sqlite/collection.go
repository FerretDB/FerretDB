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
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlitelib "modernc.org/sqlite/lib"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// collection implements backends.Collection interface.
type collection struct {
	r      *metadata.Registry
	dbName string
	name   string
}

// newCollection creates a new Collection.
func newCollection(r *metadata.Registry, dbName, name string) backends.Collection {
	return backends.CollectionContract(&collection{
		r:      r,
		dbName: dbName,
		name:   name,
	})
}

// Query implements backends.Collection interface.
func (c *collection) Query(ctx context.Context, params *backends.QueryParams) (*backends.QueryResult, error) {
	db := c.r.DatabaseGetExisting(ctx, c.dbName)
	if db == nil {
		return &backends.QueryResult{
			Iter: newQueryIterator(ctx, nil),
		}, nil
	}

	tableName, exists, err := c.r.GetTableName(ctx, c.dbName, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !exists {
		return &backends.QueryResult{
			Iter: newQueryIterator(ctx, nil),
		}, nil
	}

	query := fmt.Sprintf(`SELECT _ferretdb_sjson FROM %q`, tableName)

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		// No such table, return empty result.
		var e *sqlite.Error
		if errors.As(err, &e) && e.Code() == sqlitelib.SQLITE_ERROR {
			return &backends.QueryResult{Iter: newQueryIterator(ctx, nil)}, nil
		}

		return nil, lazyerrors.Error(err)
	}

	return &backends.QueryResult{
		Iter: newQueryIterator(ctx, rows),
	}, nil
}

// Insert implements backends.Collection interface.
func (c *collection) Insert(ctx context.Context, params *backends.InsertParams) (*backends.InsertResult, error) {
	if _, err := c.r.CollectionCreate(ctx, c.dbName, c.name); err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/2750

	tableName, _, err := c.r.GetTableName(ctx, c.dbName, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	db := c.r.DatabaseGetExisting(ctx, c.dbName)
	query := fmt.Sprintf(`INSERT INTO %q (_ferretdb_sjson) VALUES (?)`, tableName)

	var res backends.InsertResult

	for {
		_, d, err := params.Iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		doc, ok := d.(*types.Document)
		if !ok {
			panic(fmt.Sprintf("expected document, got %T", d))
		}

		b, err := sjson.Marshal(doc)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if _, err = db.ExecContext(ctx, query, b); err != nil {
			return nil, lazyerrors.Error(err)
		}

		res.Inserted++
	}

	return &res, nil
}

// Update implements backends.Collection interface.
func (c *collection) Update(ctx context.Context, params *backends.UpdateParams) (*backends.UpdateResult, error) {
	db := c.r.DatabaseGetExisting(ctx, c.dbName)
	if db == nil {
		return nil, lazyerrors.Errorf("no database %q", c.dbName)
	}

	tableName, exists, err := c.r.GetTableName(ctx, c.dbName, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var res backends.UpdateResult

	if !exists {
		return &res, nil
	}

	query := fmt.Sprintf(`UPDATE %q SET _ferretdb_sjson = ? WHERE _ferretdb_sjson -> '$._id' = ?`, tableName)

	iter := params.Docs.Iterator()
	defer iter.Close()

	for {
		_, d, err := iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		doc, ok := d.(*types.Document)
		if !ok {
			panic(fmt.Sprintf("expected document, got %T", d))
		}

		id, _ := doc.Get("_id")
		must.NotBeZero(id)
		docArg := must.NotFail(sjson.Marshal(doc))
		idArg := string(must.NotFail(sjson.MarshalSingleValue(id)))

		r, err := db.ExecContext(ctx, query, docArg, idArg)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		rowsAffected, err := r.RowsAffected()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		res.Updated += rowsAffected
	}

	return &res, nil
}

// Delete implements backends.Collection interface.
func (c *collection) Delete(ctx context.Context, params *backends.DeleteParams) (*backends.DeleteResult, error) {
	db := c.r.DatabaseGetExisting(ctx, c.dbName)
	if db == nil {
		return &backends.DeleteResult{Deleted: 0}, nil
	}

	tableName := c.r.CollectionToTable(c.name)

	query := fmt.Sprintf(`DELETE FROM %q WHERE _ferretdb_sjson -> '$._id' = ?`, tableName)

	var deleted int64

	for _, id := range params.IDs {
		idArg := string(must.NotFail(sjson.MarshalSingleValue(id)))

		res, err := db.ExecContext(ctx, query, idArg)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		deleted += rowsAffected
	}

	return &backends.DeleteResult{
		Deleted: deleted,
	}, nil
}

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)
