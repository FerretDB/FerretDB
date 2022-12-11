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
	t.Parallel()

	for name, tc := range map[string]struct {
		json   string
		schema schema
	}{
		"AllTypes": {
			schema: schema{
				Properties: map[string]*elem{
					"_id": objectIDSchema,
					"arr": {
						Type: elemTypeArray,
						Items: []*elem{
							boolSchema,
							dateSchema,
							regexSchema("i"),
							{
								Type: elemTypeObject,
								Schema: &schema{
									Properties: map[string]*elem{
										"bar": nullSchema,
										"baz": longSchema,
										"arr": {
											Type: elemTypeArray,
											Items: []*elem{
												intSchema,
												timestampSchema,
											},
										},
									},
									Keys: []string{"bar", "baz", "arr"},
								},
							},
						},
					},
					"data":     binDataSchema(byte(types.BinaryFunction)),
					"distance": doubleSchema,
					"name":     stringSchema,
				},
				Keys: []string{"_id", "arr", "data", "distance", "name"},
			},
			json: `{
				"p": {
					"_id": {"t": "objectId"},
					"arr": {"t": "array", "i": [
						{"t": "bool"},
						{"t": "date"},
						{"t": "regex", "o": "i"},
						{"t": "object", "$s": {
							"p": {
								"arr": {"t": "array", "i": [
									{"t": "int"}, 
									{"t": "timestamp"}
								]},
								"bar": {"t": "null"},
								"baz": {"t": "long"}
							},
							"$k": ["bar", "baz", "arr"]
						}}
					]},
					"data": {"t": "binData", "s": 1},
					"distance": {"t": "double"},
					"name": {"t": "string"}
				},
				"$k": ["_id", "arr", "data", "distance", "name"]
			}`,
		},
		"Embedded": {
			schema: schema{
				Properties: map[string]*elem{
					"obj": {
						Type: elemTypeObject,
						Schema: &schema{
							Properties: map[string]*elem{
								"arr": {
									Type: elemTypeArray,
									Items: []*elem{
										{
											Type: elemTypeObject,
										},
										{
											Type: elemTypeObject,
											Schema: &schema{
												Properties: map[string]*elem{
													"foo": {
														Type: elemTypeArray,
													},
												},
												Keys: []string{"foo"},
											},
										},
									},
								},
								"empty-arr": {
									Type: elemTypeArray,
								},
							},
							Keys: []string{"arr", "empty-arr"},
						},
					},
				},
				Keys: []string{"obj"},
			},
			json: `{
				"p": {
					"obj": {"t": "object", "$s": {
						"p": {	
							"arr": {"t": "array", "i": [
								{"t": "object"}, {"t": "object", "$s": {
									"p": {"foo": {"t": "array"}}, "$k": ["foo"]
								}}
							]},
							"empty-arr": {"t": "array"}				
						},
						"$k": ["arr", "empty-arr"]
					}}
				},
				"$k": ["obj"]
			}`,
		},
	} {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actualB, err := tc.schema.Marshal()
			require.NoError(t, err)
			actualB = testutil.IndentJSON(t, actualB)

			expectedB := testutil.IndentJSON(t, []byte(tc.json))
			require.Equal(t, string(expectedB), string(actualB))

			var actual schema
			err = actual.Unmarshal(expectedB)
			require.NoError(t, err)

			assert.Equal(t, tc.schema, actual)
		})
	}
}
