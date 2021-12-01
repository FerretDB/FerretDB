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
	"testing"

	"github.com/AlekSi/pointer"
)

var decimal128TestCases = []testCase{{
	name: "100.5",
	v:    pointer.To(Decimal128(107666)),
	b:    []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	j:    `{"$n":"107666"}`,
}, {
	name: "0",
	v:    pointer.To(Decimal128(0)),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	j:    `{"$n":"0"}`,
}}

func TestDecimal128(t *testing.T) {
	t.Parallel()

	t.Run("Binary", func(t *testing.T) {
		t.Parallel()
		testBinary(t, decimal128TestCases, func() bsontype { return new(Decimal128) })
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		testJSON(t, decimal128TestCases, func() bsontype { return new(Decimal128) })
	})
}

func FuzzDecimal128Binary(f *testing.F) {
	fuzzBinary(f, decimal128TestCases, func() bsontype { return new(Decimal128) })
}

func FuzzDecimal128JSON(f *testing.F) {
	fuzzJSON(f, decimal128TestCases, func() bsontype { return new(Decimal128) })
}

func BenchmarkDecimal128(b *testing.B) {
	benchmark(b, decimal128TestCases, func() bsontype { return new(Decimal128) })
}
