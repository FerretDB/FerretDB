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

	"github.com/FerretDB/FerretDB/internal/types"
)

var binarySchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"$b": map[string]any{"type": "string", "format": "byte"},   // binary data
		"s":  map[string]any{"type": "integer", "format": "int32"}, // binary subtype
	},
}

var binaryTestCases = []testCase{{
	name: "foo",
	v: types.Binary{
		Subtype: types.BinaryUser,
		B:       []byte("foo"),
	},
	s: binarySchema,
	j: `{"$b":"Zm9v","s":128}`,
}, {
	name: "empty",
	v: types.Binary{
		Subtype: types.BinaryGeneric,
		B:       []byte{},
	},
	s:      binarySchema,
	j:      `{"$b":""}`,
	canonJ: `{"$b":"","s":0}`,
}, {
	name: "invalid subtype",
	v: types.Binary{
		Subtype: 0xff,
		B:       []byte{},
	},
	s: binarySchema,
	j: `{"$b":"","s":255}`,
}, {
	name: "extra JSON fields",
	v: types.Binary{
		Subtype: types.BinaryUser,
		B:       []byte("foo"),
	},
	s:      binarySchema,
	j:      `{"$b":"Zm9v","s":128,"foo":"bar"}`,
	canonJ: `{"$b":"Zm9v","s":128}`,
	jErr:   `json: unknown field "foo"`,
}, {
	name: "EOF",
	j:    `{`,
	s:    binarySchema,
	jErr: `unexpected EOF`,
}}

func TestBinary(t *testing.T) {
	t.Parallel()
	testJSON(t, binaryTestCases, func() tjsontype { return new(binaryType) })
}

func FuzzBinary(f *testing.F) {
	fuzzJSON(f, binaryTestCases, func() tjsontype { return new(binaryType) })
}

func BenchmarkBinary(b *testing.B) {
	benchmark(b, binaryTestCases, func() tjsontype { return new(binaryType) })
}
