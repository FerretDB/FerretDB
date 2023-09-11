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
	"strings"

	sqlite3 "modernc.org/sqlite"
	sqlite3lib "modernc.org/sqlite/lib"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
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

	meta := c.r.CollectionGet(ctx, c.dbName, c.name)
	if meta == nil {
		return &backends.QueryResult{
			Iter: newQueryIterator(ctx, nil),
		}, nil
	}

	var whereClause string
	var args []any

	// that logic should exist in one place
	// TODO https://github.com/FerretDB/FerretDB/issues/3235
	if params != nil && params.Filter.Len() == 1 {
		v, _ := params.Filter.Get("_id")
		if v != nil {
			if id, ok := v.(types.ObjectID); ok {
				whereClause = fmt.Sprintf(` WHERE %s = ?`, metadata.IDColumn)
				args = []any{must.NotFail(sjson.MarshalSingleValue(id))}
			}
		}
	}

	q := fmt.Sprintf(`SELECT %s FROM %q`+whereClause, metadata.DefaultColumn, meta.TableName)

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &backends.QueryResult{
		Iter: newQueryIterator(ctx, rows),
	}, nil
}

// InsertAll implements backends.Collection interface.
func (c *collection) InsertAll(ctx context.Context, params *backends.InsertAllParams) (*backends.InsertAllResult, error) {
	if _, err := c.r.CollectionCreate(ctx, c.dbName, c.name); err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/2750

	db := c.r.DatabaseGetExisting(ctx, c.dbName)
	meta := c.r.CollectionGet(ctx, c.dbName, c.name)

	err := db.InTransaction(ctx, func(tx *fsql.Tx) error {
		for _, doc := range params.Docs {
			b, err := sjson.Marshal(doc)
			if err != nil {
				return lazyerrors.Error(err)
			}

			// use batches: INSERT INTO %q %s VALUES (?), (?), (?), ... up to, say, 100 documents
			// TODO https://github.com/FerretDB/FerretDB/issues/3271
			q := fmt.Sprintf(`INSERT INTO %q (%s) VALUES (?)`, meta.TableName, metadata.DefaultColumn)

			if _, err = tx.ExecContext(ctx, q, string(b)); err != nil {
				var se *sqlite3.Error
				if errors.As(err, &se) && se.Code() == sqlite3lib.SQLITE_CONSTRAINT_UNIQUE {
					return backends.NewError(backends.ErrorCodeInsertDuplicateID, err)
				}

				return lazyerrors.Error(err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return new(backends.InsertAllResult), nil
}

// UpdateAll implements backends.Collection interface.
func (c *collection) UpdateAll(ctx context.Context, params *backends.UpdateAllParams) (*backends.UpdateAllResult, error) {
	db := c.r.DatabaseGetExisting(ctx, c.dbName)
	if db == nil {
		return nil, lazyerrors.Errorf("no database %q", c.dbName)
	}

	var res backends.UpdateAllResult
	meta := c.r.CollectionGet(ctx, c.dbName, c.name)
	if meta == nil {
		return &res, nil
	}

	q := fmt.Sprintf(`UPDATE %q SET %s = ? WHERE %s = ?`, meta.TableName, metadata.DefaultColumn, metadata.IDColumn)

	err := db.InTransaction(ctx, func(tx *fsql.Tx) error {
		for _, doc := range params.Docs {
			b, err := sjson.Marshal(doc)
			if err != nil {
				return lazyerrors.Error(err)
			}

			id, _ := doc.Get("_id")
			must.NotBeZero(id)

			arg := must.NotFail(sjson.MarshalSingleValue(id))

			r, err := tx.ExecContext(ctx, q, string(b), arg)
			if err != nil {
				return lazyerrors.Error(err)
			}

			ra, err := r.RowsAffected()
			if err != nil {
				return lazyerrors.Error(err)
			}

			res.Updated += int32(ra)
		}

		return nil
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &res, nil
}

// DeleteAll implements backends.Collection interface.
func (c *collection) DeleteAll(ctx context.Context, params *backends.DeleteAllParams) (*backends.DeleteAllResult, error) {
	db := c.r.DatabaseGetExisting(ctx, c.dbName)
	if db == nil {
		return &backends.DeleteAllResult{Deleted: 0}, nil
	}

	meta := c.r.CollectionGet(ctx, c.dbName, c.name)
	if meta == nil {
		return &backends.DeleteAllResult{Deleted: 0}, nil
	}

	placeholders := make([]string, len(params.IDs))
	args := make([]any, len(params.IDs))

	for i, id := range params.IDs {
		placeholders[i] = "?"
		args[i] = must.NotFail(sjson.MarshalSingleValue(id))
	}

	q := fmt.Sprintf(`DELETE FROM %q WHERE %s IN (%s)`, meta.TableName, metadata.IDColumn, strings.Join(placeholders, ", "))

	res, err := db.ExecContext(ctx, q, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &backends.DeleteAllResult{
		Deleted: int32(ra),
	}, nil
}

// Explain implements backends.Collection interface.
func (c *collection) Explain(ctx context.Context, params *backends.ExplainParams) (*backends.ExplainResult, error) {
	db := c.r.DatabaseGetExisting(ctx, c.dbName)
	if db == nil {
		return &backends.ExplainResult{
			QueryPlanner: must.NotFail(types.NewDocument()),
		}, nil
	}

	meta := c.r.CollectionGet(ctx, c.dbName, c.name)
	if meta == nil {
		return &backends.ExplainResult{
			QueryPlanner: must.NotFail(types.NewDocument()),
		}, nil
	}

	var queryPushdown bool
	var whereClause string
	var args []any

	// that logic should exist in one place
	// TODO https://github.com/FerretDB/FerretDB/issues/3235
	if params != nil && params.Filter.Len() == 1 {
		v, _ := params.Filter.Get("_id")
		if v != nil {
			if id, ok := v.(types.ObjectID); ok {
				queryPushdown = true
				whereClause = fmt.Sprintf(` WHERE %s = ?`, metadata.IDColumn)
				args = []any{must.NotFail(sjson.MarshalSingleValue(id))}
			}
		}
	}

	q := fmt.Sprintf(`EXPLAIN QUERY PLAN SELECT %s FROM %q`+whereClause, metadata.DefaultColumn, meta.TableName)

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer rows.Close()

	queryPlan, err := types.NewArray()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	for rows.Next() {
		var id int32
		var parent int32
		var notused int32
		var detail string

		// SQLite query plan can be interpreted as a tree.
		// Each row of query plan represents a node of this tree,
		// it contains node id, parent id, auxiliary integer field, and a description.
		// See https://www.sqlite.org/eqp.html for further details.
		if err := rows.Scan(&id, &parent, &notused, &detail); err != nil {
			return nil, lazyerrors.Error(err)
		}

		queryPlan.Append(fmt.Sprintf("id=%d parent=%d notused=%d detail=%s", id, parent, notused, detail))
	}

	return &backends.ExplainResult{
		QueryPlanner:  must.NotFail(types.NewDocument("Plan", queryPlan)),
		QueryPushdown: queryPushdown,
	}, nil
}

// Stats implements backends.Collection interface.
func (c *collection) Stats(ctx context.Context, params *backends.CollectionStatsParams) (*backends.CollectionStatsResult, error) {
	db := c.r.DatabaseGetExisting(ctx, c.dbName)
	if db == nil {
		return nil, backends.NewError(
			backends.ErrorCodeDatabaseDoesNotExist,
			lazyerrors.Errorf("no ns %s.%s", c.dbName, c.name),
		)
	}

	coll := c.r.CollectionGet(ctx, c.dbName, c.name)
	if coll == nil {
		return nil, backends.NewError(
			backends.ErrorCodeCollectionDoesNotExist,
			lazyerrors.Errorf("no ns %s.%s", c.dbName, c.name),
		)
	}

	stats, err := relationStats(ctx, db, []*metadata.Collection{coll})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &backends.CollectionStatsResult{
		CountObjects:   stats.countRows,
		CountIndexes:   stats.countIndexes,
		SizeTotal:      stats.sizeTables + stats.sizeIndexes,
		SizeIndexes:    stats.sizeIndexes,
		SizeCollection: stats.sizeTables,
	}, nil
}

// ListIndexes implements backends.Collection interface.
func (c *collection) ListIndexes(ctx context.Context, params *backends.ListIndexesParams) (*backends.ListIndexesResult, error) {
	db := c.r.DatabaseGetExisting(ctx, c.dbName)
	if db == nil {
		return nil, backends.NewError(
			backends.ErrorCodeCollectionDoesNotExist,
			lazyerrors.Errorf("no ns %s.%s", c.dbName, c.name),
		)
	}

	coll := c.r.CollectionGet(ctx, c.dbName, c.name)
	if coll == nil {
		return nil, backends.NewError(
			backends.ErrorCodeCollectionDoesNotExist,
			lazyerrors.Errorf("no ns %s.%s", c.dbName, c.name),
		)
	}

	res := backends.ListIndexesResult{
		Indexes: make([]backends.IndexInfo, len(coll.Settings.Indexes)),
	}

	for i, index := range coll.Settings.Indexes {
		res.Indexes[i] = backends.IndexInfo{
			Name:   index.Name,
			Unique: index.Unique,
			Key:    make([]backends.IndexKeyPair, len(index.Key)),
		}

		for j, key := range index.Key {
			res.Indexes[i].Key[j] = backends.IndexKeyPair{
				Field:      key.Field,
				Descending: key.Descending,
			}
		}
	}

	return &res, nil
}

// CreateIndexes implements backends.Collection interface.
func (c *collection) CreateIndexes(ctx context.Context, params *backends.CreateIndexesParams) (*backends.CreateIndexesResult, error) { //nolint:lll // for readability
	indexes := make([]metadata.IndexInfo, len(params.Indexes))
	for i, index := range params.Indexes {
		indexes[i] = metadata.IndexInfo{
			Name:   index.Name,
			Key:    make([]metadata.IndexKeyPair, len(index.Key)),
			Unique: index.Unique,
		}

		for j, key := range index.Key {
			indexes[i].Key[j] = metadata.IndexKeyPair{
				Field:      key.Field,
				Descending: key.Descending,
			}
		}
	}

	err := c.r.IndexesCreate(ctx, c.dbName, c.name, indexes)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return new(backends.CreateIndexesResult), nil
}

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)
