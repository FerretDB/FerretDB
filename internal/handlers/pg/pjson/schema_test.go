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
		Keys: []string{"_id", "name", "distance", "data"},
		Properties: map[string]*elem{
			"_id":      objectIDSchema,
			"name":     stringSchema,
			"distance": doubleSchema,
			"data":     binDataSchema(byte(types.BinaryGeneric)),
		},
	}

	actualB, err := expected.Marshal()
	require.NoError(t, err)
	actualB = testutil.IndentJSON(t, actualB)

	expectedB := testutil.IndentJSON(t, []byte(`{
		"$k": ["_id", "balance", "data", "name"],
		"_id": {"type": "string"},
		"distance": {"type": "double"},
		"data": {"type": "binData", "subtype": "generic"},
		"name": {"type": "string"}
	}`))

	assert.Equal(t, string(expectedB), string(actualB))

	var actual schema
	err = actual.unmarshal(expectedB)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)
}
