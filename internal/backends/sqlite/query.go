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

	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
)

// prepareSelectClause returns SELECT clause for default column of provided table name.
//
// For capped table, it returns select clause for recordID column and default column.
func prepareSelectClause(table string, capped bool) string {
	if capped {
		return fmt.Sprintf(`SELECT %s,%s FROM %q`, metadata.RecordIDColumn, metadata.DefaultColumn, table)
	}

	return fmt.Sprintf(`SELECT %s FROM %q`, metadata.DefaultColumn, table)
}
