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

// SQLParam represents options/parameters used for SQL query.
type SQLParam struct {
	DB         string
	Collection string
	Comment    string
	Explain    bool
	Filter     *types.Document
}

// Explain returns SQL EXPLAIN results for given query parameters.
func Explain(ctx context.Context, tx pgx.Tx, sp *SQLParam) (*types.Document, error) {
	exists, err := CollectionExists(ctx, tx, sp.DB, sp.Collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !exists {
		return nil, lazyerrors.Error(ErrTableNotExist)
	}

	table, err := getMetadata(ctx, tx, sp.DB, sp.Collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var query string

	if sp.Explain {
		query = `EXPLAIN (VERBOSE true, FORMAT JSON) `
	}

	query += `SELECT _jsonb `

	if c := sp.Comment; c != "" {
		// prevent SQL injections
		c = strings.ReplaceAll(c, "/*", "/ *")
		c = strings.ReplaceAll(c, "*/", "* /")

		query += `/* ` + c + ` */ `
	}

	query += ` FROM ` + pgx.Identifier{sp.DB, table}.Sanitize()

	var args []any

	if sp.Filter != nil {
		var where string

		where, args = prepareWhereClause(sp.Filter)
		query += where
	}

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

// GetDocuments returns an queryIterator to fetch documents for given SQLParams.
// If the collection doesn't exist, it returns an empty iterator and no error.
// If an error occurs, it returns nil and that error, possibly wrapped.
//
// Transaction is not closed by this function. Use iterator.WithClose if needed.
func GetDocuments(ctx context.Context, tx pgx.Tx, sp *SQLParam) (iterator.Interface[int, *types.Document], error) {
	table, err := getMetadata(ctx, tx, sp.DB, sp.Collection)

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, ErrTableNotExist):
		return newIterator(ctx, nil), nil
	default:
		return nil, lazyerrors.Error(err)
	}

	iter, err := buildIterator(ctx, tx, &iteratorParams{
		schema:  sp.DB,
		table:   table,
		comment: sp.Comment,
		explain: sp.Explain,
		filter:  sp.Filter,
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

	where, args := prepareWhereClause(must.NotFail(types.NewDocument("_id", id)))
	query += where

	var b []byte
	err := tx.QueryRow(ctx, query, args...).Scan(&b)

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

	var args []any

	if p.filter != nil {
		var where string

		where, args = prepareWhereClause(p.filter)
		query += where
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return newIterator(ctx, rows), nil
}

// prepareWhereClause adds WHERE clause with given filters to the query and returns the query and arguments.
func prepareWhereClause(sqlFilters *types.Document) (string, []any) {
	var filters []string
	var args []any
	var p Placeholder

	for k, v := range sqlFilters.Map() {
		keyOperator := "->" // keyOperator is the operator that is used to access the field. (->/#>)

		var key any = k   // key can be either a string '"v"' or PostgreSQL path '{v,foo}'
		var prefix string // prefix is the first key in path, if the filter key is not a path - the prefix is empty

		if k != "" {
			// skip $comment
			if k[0] == '$' {
				continue
			}

			// If the key is in dot notation use path operator (#>)
			if path := types.NewPathFromString(k); path.Len() > 1 {
				keyOperator = "#>"
				key = path.Slice()     // '{v,foo}'
				prefix = path.Prefix() // 'v'
			}
		}

		// Handle _id with a simpler query, as it can't be an array
		if k == "_id" || prefix == "_id" {
			switch v := v.(type) {
			case *types.Document, *types.Array, types.Binary, bool, time.Time, types.NullType, types.Regex, types.Timestamp:
				// type not supported for pushdown
				// TODO $eq and $ne https://github.com/FerretDB/FerretDB/issues/1840
				// TODO $gt and $lt https://github.com/FerretDB/FerretDB/issues/1875
			case float64, string, types.ObjectID, int32, int64:
				// Select if value under the key is equal to provided value.
				sql := `((_jsonb%[1]s%[2]s)::jsonb = %[3]s)`

				// operator is -> for non-array, and #> for array
				// placeholder p.Next() returns SQL argument references such as $1, $2 to prevent SQL injections.
				// placeholder $1 is used for field key or it's path,
				// placeholder $2 is used for field value v.
				filters = append(filters, fmt.Sprintf(sql, keyOperator, p.Next(), p.Next()))
				args = append(args, key, string(must.NotFail(pjson.MarshalSingleValue(v))))

			default:
				panic(fmt.Sprintf("Unexpected type of value: %v", v))
			}

			continue
		}

		switch v := v.(type) {
		case *types.Document, *types.Array, types.Binary, bool, time.Time, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown
			// TODO $eq and $ne https://github.com/FerretDB/FerretDB/issues/1840
			// TODO $gt and $lt https://github.com/FerretDB/FerretDB/issues/1875
			continue

		case float64, string, types.ObjectID, int32, int64:
			// Select if value under the key is equal to provided value.
			// If the value under the key is not equal to v,
			// but the value under the key k is an array - select if it contains the value equal to v.
			sql := `((_jsonb%[1]s%[2]s)::jsonb = %[3]s) OR (_jsonb%[1]s%[2]s)::jsonb @> %[3]s`

			// operator is -> for non-array, and #> for array
			// placeholder p.Next() returns SQL argument references such as $1, $2 to prevent SQL injections.
			// placeholder $1 is used for field key or it's path,
			// placeholder $2 is used for field value v.
			filters = append(filters, fmt.Sprintf(sql, keyOperator, p.Next(), p.Next()))
			args = append(args, key, string(must.NotFail(pjson.MarshalSingleValue(v))))

		default:
			panic(fmt.Sprintf("Unexpected type of value: %v", v))
		}
	}

	var query string

	if len(filters) > 0 {
		query = ` WHERE ` + strings.Join(filters, " AND ")
	}

	return query, args
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
