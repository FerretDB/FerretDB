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
	"math"
	"testing"

	"github.com/AlekSi/pointer"
)

var int64TestCases = []testCase{{
	name: "42",
	v:    pointer.To(int64Type(42)),
	b:    []byte{0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
}, {
	name: "zero",
	v:    pointer.To(int64Type(0)),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
}, {
	name: "max int64",
	v:    pointer.To(int64Type(math.MaxInt64)),
	b:    []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
}, {
	name: "min int64",
	v:    pointer.To(int64Type(math.MinInt64)),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
}, {
	name: "EOF",
	b:    []byte{0x00},
	bErr: `unexpected EOF`,
}}

func TestInt64(t *testing.T) {
	t.Parallel()
	testBinary(t, int64TestCases, func() bsontype { return new(int64Type) })
}

func FuzzInt64(f *testing.F) {
	fuzzBinary(f, int64TestCases, func() bsontype { return new(int64Type) })
}

func BenchmarkInt64(b *testing.B) {
	benchmark(b, int64TestCases, func() bsontype { return new(int64Type) })
}
