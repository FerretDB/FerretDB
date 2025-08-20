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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"
)

type testCase struct {
	Extracted []map[string]any
	Converted templateData
}

var testCases = map[string]testCase{
	"documentdb_api.binary_extended_version_19140": {
		Extracted: []map[string]any{
			{
				"specific_schema":   "documentdb_api",
				"specific_name":     "binary_extended_version_19140",
				"routine_name":      "binary_extended_version",
				"routine_type":      "FUNCTION",
				"parameter_name":    nil,
				"parameter_mode":    nil,
				"parameter_default": nil,
				"data_type":         nil,
				"routine_data_type": "text",
			},
		},
		Converted: templateData{},
	},
}

func TestExtract(t *testing.T) {
	t.Parallel()

	extracted, err := Extract(testutil.Ctx(t), testutil.PostgreSQLURL(t), []string{
		"documentdb_api",
		"documentdb_api_catalog",
		"documentdb_api_internal",
		"documentdb_core",
	})
	require.NoError(t, err)
	require.NotZero(t, extracted)

	// for _, name := range slices.Sorted(maps.Keys(extracted)) {
	// 	t.Log(name)
	// }

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			expected := testCase.Extracted
			actual := extracted[name]

			assert.NotEmpty(t, expected)
			assert.NotEmpty(t, actual)
			assert.Equal(t, expected, actual)
		})
	}
}
