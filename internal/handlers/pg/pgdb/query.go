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

package pgdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// FetchedDocs is a struct that contains a list of documents and an error.
// It is used in the fetched channel returned by QueryDocuments.
type FetchedDocs struct {
	Docs []*types.Document
	Err  error
}

// QueryParams represents options/parameters used for SQL query.
type QueryParams struct {
	// Query filter for possible pushdown; may be ignored in part or entirely.
	Filter     *types.Document
	Sort       *types.Document
	Limit      int64 // 0 does not apply limit to the query
	DB         string
	Collection string
	Comment    string
	Explain    bool
}

// Explain returns SQL EXPLAIN results for given query parameters.
//
// It returns (possibly wrapped) ErrTableNotExist if database or collection does not exist.
func Explain(ctx context.Context, tx pgx.Tx, qp *QueryParams) (*types.Document, QueryResults, error) {
	var res QueryResults

	table, err := newMetadataStorage(tx, qp.DB, qp.Collection).getTableName(ctx)
	if err != nil {
		return nil, res, lazyerrors.Error(err)
	}

	var iter types.DocumentsIterator
	iter, res, err = buildIterator(ctx, tx, &iteratorParams{
		schema:    qp.DB,
		table:     table,
		comment:   qp.Comment,
		explain:   qp.Explain,
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
}

// QueryResults represents operations that were done by query builder.
type QueryResults struct {
	FilterPushdown bool
	SortPushdown   bool
	LimitPushdown  bool
}

// QueryDocuments returns an queryIterator to fetch documents for given SQLParams.
// If the collection doesn't exist, it returns an empty iterator and no error.
// If an error occurs, it returns nil and that error, possibly wrapped.
//
// Transaction is not closed by this function. Use iterator.WithClose if needed.
func QueryDocuments(ctx context.Context, tx pgx.Tx, qp *QueryParams) (types.DocumentsIterator, QueryResults, error) {
	table, err := newMetadataStorage(tx, qp.DB, qp.Collection).getTableName(ctx)

	var res QueryResults

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, ErrTableNotExist):
		return newIterator(ctx, nil, new(iteratorParams)), res, nil
	default:
		return nil, res, lazyerrors.Error(err)
	}

	var iter types.DocumentsIterator
	iter, res, err = buildIterator(ctx, tx, &iteratorParams{
		schema:  qp.DB,
		table:   table,
		comment: qp.Comment,
		explain: qp.Explain,
		filter:  qp.Filter,
		sort:    qp.Sort,
		limit:   qp.Limit,
	})
	if err != nil {
		return nil, res, lazyerrors.Error(err)
	}

	return iter, res, nil
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

// convertJSON transforms decoded JSON map[string]any value into *types.Document.
func convertJSON(value any) any {
	switch value := value.(type) {
	case map[string]any:
		d := types.MakeDocument(len(value))
		keys := maps.Keys(value)
		for _, k := range keys {
			v := value[k]
			d.Set(k, convertJSON(v))
		}
		return d

	case []any:
		a := types.MakeArray(len(value))
		for _, v := range value {
			a.Append(convertJSON(v))
		}
		return a

	case nil:
		return types.Null

	case float64, string, bool:
		return value

	default:
		panic(fmt.Sprintf("unsupported type: %[1]T (%[1]v)", value))
	}
}
