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

	"github.com/AlekSi/pointer"
)

var stringTestCases = []testCase{{
	name: "foo",
	v:    pointer.To(stringType("foo")),
	b:    []byte{0x04, 0x00, 0x00, 0x00, 0x66, 0x6f, 0x6f, 0x00},
}, {
	name: "empty",
	v:    pointer.To(stringType("")),
	b:    []byte{0x01, 0x00, 0x00, 0x00, 0x00},
}, {
	name: "zero",
	v:    pointer.To(stringType("\x00")),
	b:    []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00},
}, {
	name: "EOF",
	b:    []byte{0x00},
	bErr: `unexpected EOF`,
}}

func TestString(t *testing.T) {
	t.Parallel()
	testBinary(t, stringTestCases, func() bsontype { return new(stringType) })
}

func FuzzString(f *testing.F) {
	fuzzBinary(f, stringTestCases, func() bsontype { return new(stringType) })
}

func BenchmarkString(b *testing.B) {
	benchmark(b, stringTestCases, func() bsontype { return new(stringType) })
}
