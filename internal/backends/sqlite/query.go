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

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
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

// prepareOrderByClause returns ORDER BY clause.
//
// For capped collection, it returns ORDER BY recordID only if sort field is nil.
func prepareOrderByClause(sort *backends.SortField, capped bool) string {
	// TODO
	// also, this function applies ORDER BY recordID in situation where
	// sort is specified, but not pushed down (for example because of more that 1 sort fields)
	// technically we don't lose anything, apart from processing time on the backend.
	//
	// TODO disabling self pushdown needs to be applied here, that could also solve the above issue
	if sort == nil && capped {
		return fmt.Sprintf(` ORDER BY %s`, metadata.RecordIDColumn)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3181
	return ""
}
