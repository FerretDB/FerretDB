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
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// QueryDocuments returns a list of documents for given FerretDB database and collection.
func (pgPool *Pool) QueryDocuments(ctx context.Context, db, collection, comment string) ([]*types.Document, error) {
	var res []*types.Document
	err := pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		table, err := getTableName(ctx, tx, db, collection)
		if err != nil {
			return err
		}

		sql := `SELECT _jsonb `
		if comment != "" {
			comment = strings.ReplaceAll(comment, "/*", "/ *")
			comment = strings.ReplaceAll(comment, "*/", "* /")

			sql += `/* ` + comment + ` */ `
		}

		sql += `FROM ` + pgx.Identifier{db, table}.Sanitize()

		rows, err := tx.Query(ctx, sql)
		if err != nil {
			return lazyerrors.Error(err)
		}
		defer rows.Close()

		for rows.Next() {
			var b []byte
			if err := rows.Scan(&b); err != nil {
				return lazyerrors.Error(err)
			}

			doc, err := fjson.Unmarshal(b)
			if err != nil {
				return lazyerrors.Error(err)
			}

			res = append(res, doc.(*types.Document))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
