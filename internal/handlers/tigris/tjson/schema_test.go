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
		Title: "users",
		Properties: map[string]*Schema{
			"$k":      {Type: Array, Items: stringSchema},
			"_id":     objectIDSchema,
			"name":    stringSchema,
			"balance": doubleSchema,
			"data":    binarySchema,
		},
		Type:       Object,
		PrimaryKey: []string{"_id"},
	}

	actualB, err := expected.Marshal()
	require.NoError(t, err)
	actualB = testutil.IndentJSON(t, actualB)

	expectedB := testutil.IndentJSON(t, []byte(`{
		"title": "users",
		"type": "object",
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

func TestSchemaUnmarshal(t *testing.T) {
	var actual Schema
	err := actual.Unmarshal([]byte(`{"properties": {"foo": {"type": "object"}}}`))
	assert.NoError(t, err)
	expected := &Schema{
		Type: Object,
		Properties: map[string]*Schema{
			"$k": {Type: Array, Items: stringSchema},
			"foo": {
				Type: Object,
				Properties: map[string]*Schema{
					"$k": {Type: Array, Items: stringSchema},
				},
			},
		},
	}
	assert.Equal(t, expected, &actual)
}

func TestSchemaEqual(t *testing.T) {
	t.Parallel()

	caseInt64Schema := Schema{
		Type:   Integer,
		Format: Int64,
	}
	caseIntEmptySchema := Schema{
		Type:   Integer,
		Format: EmptyFormat,
	}
	caseDoubleSchema := Schema{
		Type:   Number,
		Format: Double,
	}
	caseDoubleEmptySchema := Schema{
		Type:   Number,
		Format: EmptyFormat,
	}
	caseObjectSchema := Schema{
		Type: Object,
		Properties: map[string]*Schema{
			"a":  stringSchema,
			"42": &caseIntEmptySchema,
		},
	}
	caseObjectSchemaEqual := Schema{
		Type: Object,
		Properties: map[string]*Schema{
			"42": &caseIntEmptySchema,
			"a":  stringSchema,
		},
	}
	caseObjectSchemaNotEqual := Schema{
		Type: Object,
		Properties: map[string]*Schema{
			"42": &caseIntEmptySchema,
			"a":  boolSchema,
		},
	}
	caseObjectSchemaKeyMissing := Schema{
		Type: Object,
		Properties: map[string]*Schema{
			"42": &caseIntEmptySchema,
			"b":  stringSchema,
		},
	}
	caseObjectSchemaEmpty := Schema{
		Type:       Object,
		Properties: map[string]*Schema{},
	}
	caseArrayDoubleSchema := Schema{
		Type:  Array,
		Items: &caseDoubleSchema,
	}
	caseArrayDoubleEmptySchema := Schema{
		Type:  Array,
		Items: &caseDoubleEmptySchema,
	}
	caseArrayObjectsSchema := Schema{
		Type:  Array,
		Items: &caseObjectSchema,
	}
	caseArrayObjectsSchemaEqual := Schema{
		Type:  Array,
		Items: &caseObjectSchemaEqual,
	}
	caseArrayObjectsSchemaNotEqual := Schema{
		Type:  Array,
		Items: &caseObjectSchemaNotEqual,
	}

	for name, tc := range map[string]struct {
		s        *Schema
		other    *Schema
		expected bool
	}{
		"StringString": {
			s:        stringSchema,
			other:    stringSchema,
			expected: true,
		},
		"StringNumber": {
			s:        stringSchema,
			other:    doubleSchema,
			expected: false,
		},
		"NumberString": {
			s:        doubleSchema,
			other:    stringSchema,
			expected: false,
		},
		"EmptyInt64": {
			s:        &caseIntEmptySchema,
			other:    &caseInt64Schema,
			expected: true,
		},
		"Int64Empty": {
			s:        &caseInt64Schema,
			other:    &caseIntEmptySchema,
			expected: true,
		},
		"Int64Int32": {
			s:        &caseInt64Schema,
			other:    int32Schema,
			expected: false,
		},
		"EmptyInt32": {
			s:        &caseIntEmptySchema,
			other:    int32Schema,
			expected: false,
		},
		"DoubleEmpty": {
			s:        &caseDoubleSchema,
			other:    &caseDoubleEmptySchema,
			expected: true,
		},
		"ObjectsEqual": {
			s:        &caseObjectSchema,
			other:    &caseObjectSchemaEqual,
			expected: true,
		},
		"ObjectsNotEqual": {
			s:        &caseObjectSchemaEqual,
			other:    &caseObjectSchemaNotEqual,
			expected: false,
		},
		"ObjectsKeyMissing": {
			s:        &caseObjectSchema,
			other:    &caseObjectSchemaKeyMissing,
			expected: false,
		},
		"ObjectsEmpty": {
			s:        &caseObjectSchema,
			other:    &caseObjectSchemaEmpty,
			expected: false,
		},
		"ArrayDouble": {
			s:        &caseArrayDoubleSchema,
			other:    &caseArrayDoubleEmptySchema,
			expected: true,
		},
		"ArrayObjects": {
			s:        &caseArrayObjectsSchema,
			other:    &caseArrayObjectsSchemaEqual,
			expected: true,
		},
		"ArrayObjectsNotEqual": {
			s:        &caseArrayObjectsSchemaNotEqual,
			other:    &caseArrayObjectsSchemaEqual,
			expected: false,
		},
		"ArrayObjectsDouble": {
			s:        &caseArrayObjectsSchema,
			other:    &caseArrayDoubleSchema,
			expected: false,
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.s.Equal(tc.other))
		})
	}
}
