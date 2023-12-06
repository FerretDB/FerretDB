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
	"github.com/FerretDB/FerretDB/internal/types"
)

// prepareSelectClause returns SELECT clause for default column of provided table name.
//
// For capped collection with onlyRecordIDs, it returns select clause for recordID column.
//
// For capped collection, it returns select clause for recordID column and default column.
func prepareSelectClause(table, comment string, capped, onlyRecordIDs bool) string {
	if comment != "" {
		comment = strings.ReplaceAll(comment, "/*", "/ *")
		comment = strings.ReplaceAll(comment, "*/", "* /")
		comment = `/* ` + comment + ` */`
	}

	if capped && onlyRecordIDs {
		return fmt.Sprintf(`SELECT %s %s FROM %q`, comment, metadata.RecordIDColumn, table)
	}

	if capped {
		return fmt.Sprintf(`SELECT %s %s, %s FROM %q`, comment, metadata.RecordIDColumn, metadata.DefaultColumn, table)
	}

	return fmt.Sprintf(`SELECT %s %s FROM %q`, comment, metadata.DefaultColumn, table)
}

// prepareOrderByClause returns ORDER BY clause for given sort document.
//
// The provided sort document should be already validated.
// Provided document should only contain a single value.
func prepareOrderByClause(sort *types.Document) string {
	if sort.Len() == 0 {
		return ""
	}

	// the following code could be simplified now

	v, err := sort.Get("$natural")
	if err != nil {
		return ""
	}

	sortOrder := v.(int64)
	if sortOrder != 1 {
		// FIXME support -1 for $natural
		return ""
	}

	return fmt.Sprintf(` ORDER BY %s`, metadata.RecordIDColumn)
}
