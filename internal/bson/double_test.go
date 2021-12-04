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

var doubleTestCases = []testCase{{
	name: "42.13",
	v:    pointer.To(Double(42.13)),
	b:    []byte{0x71, 0x3d, 0x0a, 0xd7, 0xa3, 0x10, 0x45, 0x40},
	j:    `{"$f":42.13}`,
}, {
	name: "zero",
	v:    pointer.To(Double(0)),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	j:    `{"$f":0}`,
}, {
	name: "max float64",
	v:    pointer.To(Double(math.MaxFloat64)),
	b:    []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xef, 0x7f},
	j:    `{"$f":1.7976931348623157e+308}`,
}, {
	name: "smallest positive float64",
	v:    pointer.To(Double(math.SmallestNonzeroFloat64)),
	b:    []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	j:    `{"$f":5e-324}`,
}, {
	name: "+Infinity",
	v:    pointer.To(Double(math.Inf(1))),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0x7f},
	j:    `{"$f":"Infinity"}`,
}, {
	name: "-Infinity",
	v:    pointer.To(Double(math.Inf(-1))),
	b:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf0, 0xff},
	j:    `{"$f":"-Infinity"}`,
}, {
	name: "NaN",
	v:    pointer.To(Double(math.NaN())),
	b:    []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xf8, 0x7f},
	j:    `{"$f":"NaN"}`,
}}

func TestDouble(t *testing.T) {
	t.Parallel()

	t.Run("Binary", func(t *testing.T) {
		t.Parallel()
		testBinary(t, doubleTestCases, func() bsontype { return new(Double) })
	})

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		testJSON(t, doubleTestCases, func() bsontype { return new(Double) })
	})
}

func FuzzDoubleBinary(f *testing.F) {
	fuzzBinary(f, doubleTestCases, func() bsontype { return new(Double) })
}

func FuzzDoubleJSON(f *testing.F) {
	fuzzJSON(f, doubleTestCases, func() bsontype { return new(Double) })
}

func BenchmarkDouble(b *testing.B) {
	benchmark(b, doubleTestCases, func() bsontype { return new(Double) })
}
