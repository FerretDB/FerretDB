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
//
// It returns (possibly wrapped) ErrTableNotExist if database or collection does not exist.
func Explain(ctx context.Context, tx pgx.Tx, qp *QueryParams) (*types.Document, error) {
	table, err := newMetadata(tx, qp.DB, qp.Collection).getTableName(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	iter, err := buildIterator(ctx, tx, &iteratorParams{
		schema:    qp.DB,
		table:     table,
		comment:   qp.Comment,
		explain:   qp.Explain,
		filter:    qp.Filter,
		unmarshal: unmarshalExplain,
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer iter.Close()

	_, plan, err := iter.Next()

	switch {
	case errors.Is(err, iterator.ErrIteratorDone):
		return nil, lazyerrors.Error(errors.New("no rows returned from EXPLAIN"))
	case err != nil:
		return nil, lazyerrors.Error(err)
	}

	return plan, nil
}

// unmarshalExplain unmarshalls the plan from EXPLAIN postgreSQL command.
// EXPLAIN result is not pjson, so it cannot be unmarshalled by pjson.Unmarshal.
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

// QueryDocuments returns an queryIterator to fetch documents for given SQLParams.
// If the collection doesn't exist, it returns an empty iterator and no error.
// If an error occurs, it returns nil and that error, possibly wrapped.
//
// Transaction is not closed by this function. Use iterator.WithClose if needed.
func QueryDocuments(ctx context.Context, tx pgx.Tx, qp *QueryParams) (iterator.Interface[int, *types.Document], error) {
	table, err := newMetadata(tx, qp.DB, qp.Collection).getTableName(ctx)

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, ErrTableNotExist):
		return newIterator(ctx, nil, new(iteratorParams)), nil
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

// iteratorParams contains parameters for building an iterator.
type iteratorParams struct {
	schema    string
	table     string
	comment   string
	explain   bool
	filter    *types.Document
	forUpdate bool                                    // if SELECT FOR UPDATE is needed.
	unmarshal func(b []byte) (*types.Document, error) // if set, iterator uses unmarshal to convert row to *types.Document.
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

	if p.forUpdate {
		query += ` FOR UPDATE`
	}

	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return newIterator(ctx, rows, p), nil
}

// prepareWhereClause adds WHERE clause with given filters to the query and returns the query and arguments.
func prepareWhereClause(sqlFilters *types.Document) (string, []any, error) {
	var filters []string
	var args []any
	var p Placeholder

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

		var pe *types.DocumentPathError

		switch {
		case err == nil:
			// TODO dot notation https://github.com/FerretDB/FerretDB/issues/2069
			if path.Len() > 1 {
				continue
			}
		case errors.As(err, &pe):
			// ignore empty key error, otherwise return error
			if pe.Code() != types.ErrDocumentPathEmptyKey {
				return "", nil, lazyerrors.Error(err)
			}
		default:
			panic("Invalid error type: DocumentPathError expected ")
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
					switch v := v.(type) {
					case *types.Document, *types.Array, types.Binary,
						types.NullType, types.Regex, types.Timestamp:
						// type not supported for pushdown
					case float64, string, types.ObjectID, bool, time.Time, int32, int64:
						// Select if value under the key is equal to provided value.
						sql := `_jsonb->%[1]s @> %[2]s`

						filters = append(filters, fmt.Sprintf(sql, p.Next(), p.Next()))
						args = append(args, rootKey, string(must.NotFail(pjson.MarshalSingleValue(v))))
					default:
						panic(fmt.Sprintf("Unexpected type of value: %v", v))
					}

				case "$ne":
					switch v := v.(type) {
					case *types.Document, *types.Array, types.Binary,
						types.NullType, types.Regex, types.Timestamp:
						// type not supported for pushdown
					case float64, string, types.ObjectID, bool, time.Time, int32, int64:
						sql := `NOT ( ` +
							// does document contain the key,
							// it is necessary, as NOT won't work correctly if the key does not exist.
							`_jsonb ? %[1]s AND ` +
							// does the value under the key is equal to filter value
							`_jsonb->%[1]s @> %[2]s AND ` +
							// does the value type is equal to the filter's one
							`_jsonb->'$s'->'p'->%[1]s->'t' = '"%[3]s"' )`

						filters = append(filters, fmt.Sprintf(sql, p.Next(), p.Next(), pjson.GetTypeOfValue(v)))
						args = append(args, rootKey, string(must.NotFail(pjson.MarshalSingleValue(v))))
					default:
						panic(fmt.Sprintf("Unexpected type of value: %v", v))
					}

				default:
					// TODO $gt and $lt https://github.com/FerretDB/FerretDB/issues/1875
					continue
				}
			}

		case *types.Array, types.Binary, types.NullType, types.Regex, types.Timestamp:
			// type not supported for pushdown
			continue

		case float64, string, types.ObjectID, bool, time.Time, int32, int64:
			// Select if value under the key is equal to provided value.
			sql := `_jsonb->%[1]s @> %[2]s`

			filters = append(filters, fmt.Sprintf(sql, p.Next(), p.Next()))
			args = append(args, rootKey, string(must.NotFail(pjson.MarshalSingleValue(v))))

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
