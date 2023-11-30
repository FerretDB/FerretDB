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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestPrepareSelectClause(t *testing.T) {
	t.Parallel()
	table := "table"
	comment := "*/ 1; DROP TABLE " + table + " CASCADE -- "

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		capped        bool
		onlyRecordIDs bool

		expectQuery string
	}{
		"CappedRecordID": {
			capped:        true,
			onlyRecordIDs: true,
			expectQuery: fmt.Sprintf(
				`SELECT %s %s FROM %q`,
				"/* * / 1; DROP TABLE "+table+" CASCADE --  */",
				metadata.RecordIDColumn,
				table,
			),
		},
		"Capped": {
			capped: true,
			expectQuery: fmt.Sprintf(
				`SELECT %s %s, %s FROM %q`,
				"/* * / 1; DROP TABLE "+table+" CASCADE --  */",
				metadata.RecordIDColumn,
				metadata.DefaultColumn,
				table,
			),
		},
		"FullRecord": {
			expectQuery: fmt.Sprintf(
				`SELECT %s %s FROM %q`,
				"/* * / 1; DROP TABLE "+table+" CASCADE --  */",
				metadata.DefaultColumn,
				table,
			),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			query := prepareSelectClause(table, comment, tc.capped, tc.onlyRecordIDs)
			assert.Equal(t, tc.expectQuery, query)
		})
	}
}

func TestPrepareOrderByClause(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		sort   *types.Document
		capped bool

		orderBy string
	}{
		"Ascending": {
			sort:    must.NotFail(types.NewDocument("field", int64(1))),
			orderBy: "",
		},
		"Descending": {
			sort:    must.NotFail(types.NewDocument("field", int64(-1))),
			orderBy: "",
		},
		"SortNil": {
			orderBy: "",
		},
		"Capped": {
			capped:  true,
			orderBy: ` ORDER BY _ferretdb_record_id`,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			orderBy := prepareOrderByClause(tc.sort, tc.capped)

			assert.Equal(t, tc.orderBy, orderBy)
		})
	}
}
