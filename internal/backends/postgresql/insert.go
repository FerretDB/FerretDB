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

package postgresql

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// prepareInsertStatement returns a statement and arguments for inserting the given documents
// marshalled in bytes into default column.
//
// If capped is true, the statement and arguments corresponds for inserting record ID and bytes
// into record ID column and default column.
func prepareInsertStatement(schema, tableName string, capped bool, docs []*types.Document) (string, []any, error) {
	var placeholder metadata.Placeholder
	var args []any
	rows := make([]string, len(docs))

	for i, doc := range docs {
		b, err := sjson.Marshal(doc)
		if err != nil {
			return "", nil, lazyerrors.Error(err)
		}

		if capped {
			rows[i] = fmt.Sprintf("(%s, %s)", placeholder.Next(), placeholder.Next())
			args = append(args, doc.RecordID(), string(b))
		}

		rows[i] = fmt.Sprintf("(%s)", placeholder.Next())
		args = append(args, string(b))
	}

	columns := metadata.DefaultColumn
	if capped {
		columns = strings.Join([]string{metadata.RecordIDColumn, metadata.DefaultColumn}, ", ")
	}

	return fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES %s`,
		pgx.Identifier{schema, tableName}.Sanitize(),
		columns,
		strings.Join(rows, ", "),
	), args, nil
}
