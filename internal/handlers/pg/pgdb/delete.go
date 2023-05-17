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

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// DeleteDocumentsByID deletes documents by given IDs.
func DeleteDocumentsByID(ctx context.Context, tx pgx.Tx, qp *QueryParams, ids []any) (int64, error) {
	table, err := newMetadataStorage(tx, qp.DB, qp.Collection).getTableName(ctx)
	if err != nil {
		return 0, err
	}

	return deleteByIDs(ctx, tx, execDeleteParams{
		schema:  qp.DB,
		table:   table,
		comment: qp.Comment,
	}, ids,
	)
}

// execDeleteParams describes the parameters for deleting from a table.
type execDeleteParams struct {
	schema  string // pg schema name
	table   string // pg table name
	comment string // comment to add to the query
}

// deleteByIDs deletes documents by given IDs.
func deleteByIDs(ctx context.Context, tx pgx.Tx, d execDeleteParams, ids []any) (int64, error) {
	var p Placeholder
	idsMarshalled := make([]any, len(ids))
	placeholders := make([]string, len(ids))

	for i, id := range ids {
		placeholders[i] = p.Next()
		idsMarshalled[i] = must.NotFail(sjson.MarshalSingleValue(id))
	}

	sql := `DELETE `

	if d.comment != "" {
		d.comment = strings.ReplaceAll(d.comment, "/*", "/ *")
		d.comment = strings.ReplaceAll(d.comment, "*/", "* /")

		sql += `/* ` + d.comment + ` */ `
	}

	sql += `FROM ` + pgx.Identifier{d.schema, d.table}.Sanitize() +
		` WHERE _jsonb->'_id' IN (` + strings.Join(placeholders, ", ") + `)`

	tag, err := tx.Exec(ctx, sql, idsMarshalled...)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
