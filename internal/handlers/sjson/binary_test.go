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

package sjson

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/types"
)

var binaryTestCases = []testCase{{
	name: "foo",
	v: &binaryType{
		Subtype: types.BinaryUser,
		B:       []byte("foo"),
	},
	sch: binDataSchema(types.BinaryUser),
	j:   `"Zm9v"`,
}, {
	name: "empty",
	v: &binaryType{
		Subtype: types.BinaryGeneric,
		B:       []byte{},
	},
	sch: binDataSchema(types.BinaryGeneric),
	j:   `""`,
}, {
	name: "invalid subtype",
	v: &binaryType{
		Subtype: 0xff,
		B:       []byte{},
	},
	sch: binDataSchema(0xff),
	j:   `""`,
}, {
	name: "EOF",
	j:    `{`,
	jErr: `unexpected EOF`,
}, {
	name: "NilSubtype",
	sch: &elem{
		Type:    elemTypeBinData,
		Subtype: nil,
	},
	j:    `"Zm9v"`,
	jErr: `binary subtype in the schema is nil`,
}}

func TestBinary(t *testing.T) {
	t.Parallel()
	testJSON(t, binaryTestCases, func() sjsontype { return new(binaryType) })
}

func FuzzBinary(f *testing.F) {
	fuzzJSON(f, binaryTestCases, func() sjsontype { return new(binaryType) })
}

func BenchmarkBinary(b *testing.B) {
	benchmark(b, binaryTestCases, func() sjsontype { return new(binaryType) })
}
