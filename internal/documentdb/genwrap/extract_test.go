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

package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

func TestExtract(t *testing.T) {
	uri := testutil.PostgreSQLURI(t)

	t.Parallel()

	ctx := testutil.Ctx(t)

	rows, err := Extract(ctx, uri, []string{"documentdb_core", "documentdb_api"})
	require.NoError(t, err)
	require.NotZero(t, rows)

	expected := map[string]any{
		"specific_schema":    "documentdb_api",
		"specific_name":      "aggregate_cursor_first_page_19119",
		"routine_name":       "aggregate_cursor_first_page",
		"routine_type":       "FUNCTION",
		"routine_data_type":  "record",
		"routine_udt_schema": "pg_catalog",
		"routine_udt_name":   "record",
		"parameter_name":     "database",
		"parameter_mode":     "IN",
		"parameter_default":  nil,
		"data_type":          "text",
		"udt_schema":         "pg_catalog",
		"udt_name":           "text",
	}
	require.Equal(t, expected, rows[0])

	expected = map[string]any{
		"specific_schema":    "documentdb_api",
		"specific_name":      "aggregate_cursor_first_page_19119",
		"routine_name":       "aggregate_cursor_first_page",
		"routine_type":       "FUNCTION",
		"routine_data_type":  "record",
		"routine_udt_schema": "pg_catalog",
		"routine_udt_name":   "record",
		"parameter_name":     "commandspec",
		"parameter_mode":     "IN",
		"parameter_default":  nil,
		"data_type":          "USER-DEFINED",
		"udt_schema":         "documentdb_core",
		"udt_name":           "bson",
	}
	require.Equal(t, expected, rows[1])
}
