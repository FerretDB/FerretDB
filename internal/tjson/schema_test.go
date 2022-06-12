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

package tjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestSchemaMarshalUnmarshal(t *testing.T) {
	expected := Schema{
		Title:       "users",
		Description: "FerretDB users collection",
		Properties: map[string]*Schema{
			"$k":      {Type: Array, Items: stringSchema},
			"_id":     objectIDSchema,
			"name":    stringSchema,
			"balance": doubleSchema,
			"data":    binarySchema,
		},
		PrimaryKey: []string{"_id"},
	}

	actualB, err := expected.Marshal()
	require.NoError(t, err)
	actualB = testutil.IndentJSON(t, actualB)

	expectedB := testutil.IndentJSON(t, []byte(`{
		"title": "users",
		"description": "FerretDB users collection",
		"properties": {
			"$k": {"type": "array", "items": {"type": "string"}},
			"_id": {"type": "string", "format": "byte"},
			"balance": {"type": "number"},
			"data": {
				"type": "object",
				"properties": {
					"$b": {"type": "string", "format": "byte"},
					"s": {"type": "integer", "format": "int32"}
				}
			},
			"name": {"type": "string"}
		},
		"primary_key": ["_id"]
	}`))
	assert.Equal(t, string(expectedB), string(actualB))

	var actual Schema
	err = actual.Unmarshal(expectedB)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}
