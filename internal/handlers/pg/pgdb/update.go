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
func SetDocumentByID(ctx context.Context, tx pgx.Tx, sp *SQLParam, id any, doc *types.Document) (int64, error) {
	if err := doc.ValidateData(true); err != nil {
		return 0, err
	}

	table, err := getMetadata(ctx, tx, sp.DB, sp.Collection)
	if err != nil {
		return 0, err
	}

	sql := "UPDATE "

	if sp.Comment != "" {
		sp.Comment = strings.ReplaceAll(sp.Comment, "/*", "/ *")
		sp.Comment = strings.ReplaceAll(sp.Comment, "*/", "* /")

		sql += `/* ` + sp.Comment + ` */ `
	}

	sql += pgx.Identifier{sp.DB, table}.Sanitize() + " SET _jsonb = $1 WHERE _jsonb->'_id' = $2"

	tag, err := tx.Exec(ctx, sql, must.NotFail(pjson.Marshal(doc)), must.NotFail(pjson.MarshalSingleValue(id)))
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
