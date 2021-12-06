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
	"math/big"
	"testing"

	"github.com/AlekSi/pointer"
)

var decimal128TestCases = []testCase{{
	name: "1",
	v:    pointer.To(Decimal128(*big.NewInt(1))),
	b:    []byte{0x01},
	j:    `{"$n":"1"}`,
}, {
	name: "17",
	v:    pointer.To(Decimal128(*new(big.Int).SetBytes([]byte{0x11}))),
	b:    []byte{0x11},
	j:    `{"$n":"17"}`,
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
