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

var regexTestCases = []testCase{{
	name: "normal",
	v:    pointer.To(regexType{Pattern: "hoffman", Options: "i"}),
	b:    []byte{0x68, 0x6f, 0x66, 0x66, 0x6d, 0x61, 0x6e, 0x00, 0x69, 0x00},
}, {
	name: "empty",
	v:    pointer.To(regexType{Pattern: "", Options: ""}),
	b:    []byte{0x00, 0x00},
}, {
	name: "EOF",
	b:    []byte{0x00},
	bErr: `EOF`,
}}

func TestRegex(t *testing.T) {
	t.Parallel()
	testBinary(t, regexTestCases, func() bsontype { return new(regexType) })
}

func FuzzRegex(f *testing.F) {
	fuzzBinary(f, regexTestCases, func() bsontype { return new(regexType) })
}

func BenchmarkRegex(b *testing.B) {
	benchmark(b, regexTestCases, func() bsontype { return new(regexType) })
}
