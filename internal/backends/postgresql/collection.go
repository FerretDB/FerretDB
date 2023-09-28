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
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/resource"
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

// newIterator returns a new queryIterator for the given pgx.Rows.
//
// Iterator's Close method closes rows.
//
// Nil rows are possible and return already done iterator.
func newIterator(ctx context.Context, rows pgx.Rows, p *iteratorParams) types.DocumentsIterator {
	unmarshalFunc := p.unmarshal
	if unmarshalFunc == nil {
		unmarshalFunc = sjson.Unmarshal
	}

	iter := &queryIterator{
		ctx:       ctx,
		unmarshal: unmarshalFunc,
		rows:      rows,
		token:     resource.NewToken(),
	}
	resource.Track(iter, iter.token)

	return iter
}

// prepareWhereClause adds WHERE clause with given filters to the query and returns the query and arguments.
func prepareWhereClause(p *Placeholder, sqlFilters *types.Document) (string, []any, error) {
	var filters []string
	var args []any

	iter := sqlFilters.Iterator()
	defer iter.Close()

	// iterate through root document
	for {
		rootKey, rootVal, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return "", nil, lazyerrors.Error(err)
		}

		// don't pushdown $comment, it's attached to query in handlers
		if strings.HasPrefix(rootKey, "$") {
			continue
		}

		path, err := types.NewPathFromString(rootKey)

		var pe *types.PathError

		switch {
		case err == nil:
			// Handle dot notation.
			// TODO https://github.com/FerretDB/FerretDB/issues/2069
			if path.Len() > 1 {
				continue
			}
		case errors.As(err, &pe):
			// ignore empty key error, otherwise return error
			if pe.Code() != types.ErrPathElementEmpty {
				return "", nil, lazyerrors.Error(err)
			}
		default:
			panic("Invalid error type: PathError expected")
		}

		switch v := rootVal.(type) {
		case *types.Document:
			iter := v.Iterator()
			defer iter.Close()

			// iterate through subdocument, as it may contain operators
			for {
				k, v, err := iter.Next()
				if err != nil {
					if errors.Is(err, iterator.ErrIteratorDone) {
						break
					}

					return "", nil, lazyerrors.Error(err)
				}

				switch k {
				case "$eq":
					if f, a := filterEqual(p, rootKey, v); f != "" {
						filters = append(filters, f)
						args = append(args, a...)
					}

				case "$ne":
					sql := `NOT ( ` +
						// does document contain the key,
						// it is necessary, as NOT won't work correctly if the key does not exist.
						`_jsonb ? %[1]s AND ` +
						// does the value under the key is equal to filter value
						`_jsonb->%[1]s @> %[2]s AND ` +
						// does the value type is equal to the filter's one
						`_jsonb->'$s'->'p'->%[1]s->'t' = '"%[3]s"' )`

					switch v := v.(type) {
					case *types.Document, *types.Array, types.Binary,
						types.NullType, types.Regex, types.Timestamp:
						// type not supported for pushdown

					case float64, bool, int32, int64:
						filters = append(filters, fmt.Sprintf(sql, p.Next(), p.Next(), sjson.GetTypeOfValue(v)))
						args = append(args, rootKey, v)

					case string, types.ObjectID, time.Time:
						filters = append(filters, fmt.Sprintf(sql, p.Next(), p.Next(), sjson.GetTypeOfValue(v)))
						args = append(args, rootKey, string(must.NotFail(sjson.MarshalSingleValue(v))))

					default:
						panic(fmt.Sprintf("Unexpected type of value: %v", v))
					}

				default:
					// $gt and $lt
					// TODO https://github.com/FerretDB/FerretDB/issues/1875
					continue
				}
			}

		case *types.Array, types.Binary, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown

		case float64, string, types.ObjectID, bool, time.Time, int32, int64:
			if f, a := filterEqual(p, rootKey, v); f != "" {
				filters = append(filters, f)
				args = append(args, a...)
			}

		default:
			panic(fmt.Sprintf("Unexpected type of value: %v", v))
		}
	}

	var filter string
	if len(filters) > 0 {
		filter = ` WHERE ` + strings.Join(filters, " AND ")
	}

	return filter, args, nil
}

// filterEqual returns the proper SQL filter with arguments that filters documents
// where the value under k is equal to v.
func filterEqual(p *Placeholder, k string, v any) (filter string, args []any) {
	// Select if value under the key is equal to provided value.
	sql := `_jsonb->%[1]s @> %[2]s`

	switch v := v.(type) {
	case *types.Document, *types.Array, types.Binary,
		types.NullType, types.Regex, types.Timestamp:
		// type not supported for pushdown

	case float64:
		// If value is not safe double, fetch all numbers out of safe range.
		switch {
		case v > types.MaxSafeDouble:
			sql = `_jsonb->%[1]s > %[2]s`
			v = types.MaxSafeDouble

		case v < -types.MaxSafeDouble:
			sql = `_jsonb->%[1]s < %[2]s`
			v = -types.MaxSafeDouble
		default:
			// don't change the default eq query
		}

		filter = fmt.Sprintf(sql, p.Next(), p.Next())
		args = append(args, k, v)

	case string, types.ObjectID, time.Time:
		// don't change the default eq query
		filter = fmt.Sprintf(sql, p.Next(), p.Next())
		args = append(args, k, string(must.NotFail(sjson.MarshalSingleValue(v))))

	case bool, int32:
		// don't change the default eq query
		filter = fmt.Sprintf(sql, p.Next(), p.Next())
		args = append(args, k, v)

	case int64:
		maxSafeDouble := int64(types.MaxSafeDouble)

		// If value cannot be safe double, fetch all numbers out of the safe range.
		switch {
		case v > maxSafeDouble:
			sql = `_jsonb->%[1]s > %[2]s`
			v = maxSafeDouble

		case v < -maxSafeDouble:
			sql = `_jsonb->%[1]s < %[2]s`
			v = -maxSafeDouble
		default:
			// don't change the default eq query
		}

		filter = fmt.Sprintf(sql, p.Next(), p.Next())
		args = append(args, k, v)

	default:
		panic(fmt.Sprintf("Unexpected type of value: %v", v))
	}

	return
}

// Placeholder stores the number of the relevant placeholder of the query.
type Placeholder int

// Next increases the identifier value for the next variable in the PostgreSQL query.
func (p *Placeholder) Next() string {
	*p++
	return "$" + strconv.Itoa(int(*p))
}

// prepareOrderByClause adds ORDER BY clause with given sort document and returns the query and arguments.
func prepareOrderByClause(p *Placeholder, sort *types.Document) (string, []any, error) {
	iter := sort.Iterator()
	defer iter.Close()

	var key string
	var order types.SortType

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return "", nil, lazyerrors.Error(err)
		}

		// Skip sorting if there are more than one sort parameters
		if order != 0 {
			return "", nil, nil
		}

		order, err = common.GetSortType(k, v)
		if err != nil {
			return "", nil, err
		}

		key = k
	}

	// Skip sorting dot notation
	if strings.ContainsRune(key, '.') {
		return "", nil, nil
	}

	var sqlOrder string

	switch order {
	case types.Descending:
		sqlOrder = "DESC"
	case types.Ascending:
		sqlOrder = "ASC"
	case 0:
		return "", nil, nil
	default:
		panic(fmt.Sprint("forbidden order:", order))
	}

	return fmt.Sprintf(" ORDER BY _jsonb->%s %s", p.Next(), sqlOrder), []any{key}, nil
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
