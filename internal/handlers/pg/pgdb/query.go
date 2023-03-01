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

	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
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
	DB         string
	Collection string
	Comment    string
	Explain    bool
}

// Explain returns SQL EXPLAIN results for given query parameters.
func Explain(ctx context.Context, tx pgx.Tx, qp *QueryParams) (*types.Document, error) {
	exists, err := CollectionExists(ctx, tx, qp.DB, qp.Collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !exists {
		return nil, lazyerrors.Error(ErrTableNotExist)
	}

	table, err := getMetadata(ctx, tx, qp.DB, qp.Collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var query string

	if qp.Explain {
		query = `EXPLAIN (VERBOSE true, FORMAT JSON) `
	}

	query += `SELECT _jsonb `

	if c := qp.Comment; c != "" {
		// prevent SQL injections
		c = strings.ReplaceAll(c, "/*", "/ *")
		c = strings.ReplaceAll(c, "*/", "* /")

		query += `/* ` + c + ` */ `
	}

	query += ` FROM ` + pgx.Identifier{qp.DB, table}.Sanitize()

	where, args, err := prepareWhereClause(qp.Filter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	query += where

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	var res *types.Document

	if !rows.Next() {
		return nil, lazyerrors.Error(errors.New("no rows returned from EXPLAIN"))
	}

	var b []byte
	if err = rows.Scan(&b); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var plans []map[string]any
	if err = json.Unmarshal(b, &plans); err != nil {
		return nil, lazyerrors.Error(err)
	}

	if len(plans) == 0 {
		return nil, lazyerrors.Error(errors.New("no execution plan returned"))
	}

	res = convertJSON(plans[0]).(*types.Document)

	if err = rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// QueryDocuments returns an queryIterator to fetch documents for given SQLParams.
// If the collection doesn't exist, it returns an empty iterator and no error.
// If an error occurs, it returns nil and that error, possibly wrapped.
//
// Transaction is not closed by this function. Use iterator.WithClose if needed.
func QueryDocuments(ctx context.Context, tx pgx.Tx, qp *QueryParams) (iterator.Interface[int, *types.Document], error) {
	table, err := getMetadata(ctx, tx, qp.DB, qp.Collection)

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, ErrTableNotExist):
		return newIterator(ctx, nil), nil
	default:
		return nil, lazyerrors.Error(err)
	}

	iter, err := buildIterator(ctx, tx, &iteratorParams{
		schema:  qp.DB,
		table:   table,
		comment: qp.Comment,
		explain: qp.Explain,
		filter:  qp.Filter,
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return iter, nil
}

// queryById returns the first found document by its ID from the given PostgreSQL schema and table.
// If the document is not found, it returns nil and no error.
func queryById(ctx context.Context, tx pgx.Tx, schema, table string, id any) (*types.Document, error) {
	query := `SELECT _jsonb FROM ` + pgx.Identifier{schema, table}.Sanitize()

	where, args, err := prepareWhereClause(must.NotFail(types.NewDocument("_id", id)))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	query += where

	var b []byte
	err = tx.QueryRow(ctx, query, args...).Scan(&b)

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil
	default:
		return nil, lazyerrors.Error(err)
	}

	doc, err := pjson.Unmarshal(b)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// iteratorParams contains parameters for building an iterator.
type iteratorParams struct {
	schema  string
	table   string
	comment string
	explain bool
	filter  *types.Document
}

// buildIterator returns an iterator to fetch documents for given iteratorParams.
func buildIterator(ctx context.Context, tx pgx.Tx, p *iteratorParams) (iterator.Interface[int, *types.Document], error) {
	var query string

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

	where, args, err := prepareWhereClause(p.filter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	query += where

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return newIterator(ctx, rows), nil
}

// prepareWhereClause adds WHERE clause with given filters to the query and returns the query and arguments.
func prepareWhereClause(sqlFilters *types.Document) (string, []any, error) {
	var builder filtersBuilder

	iter := sqlFilters.Iterator()
	defer iter.Close()

	for {
		k, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return "", nil, lazyerrors.Error(err)
		}

		switch {
		case k == "":
			// do nothing
		case k[0] == '$':
			// skip $comment
			continue
		default:
			path, err := types.NewPathFromString(k)
			if err != nil {
				return "", nil, lazyerrors.Error(err)
			}

			// TODO dot notation https://github.com/FerretDB/FerretDB/issues/2069
			if path.Len() > 1 {
				continue
			}
		}

		switch v := v.(type) {
		case *types.Document:
			iter := v.Iterator()
			defer iter.Close()

			for {
				docKey, docVal, err := iter.Next()
				if err != nil {
					if errors.Is(err, iterator.ErrIteratorDone) {
						break
					}

					return "", nil, lazyerrors.Error(err)
				}

				switch docKey {
				case "$eq":
					switch docVal := docVal.(type) {
					case *types.Document, *types.Array, types.Binary, bool,
						time.Time, types.NullType, types.Regex, types.Timestamp:
						// type not supported for pushdown
					case float64, string, types.ObjectID, int32, int64:
						builder.addFilter(k, eq, docVal)

					default:
						panic(fmt.Sprintf("Unexpected type of value: %v", v))
					}

				case "$ne":
					switch docVal := docVal.(type) {
					case *types.Document, *types.Array, types.Binary, bool,
						time.Time, types.NullType, types.Regex, types.Timestamp:
						// type not supported for pushdown
					case float64, string, types.ObjectID, int32, int64:
						builder.addFilter(k, ne, docVal)

					default:
						panic(fmt.Sprintf("Unexpected type of value: %v", v))
					}

				default:
					// TODO $gt and $lt https://github.com/FerretDB/FerretDB/issues/1875
					continue
				}
			}

		case *types.Array, types.Binary, bool, time.Time, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown
			continue

		case float64, string, types.ObjectID, int32, int64:
			builder.addFilter(k, eq, v)

		default:
			panic(fmt.Sprintf("Unexpected type of value: %v", v))
		}
	}

	query, args := builder.generateWhereClause()

	return query, args, nil
}

// builderOperator represents available filtersBuilder operators.
type builderOperator uint8

const (
	eq builderOperator = iota
	ne
)

// filtersBuilder is responsible for building SQL filters.
// It allows to procedurally add simple key-value comparisons,
// and translate them to the single SQL WHERE clause.
type filtersBuilder struct {
	filters []string
	args    []any
	p       Placeholder
}

// addFilter creates SQL filter based on provided key, value and operator.
func (fb *filtersBuilder) addFilter(key string, op builderOperator, val any) {
	// check if values are equal or if the left value contains the right one
	eqOperator := "@>"

	if key == "_id" {
		// check if values are equal
		eqOperator = "="
	}

	switch op {
	case eq:
		// Select if value under the key is equal to provided value.
		sql := `(_jsonb->%[1]s)::jsonb ` + eqOperator + ` %[2]s`

		fb.filters = append(fb.filters, fmt.Sprintf(sql, fb.p.Next(), fb.p.Next()))
		fb.args = append(fb.args, key, string(must.NotFail(pjson.MarshalSingleValue(val))))

	case ne:
		sql := `NOT ( ` +
			// does document contain the key,
			// it is necessary, as NOT won't work correctly if the key does not exist.
			`_jsonb ? %[1]s AND ` +

			// does the value under the key is equal to the filter
			`(_jsonb->%[1]s)::jsonb ` + eqOperator + ` %[2]s AND ` +

			// does the value type is equal to the filter's one
			`(_jsonb->'$s'->'p'->%[1]s->'t')::jsonb = '"` + pjson.GetTypeOfValue(val) +
			`"')`

		fb.filters = append(fb.filters, fmt.Sprintf(sql, fb.p.Next(), fb.p.Next()))
		fb.args = append(fb.args, key, string(must.NotFail(pjson.MarshalSingleValue(val))))
	default:
		panic(fmt.Sprintf("Unexpected builder operator: %v", op))
	}
}

// generateWhereClause generates SQL WHERE clause from created filters.
// It returns sanitized clause and arguments.
func (fb *filtersBuilder) generateWhereClause() (string, []any) {
	var clause string

	if len(fb.filters) > 0 {
		clause = ` WHERE ` + strings.Join(fb.filters, " AND ")
	}

	return clause, fb.args
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
