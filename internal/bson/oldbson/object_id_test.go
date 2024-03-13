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

var objectIDTestCases = []testCase{{
	name: "normal",
	v:    pointer.To(objectIDType{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}),
	b:    []byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
}, {
	name: "EOF",
	b:    []byte{0x00},
	bErr: `unexpected EOF`,
}}

func TestObjectID(t *testing.T) {
	t.Parallel()
	testBinary(t, objectIDTestCases, func() bsontype { return new(objectIDType) })
}

func FuzzObjectID(f *testing.F) {
	fuzzBinary(f, objectIDTestCases, func() bsontype { return new(objectIDType) })
}

func BenchmarkObjectID(b *testing.B) {
	benchmark(b, objectIDTestCases, func() bsontype { return new(objectIDType) })
}
