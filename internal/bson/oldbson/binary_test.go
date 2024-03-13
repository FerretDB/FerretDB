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

package oldbson

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
	b: []byte{0x03, 0x00, 0x00, 0x00, 0x80, 0x66, 0x6f, 0x6f},
}, {
	name: "empty",
	v: &binaryType{
		Subtype: types.BinaryGeneric,
		B:       []byte{},
	},
	b: []byte{0x00, 0x00, 0x00, 0x00, 0x00},
}, {
	name: "invalid subtype",
	v: &binaryType{
		Subtype: 0xff,
		B:       []byte{},
	},
	b: []byte{0x00, 0x00, 0x00, 0x00, 0xff},
}, {
	name: "extra JSON fields",
	v: &binaryType{
		Subtype: types.BinaryUser,
		B:       []byte("foo"),
	},
	b: []byte{0x03, 0x00, 0x00, 0x00, 0x80, 0x66, 0x6f, 0x6f},
}, {
	name: "EOF",
	b:    []byte{0x00},
	bErr: `unexpected EOF`,
}}

func TestBinary(t *testing.T) {
	t.Parallel()
	testBinary(t, binaryTestCases, func() bsontype { return new(binaryType) })
}

func FuzzBinary(f *testing.F) {
	fuzzBinary(f, binaryTestCases, func() bsontype { return new(binaryType) })
}

func BenchmarkBinary(b *testing.B) {
	benchmark(b, binaryTestCases, func() bsontype { return new(binaryType) })
}
