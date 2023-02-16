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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// SetDocumentByID sets a document by its ID.
// If the document is not valid, it returns *types.ValidationError.
//
// TODO replace QueryParams with own type.
func SetDocumentByID(ctx context.Context, tx pgx.Tx, qp *QueryParams, id any, doc *types.Document) (int64, error) {
	if err := doc.ValidateData(); err != nil {
		return 0, err
	}

	table, err := getMetadata(ctx, tx, qp.DB, qp.Collection)
	if err != nil {
		return 0, err
	}

	sql := "UPDATE "

	if qp.Comment != "" {
		qp.Comment = strings.ReplaceAll(qp.Comment, "/*", "/ *")
		qp.Comment = strings.ReplaceAll(qp.Comment, "*/", "* /")

		sql += `/* ` + qp.Comment + ` */ `
	}

	sql += pgx.Identifier{qp.DB, table}.Sanitize() + " SET _jsonb = $1 WHERE _jsonb->'_id' = $2"

	tag, err := tx.Exec(ctx, sql, must.NotFail(pjson.Marshal(doc)), must.NotFail(pjson.MarshalSingleValue(id)))
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
