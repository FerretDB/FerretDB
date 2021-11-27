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

var int64TestCases = []testCase{{
	name: "42",
	v:    pointer.To(Int64(42)),
	b:    []byte{0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	j:    `{"$l":"42"}`,
}, {
	name: "zero",
	v:    pointer.To(Int64(0)),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	j:    `{"$l":"0"}`,
}, {
	name: "max int64",
	v:    pointer.To(Int64(math.MaxInt64)),
	b:    []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f},
	j:    `{"$l":"9223372036854775807"}`,
}, {
	name: "min int64",
	v:    pointer.To(Int64(math.MinInt64)),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
	j:    `{"$l":"-9223372036854775808"}`,
}}

func TestInt64(t *testing.T) {
	t.Parallel()

	t.Run("Binary", func(t *testing.T) {
		t.Parallel()
		testBinary(t, int64TestCases, func() bsontype { return new(Int64) })
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		testJSON(t, int64TestCases, func() bsontype { return new(Int64) })
	})
}

func FuzzInt64Binary(f *testing.F) {
	fuzzBinary(f, int64TestCases, func() bsontype { return new(Int64) })
}

func FuzzInt64JSON(f *testing.F) {
	fuzzJSON(f, int64TestCases, func() bsontype { return new(Int64) })
}

func BenchmarkInt64(b *testing.B) {
	benchmark(b, int64TestCases, func() bsontype { return new(Int64) })
}
