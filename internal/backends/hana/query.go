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

package hana

import (
	"strings"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/types"
)

func prepareSelectClause(schema, table string) (string, []any) {
	args := []any{schema, table}
	return "SELECT * FROM %q.%q", args
}

func prepareWhereClause(filter *types.Document) (string, []any, error) {
	var filters []string
	var args []any

	// iter := filter.Iterator()
	// defer iter.Close()

	whereClause := ""
	if len(filters) > 0 {
		whereClause = " WHERE " + strings.Join(filters, " AND ")
	}

	return whereClause, args, nil
}

func prepareOrderByClause(sort *backends.SortField) (string, []any, error) {
	var args []any
	orderByClause := ""

	return orderByClause, args, nil
}
