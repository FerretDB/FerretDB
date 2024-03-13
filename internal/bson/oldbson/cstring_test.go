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

var cstringTestCases = []testCase{{
	name: "foo",
	v:    pointer.To(CString("foo")),
	b:    []byte{0x66, 0x6f, 0x6f, 0x00},
}, {
	name: "empty",
	v:    pointer.To(CString("")),
	b:    []byte{0x00},
}}

func TestCString(t *testing.T) {
	t.Parallel()
	testBinary(t, cstringTestCases, func() bsontype { return new(CString) })
}

func FuzzCString(f *testing.F) {
	fuzzBinary(f, cstringTestCases, func() bsontype { return new(CString) })
}

func BenchmarkCString(b *testing.B) {
	benchmark(b, cstringTestCases, func() bsontype { return new(CString) })
}
