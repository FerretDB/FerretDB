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

	"github.com/FerretDB/FerretDB/internal/handlers/sqlite/sjson"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func DeleteDocumentsByID(ctx context.Context, qp *QueryParams, ids []any) (int64, error) {
	var p Placeholder
	idsMarshalled := make([]any, len(ids))
	placeholders := make([]string, len(ids))

	for i, id := range ids {
		placeholders[i] = p.Next()
		idsMarshalled[i] = must.NotFail(sjson.MarshalSingleValue(id))
	}

	sqlExpr := `DELETE `

	if qp.Comment != "" {
		qp.Comment = strings.ReplaceAll(qp.Comment, "/*", "/ *")
		qp.Comment = strings.ReplaceAll(qp.Comment, "*/", "* /")

		sqlExpr += `/* ` + qp.Comment + ` */ `
	}

	sqlExpr += `FROM ` + qp.Collection +
		` WHERE  json_extract(json, '$._id') IN (` + strings.Join(placeholders, ", ") + `)`

	db, err := sql.Open("sqlite3", qp.DB)
	if err != nil {
		return 0, err
	}

	tag, err := db.ExecContext(ctx, sqlExpr, idsMarshalled...)
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected()
}
