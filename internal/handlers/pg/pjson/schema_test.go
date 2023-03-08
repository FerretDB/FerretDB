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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestSchemaMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		json   string
		schema schema
		doc    *types.Document
	}{
		"AllTypes": {
			doc: must.NotFail(types.NewDocument(
				"_id", types.NewObjectID(),
				"arr", must.NotFail(types.NewArray(
					true,
					time.Now(),
					types.Regex{Pattern: "foo$", Options: "i"},
					must.NotFail(types.NewDocument(
						"arr", must.NotFail(types.NewArray(
							int32(42), types.NextTimestamp(time.Now()),
						)),
						"bar", types.Null,
						"baz", int64(42),
					)),
				)),
				"data", types.Binary{B: []byte("foo"), Subtype: types.BinaryGeneric},
				"distance", 1.1,
				"name", "foo",
			)),
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
										"arr": {
											Type: elemTypeArray,
											Items: []*elem{
												intSchema,
												timestampSchema,
											},
										},
										"bar": nullSchema,
										"baz": longSchema,
									},
									Keys: []string{"arr", "bar", "baz"},
								},
							},
						},
					},
					"data":     binDataSchema(types.BinaryGeneric),
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
							"$k": ["arr", "bar", "baz"]
						}}
					]},
					"data": {"t": "binData", "s": 0},
					"distance": {"t": "double"},
					"name": {"t": "string"}
				},
				"$k": ["_id", "arr", "data", "distance", "name"]
			}`,
		},
		"Embedded": {
			doc: must.NotFail(types.NewDocument(
				"obj", must.NotFail(types.NewDocument(
					"arr", must.NotFail(types.NewArray(
						must.NotFail(types.NewDocument()),
						must.NotFail(types.NewDocument("foo", must.NotFail(types.NewArray()))),
					)),
					"empty-arr", must.NotFail(types.NewArray()),
				)))),
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
											Type:   elemTypeObject,
											Schema: new(schema),
										},
										{
											Type: elemTypeObject,
											Schema: &schema{
												Properties: map[string]*elem{
													"foo": {
														Type:  elemTypeArray,
														Items: []*elem{},
													},
												},
												Keys: []string{"foo"},
											},
										},
									},
								},
								"empty-arr": {
									Type:  elemTypeArray,
									Items: []*elem{},
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
								{"t": "object", "$s": {}}, {"t": "object", "$s": {
									"p": {"foo": {"t": "array", "i":[]}}, "$k": ["foo"]
								}}
							]},
							"empty-arr": {"t": "array", "i":[]}				
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

			// Schema unmarshalled from json
			var unm schema
			err := json.Unmarshal([]byte(tc.json), &unm)
			require.NoError(t, err)
			assert.Equal(t, tc.schema, unm)

			// Schema made from doc
			made, err := marshalSchemaForDoc(tc.doc)
			require.NoError(t, err)

			expectedB := testutil.IndentJSON(t, []byte(tc.json))
			actualB := testutil.IndentJSON(t, made)
			require.Equal(t, string(expectedB), string(actualB))
		})
	}
}

func TestGetTypeOfValue1(t *testing.T) {
	for _, tc := range []struct {
		input    any
		expected string
	}{
		{&types.Document{}, "object"},
		{&types.Array{}, "array"},
		{float64(1.1), "double"},
		{"foo", "string"},
		{types.Binary{}, "binData"},
		{types.ObjectID{}, "objectId"},
		{true, "bool"},
		{time.Time{}, "date"},
		{types.NullType{}, "null"},
		{types.Regex{}, "regex"},
		{int32(42), "int"},
		{types.Timestamp(1), "timestamp"},
		{int64(42), "long"},
	} {
		actual := GetTypeOfValue(tc.input)
		assert.Equal(t, tc.expected, actual)
	}
}
