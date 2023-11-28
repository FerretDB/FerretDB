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
	"testing"
	"time"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareSelectClause(t *testing.T) {
	t.Parallel()
	database := "schema"
	table := "table"

	t.Run("SelectClause", func(t *testing.T) {
		t.Parallel()

		query := prepareSelectClause(database, table)
		assert.Equal(t, "SELECT * FROM \"schema\".\"table\"", query)
	})
}

func TestPrepareWhereClause(t *testing.T) {
	t.Parallel()

	tableName := "testtable"

	objectID := types.ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0xff, 0xff, 0xff}

	for name, tc := range map[string]struct {
		filter   *types.Document
		expected string
	}{
		"EqObjectId": {
			filter:   must.NotFail(types.NewDocument("_id", objectID)),
			expected: " WHERE \"_id\" = '6256c5ba0badc0ffeeffffff'",
		},
		"EqString": {
			filter:   must.NotFail(types.NewDocument("test", "foo")),
			expected: " WHERE \"test\" = 'foo'",
		},
		"EqStringRequiresPrefix": {
			filter:   must.NotFail(types.NewDocument("testtable", "foo")),
			expected: " WHERE \"testtable\".\"testtable\" = 'foo'",
		},
		"EqInt32": {
			filter:   must.NotFail(types.NewDocument("test", int32(123))),
			expected: " WHERE \"test\" = 123",
		},
		"EqInt64": {
			filter:   must.NotFail(types.NewDocument("test", int64(123))),
			expected: " WHERE \"test\" = 123",
		},
		"EqBool": {
			filter:   must.NotFail(types.NewDocument("test", true)),
			expected: " WHERE \"test\" = TO_JSON_BOOLEAN(true)",
		},
		"EqFloat64": {
			filter:   must.NotFail(types.NewDocument("test", float64(123.456))),
			expected: " WHERE \"test\" = 123.456000",
		},
		"EqDatetime": {
			filter:   must.NotFail(types.NewDocument("test", time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))),
			expected: " WHERE \"test\" = 1635761922123",
		},
		"EqOpObjectID": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", objectID)),
			)),
			expected: " WHERE \"v\" = '6256c5ba0badc0ffeeffffff'",
		},
		"EqOpNotObjectID": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", objectID)),
			)),
			expected: " WHERE \"v\" = '6256c5ba0badc0ffeeffffff'",
		},
		"NeObjectId": {
			filter: must.NotFail(types.NewDocument(
				"_id", must.NotFail(types.NewDocument("$ne", objectID)))),
			expected: " WHERE \"_id\" <> '6256c5ba0badc0ffeeffffff'",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := prepareWhereClause(tableName, tc.filter)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestPrepareOrderByClause(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		sort     *backends.SortField
		expected string
	}{
		"DontSort": {
			sort:     nil,
			expected: "",
		},
		"OrderAsc": {
			sort:     &backends.SortField{Key: "test", Descending: false},
			expected: " ORDER BY \"test\" ASC",
		},
		"OrderDesc": {
			sort:     &backends.SortField{Key: "test", Descending: true},
			expected: " ORDER BY \"test\" DESC",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := prepareOrderByClause(tc.sort)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, actual)
		})
	}

}
