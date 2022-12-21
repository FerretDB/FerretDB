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
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/maps"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// FetchedChannelBufSize is the size of the buffer of the channel that is used in QueryDocuments.
	FetchedChannelBufSize = 3
	// FetchedSliceCapacity is the capacity of the slice in FetchedDocs.
	FetchedSliceCapacity = 2
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

// GetDocuments returns an queryIterator to fetch documents for given SQLParams.
// If the collection doesn't exist, it returns an empty iterator and no error.
// If an error occurs, it returns nil and that error, possibly wrapped.
func GetDocuments(ctx context.Context, tx pgx.Tx, sp *SQLParam) (
	iterator.Interface[uint32, *types.Document], error,
) {
	table, err := getSettings(ctx, tx, sp.DB, sp.Collection)

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, ErrTableNotExist):
		return newIterator(ctx, nil), nil
	default:
		return nil, lazyerrors.Error(err)
	}

	it, err := buildIterator(ctx, tx, iteratorParams{
		schema:  sp.DB,
		table:   table,
		explain: sp.Explain,
		comment: sp.Comment,
		filter:  sp.Filter,
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return it, nil
}

// Explain returns SQL EXPLAIN results for given query parameters.
func Explain(ctx context.Context, tx pgx.Tx, sp SQLParam) (*types.Document, error) {
	exists, err := CollectionExists(ctx, tx, sp.DB, sp.Collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !exists {
		return nil, lazyerrors.Error(ErrTableNotExist)
	}

	table, err := getSettings(ctx, tx, sp.DB, sp.Collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	it, err := buildIterator(ctx, tx, iteratorParams{
		schema:  sp.DB,
		table:   table,
		explain: sp.Explain,
		comment: sp.Comment,
		filter:  sp.Filter,
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer it.Close()

	_, doc, err := it.Next()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

type iteratorParams struct {
	schema  string
	table   string
	explain bool
	comment string
	filter  *types.Document
}

// buildIterator builds SELECT or EXPLAIN SELECT query.
//
// It returns the query string and the arguments.
// If schema/database or table/collection does not exist,
// it returns (possibly wrapped) ErrSchemaNotExist or ErrTableNotExist.
func buildIterator(ctx context.Context, tx pgx.Tx, p iteratorParams) (iterator.Interface[uint32, *types.Document], error) {
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

	/*if err != nil {
		if errors.Is(err, ErrTableNotExist) {
			return newIterator(ctx, nil), nil
		}

		return nil, lazyerrors.Error(err)
	}*/

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
		switch k {
		case "_id":
			switch v := v.(type) {
			case types.ObjectID:
				filters = append(filters, fmt.Sprintf(`((_jsonb->'_id')::jsonb = %s)`, p.Next()))

				args = append(args, string(must.NotFail(pjson.MarshalSingleValue(v))))
			}
		default:
			continue
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
			must.NoError(a.Append(convertJSON(v)))
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
