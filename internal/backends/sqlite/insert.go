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

package sqlite

import (
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// prepareInsertStatement returns a statement and arguments for inserting the given documents.
//
// If capped is true, it returns a statement and arguments for inserting record IDs and documents.
func prepareInsertStatement(tableName string, capped bool, docs []*types.Document) (string, []any, error) {
	var args []any
	rows := make([]string, len(docs))

	for i, doc := range docs {
		b, err := sjson.Marshal(doc)
		if err != nil {
			return "", nil, lazyerrors.Error(err)
		}

		if capped {
			recordID := doc.RecordID()

			/*// SQLite doesn't return an error if the value overflows the limits of the type but truncates it.
			// This is not the desired behavior for us, so we check the value and return an error if needed.
			if recordID > math.MaxInt64 {
				return "", nil, lazyerrors.Errorf("record id %d is too big", recordID)
			}*/

			rows[i] = "(?, ?)"
			args = append(args, recordID, string(b))

			continue
		}

		rows[i] = "(?)"
		args = append(args, string(b))
	}

	columns := metadata.DefaultColumn
	if capped {
		columns = strings.Join([]string{metadata.RecordIDColumn, metadata.DefaultColumn}, ", ")
	}

	return fmt.Sprintf(
		`INSERT INTO %q (%s) VALUES %s`,
		tableName,
		columns,
		strings.Join(rows, ", "),
	), args, nil
}
