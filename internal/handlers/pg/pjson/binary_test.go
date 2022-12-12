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

	"github.com/AlekSi/pointer"

	"github.com/FerretDB/FerretDB/internal/types"
)

var binaryTestCases = []testCase{{
	name: "foo",
	v: &binaryType{
		Subtype: types.BinaryUser,
		B:       []byte("foo"),
	},
	sch: elem{
		Type:    elemTypeBinData,
		Subtype: pointer.To(types.BinaryUser),
	},
	j: `"Zm9v"`,
}, {
	name: "empty",
	v: &binaryType{
		Subtype: types.BinaryGeneric,
		B:       []byte{},
	},
	sch: elem{
		Type:    elemTypeBinData,
		Subtype: pointer.To(types.BinaryGeneric),
	},
	j:      ``,
	canonJ: ``,
}, {
	name: "invalid subtype",
	v: &binaryType{
		Subtype: 0xff,
		B:       []byte{},
	},
	j: ``,
}, {
	name: "EOF",
	j:    `{`,
	jErr: `unexpected EOF`,
}}

func TestBinary(t *testing.T) {
	t.Parallel()
	testJSON(t, binaryTestCases, func() pjsontype { return new(binaryType) })
}

func FuzzBinary(f *testing.F) {
	fuzzJSON(f, binaryTestCases, func() pjsontype { return new(binaryType) })
}

func BenchmarkBinary(b *testing.B) {
	benchmark(b, binaryTestCases, func() pjsontype { return new(binaryType) })
}
