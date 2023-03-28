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

package sqlitedb

import (
	"context"
	"database/sql"
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

type QueryParams struct {
	DB         string
	Collection string
	Comment    string
	Filter     *types.Document
}

func QueryDocuments(ctx context.Context, db *sql.DB, qp *QueryParams) (types.DocumentsIterator, error) {
	iter, err := buildIterator(ctx, db, &iteratorParams{
		schema:  qp.DB,
		table:   qp.Collection,
		comment: qp.Comment,
		filter:  qp.Filter,
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return iter, nil
}

// buildIterator returns an iterator to fetch documents for given iteratorParams.
func buildIterator(ctx context.Context, db *sql.DB, p *iteratorParams) (types.DocumentsIterator, error) {
	var query string

	query += `SELECT json `

	if c := p.comment; c != "" {
		// prevent SQL injections
		c = strings.ReplaceAll(c, "/*", "/ *")
		c = strings.ReplaceAll(c, "*/", "* /")

		query += `/* ` + c + ` */ `
	}

	query += ` FROM ` + p.table

	where, args, err := prepareWhereClause(p.filter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	query += where

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return newIterator(ctx, rows), nil
}

func prepareWhereClause(filter interface{}) (string, []any, error) {
	return "", nil, nil
}
