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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// SetDocumentByID sets a document by its ID.
// If the document is not valid, it returns *types.ValidationError.
func SetDocumentByID(ctx context.Context, tx pgx.Tx, qp *QueryParams, id any, doc *types.Document) (int64, error) {
	if err := doc.ValidateData(); err != nil {
		return 0, err
	}

	table, err := newMetadataStorage(tx, qp.DB, qp.Collection).getTableName(ctx)
	if err != nil {
		return 0, err
	}

	return setById(ctx, tx, qp.DB, table, qp.Comment, id, doc)
}

// setById sets the document by its ID from the given PostgreSQL schema and table.
func setById(ctx context.Context, tx pgx.Tx, schema, table, comment string, id any, doc *types.Document) (int64, error) {
	sql := "UPDATE "

	if comment != "" {
		comment = strings.ReplaceAll(comment, "/*", "/ *")
		comment = strings.ReplaceAll(comment, "*/", "* /")

		sql += `/* ` + comment + ` */ `
	}

	sql += pgx.Identifier{schema, table}.Sanitize() + " SET _jsonb = $1 WHERE _jsonb->'_id' = $2"

	tag, err := tx.Exec(ctx, sql, must.NotFail(sjson.Marshal(doc)), must.NotFail(sjson.MarshalSingleValue(id)))
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
