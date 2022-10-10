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

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// DeleteDocumentsByID deletes documents by given IDs.
func DeleteDocumentsByID(ctx context.Context, tx pgx.Tx, sp *SQLParam, ids []any) (int64, error) {
	table, err := getTableName(ctx, tx, sp.DB, sp.Collection)
	if err != nil {
		return 0, err
	}

	var p Placeholder
	idsMarshalled := make([]any, len(ids))
	placeholders := make([]string, len(ids))

	for i, id := range ids {
		placeholders[i] = p.Next()
		idsMarshalled[i] = must.NotFail(pjson.Marshal(id))
	}

	sql := `DELETE `

	if sp.Comment != "" {
		sp.Comment = strings.ReplaceAll(sp.Comment, "/*", "/ *")
		sp.Comment = strings.ReplaceAll(sp.Comment, "*/", "* /")

		sql += `/* ` + sp.Comment + ` */ `
	}

	sql += `FROM ` + pgx.Identifier{sp.DB, table}.Sanitize() +
		` WHERE _jsonb->'_id' IN (` + strings.Join(placeholders, ", ") + `)`

	tag, err := tx.Exec(ctx, sql, idsMarshalled...)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
