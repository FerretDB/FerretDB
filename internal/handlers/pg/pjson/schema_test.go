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

package pjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestSchemaMarshalUnmarshal(t *testing.T) {
	expected := schema{
		Keys: []string{"_id", "arr", "data", "distance", "name"},
		Properties: map[string]*elem{
			"_id": objectIDSchema,
			"arr": {
				Type: schemaTypeArray,
				Items: []*elem{
					boolSchema,
					dateSchema,
					regexSchema("i"),
					{
						Type: schemaTypeObject,
						Schema: &schema{
							Keys: []string{"bar", "baz", "arr"},
							Properties: map[string]*elem{
								"bar": nullSchema,
								"baz": longSchema,
								"arr": {
									Type: schemaTypeArray,
									Items: []*elem{
										intSchema,
										timestampSchema,
									},
								},
							},
						},
					},
				},
			},
			"data":     binDataSchema(byte(types.BinaryFunction)),
			"distance": doubleSchema,
			"name":     stringSchema,
		},
	}

	actualB, err := expected.Marshal()
	require.NoError(t, err)
	actualB = testutil.IndentJSON(t, actualB)

	expectedB := testutil.IndentJSON(t, []byte(`{
		"$k": ["_id", "arr", "data", "distance", "name"],
		"_id": {"t": "objectId"},
		"arr": {"t": "array", "i": [
			{"t": "bool"},
			{"t": "date"},
			{"t": "regex", "o": "i"},
			{"t": "object", "$s": {
				"$k": ["bar", "baz", "arr"],
				"bar": {"t": "null"},
				"baz": {"t": "long"},
				"arr": {"t": "array", "i": [
					{"t": "int"}, 
					{"t": "timestamp"}
				]}
			}}
		]},
		"data": {"t": "binData", "s": 1},
		"distance": {"t": "double"},
		"name": {"t": "string"}
	}`))

	require.Equal(t, string(expectedB), string(actualB))

	var actual schema
	err = actual.Unmarshal(expectedB)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestSchemaUnmarshal(t *testing.T) {
	var actual schema
	err := actual.Unmarshal([]byte(`{
		"$k": ["_id", "data", "distance", "name"], 
		"_id": {"t": "objectId"},
		"data": {"t": "binData", "s": 1}, 
		"distance": {"t": "double"}, 
		"name": {"t": "string"}
	}`))
	require.NoError(t, err)

	expected := &schema{
		Keys: []string{"_id", "data", "distance", "name"},
		Properties: map[string]*elem{
			"_id":      objectIDSchema,
			"data":     binDataSchema(byte(types.BinaryFunction)),
			"distance": doubleSchema,
			"name":     stringSchema,
		},
	}
	assert.Equal(t, expected, &actual)
}
