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

package bson

import (
	"math"
	"testing"

	"github.com/AlekSi/pointer"
)

var int32TestCases = []testCase{{
	name: "42",
	v:    pointer.To(int32Type(42)),
	b:    []byte{0x2a, 0x00, 0x00, 0x00},
}, {
	name: "zero",
	v:    pointer.To(int32Type(0)),
	b:    []byte{0x00, 0x00, 0x00, 0x00},
}, {
	name: "max int32",
	v:    pointer.To(int32Type(math.MaxInt32)),
	b:    []byte{0xff, 0xff, 0xff, 0x7f},
}, {
	name: "min int32",
	v:    pointer.To(int32Type(math.MinInt32)),
	b:    []byte{0x00, 0x00, 0x00, 0x80},
}, {
	name: "EOF",
	b:    []byte{0x00},
	bErr: `unexpected EOF`,
}}

func TestInt32(t *testing.T) {
	t.Parallel()
	testBinary(t, int32TestCases, func() bsontype { return new(int32Type) })
}

func FuzzInt32(f *testing.F) {
	fuzzBinary(f, int32TestCases, func() bsontype { return new(int32Type) })
}

func BenchmarkInt32(b *testing.B) {
	benchmark(b, int32TestCases, func() bsontype { return new(int32Type) })
}
