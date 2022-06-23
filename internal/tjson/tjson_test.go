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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestMarshalUnmarshal(t *testing.T) {
	expected, err := types.NewDocument(
		"_id", types.ObjectID{0x00, 0x01, 0x02, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c},
		"string", "foo",
		"int32", int32(42),
		"int64", int64(123),
		"binary", types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
	)
	require.NoError(t, err)

	actualSchema, err := DocumentSchema(expected)
	require.NoError(t, err)

	expectedSchema := &Schema{
		Type: Object,
		Properties: map[string]*Schema{
			"$k":     {Type: Array, Items: stringSchema},
			"_id":    objectIDSchema,
			"string": stringSchema,
			"int32":  int32Schema,
			"int64":  int64Schema,
			"binary": binarySchema,
		},
		PrimaryKey: []string{"_id"},
	}
	assert.Equal(t, actualSchema, expectedSchema)

	actualB, err := Marshal(expected)
	require.NoError(t, err)
	actualB = testutil.IndentJSON(t, actualB)

	expectedB := testutil.IndentJSON(t, []byte(`{
		"$k": ["_id", "string", "int32", "int64", "binary"],
		"_id": "AAECBAUGBwgJCgsM",
		"string": "foo",
		"int32": 42,
		"int64": 123,
		"binary": {"$b": "Qg==", "s": 128}
	}`))
	assert.Equal(t, string(expectedB), string(actualB))

	actual, err := Unmarshal(expectedB, expectedSchema)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)

	assert.Equal(
		t,
		[]byte{0x00, 0x01, 0x02, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c},
		ObjectID(must.NotFail(expected.Get("_id")).(types.ObjectID)),
	)
}
