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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"
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
	p, err := c.r.DatabaseGetExisting(ctx, c.dbName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if p == nil {
		return &backends.QueryResult{
			Iter: newQueryIterator(ctx, nil),
		}, nil
	}

	meta, err := c.r.CollectionGet(ctx, c.dbName, c.name)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if meta == nil {
		return &backends.QueryResult{
			Iter: newQueryIterator(ctx, nil),
		}, nil
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3414
	q := fmt.Sprintf(
		`SELECT %s FROM %s`,
		metadata.DefaultColumn,
		pgx.Identifier{c.dbName, meta.TableName}.Sanitize(),
	)

	rows, err := p.Query(ctx, q)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &backends.QueryResult{
		Iter: newQueryIterator(ctx, rows),
	}, nil
}

// InsertAll implements backends.Collection interface.
func (c *collection) InsertAll(ctx context.Context, params *backends.InsertAllParams) (*backends.InsertAllResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3395
	return new(backends.InsertAllResult), nil
}

// UpdateAll implements backends.Collection interface.
func (c *collection) UpdateAll(ctx context.Context, params *backends.UpdateAllParams) (*backends.UpdateAllResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3391
	return new(backends.UpdateAllResult), nil
}

// DeleteAll implements backends.Collection interface.
func (c *collection) DeleteAll(ctx context.Context, params *backends.DeleteAllParams) (*backends.DeleteAllResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3400
	return new(backends.DeleteAllResult), nil
}

// Explain implements backends.Collection interface.
func (c *collection) Explain(ctx context.Context, params *backends.ExplainParams) (*backends.ExplainResult, error) {
	var res backends.ExplainResult

	db, err := c.r.DatabaseGetExisting(ctx, c.dbName)
	if db == nil {
		return &backends.ExplainResult{
			QueryPlanner: must.NotFail(types.NewDocument()),
		}, nil
	}
	// TODO handle err

	meta, err := c.r.CollectionGet(ctx, c.dbName, c.name)
	if meta == nil {
		return &backends.ExplainResult{
			QueryPlanner: must.NotFail(types.NewDocument()),
		}, nil
	}
	// TODO handle err

	var iter types.DocumentsIterator
	iter, res, err = buildIterator(ctx, tx, &iteratorParams{
		schema:    c.dbName,
		table:     meta.TableName,
		explain:   params.Filter,
		filter:    qp.Filter,
		sort:      qp.Sort,
		limit:     qp.Limit,
		unmarshal: unmarshalExplain,
	})
	if err != nil {
		return nil, res, lazyerrors.Error(err)
	}

	defer iter.Close()

	_, plan, err := iter.Next()

	switch {
	case errors.Is(err, iterator.ErrIteratorDone):
		return nil, res, lazyerrors.Error(errors.New("no rows returned from EXPLAIN"))
	case err != nil:
		return nil, res, lazyerrors.Error(err)
	}

	return plan, res, nil
}

// iteratorParams contains parameters for building an iterator.
type iteratorParams struct {
	schema    string
	table     string
	comment   string
	explain   bool
	filter    *types.Document
	sort      *types.Document
	limit     int64
	forUpdate bool                                    // if SELECT FOR UPDATE is needed.
	unmarshal func(b []byte) (*types.Document, error) // if set, iterator uses unmarshal to convert row to *types.Document.
}

// buildIterator returns an iterator to fetch documents for given iteratorParams.
func buildIterator(ctx context.Context, tx pgx.Tx, p *iteratorParams) (types.DocumentsIterator, QueryResults, error) {
	var query string
	var res QueryResults

	if p.explain {
		query = `EXPLAIN (VERBOSE true, FORMAT JSON) `
	}

	query += `SELECT _jsonb `

	if c := p.comment; c != "" {
		// prevent SQL injections
		c = strings.ReplaceAll(c, "/*", "/ *")
		c = strings.ReplaceAll(c, "*/", "* /")

		query += `/* ` + c + ` */ `
	}

	query += ` FROM ` + pgx.Identifier{p.schema, p.table}.Sanitize()

	var placeholder Placeholder

	where, args, err := prepareWhereClause(&placeholder, p.filter)
	if err != nil {
		return nil, res, lazyerrors.Error(err)
	}

	res.FilterPushdown = where != ""

	query += where

	if p.forUpdate {
		query += ` FOR UPDATE`
	}

	if p.sort != nil {
		var sort string
		var sortArgs []any

		sort, sortArgs, err = prepareOrderByClause(&placeholder, p.sort)
		if err != nil {
			return nil, res, lazyerrors.Error(err)
		}

		query += sort
		args = append(args, sortArgs...)

		res.SortPushdown = sort != ""
	}

	if p.limit != 0 {
		query += fmt.Sprintf(` LIMIT %s`, placeholder.Next())
		args = append(args, p.limit)
		res.LimitPushdown = true
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, res, lazyerrors.Error(err)
	}

	return newIterator(ctx, rows, p), res, nil
}

// unmarshalExplain unmarshalls the plan from EXPLAIN postgreSQL command.
// EXPLAIN result is not sjson, so it cannot be unmarshalled by sjson.Unmarshal.
func unmarshalExplain(b []byte) (*types.Document, error) {
	var plans []map[string]any
	if err := json.Unmarshal(b, &plans); err != nil {
		return nil, lazyerrors.Error(err)
	}

	if len(plans) == 0 {
		return nil, lazyerrors.Error(errors.New("no execution plan returned"))
	}

	return convertJSON(plans[0]).(*types.Document), nil
	return new(backends.ExplainResult), nil
}

// Stats implements backends.Collection interface.
func (c *collection) Stats(ctx context.Context, params *backends.CollectionStatsParams) (*backends.CollectionStatsResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3398
	return new(backends.CollectionStatsResult), nil
}

// ListIndexes implements backends.Collection interface.
func (c *collection) ListIndexes(ctx context.Context, params *backends.ListIndexesParams) (*backends.ListIndexesResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3394
	return new(backends.ListIndexesResult), nil
}

// CreateIndexes implements backends.Collection interface.
func (c *collection) CreateIndexes(ctx context.Context, params *backends.CreateIndexesParams) (*backends.CreateIndexesResult, error) { //nolint:lll // for readability
	// TODO https://github.com/FerretDB/FerretDB/issues/3399
	return new(backends.CreateIndexesResult), nil
}

// DropIndexes implements backends.Collection interface.
func (c *collection) DropIndexes(ctx context.Context, params *backends.DropIndexesParams) (*backends.DropIndexesResult, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3397
	return new(backends.DropIndexesResult), nil
}

// check interfaces
var (
	_ backends.Collection = (*collection)(nil)
)
