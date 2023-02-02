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

package tjson

import (
	"fmt"
	"math"
	"testing"

	"github.com/AlekSi/pointer"
)

var doubleTestCases = []testCase{{
	name:   "42.13",
	v:      pointer.To(doubleType(42.13)),
	schema: doubleSchema,
	j:      `42.13`,
}, {
	name:   "zero",
	v:      pointer.To(doubleType(math.Copysign(0, +1))),
	schema: doubleSchema,
	j:      `0`, // vs 0.0
}, {
	name:   "max float64",
	v:      pointer.To(doubleType(math.MaxFloat64)),
	schema: doubleSchema,
	j:      fmt.Sprintf("%f", math.MaxFloat64), // what precision do we expect here?
	//j:      `1.7976931348623157e+308`,
}, {
	name:   "smallest positive float64",
	v:      pointer.To(doubleType(math.SmallestNonzeroFloat64)),
	schema: doubleSchema,
	j:      `5e-324`,
}, {
	name:   "EOF",
	schema: doubleSchema,
	j:      `{`,
	jErr:   `unexpected EOF`,
}, {
	name:   "schema mismatch",
	schema: boolSchema,
	v:      pointer.To(doubleType(42.13)),
	j:      `42.13`,
	sErr:   "json: cannot unmarshal number into Go value of type bool",
}, {
	name:   "invalid schema",
	schema: &Schema{Type: "invalid"},
	v:      pointer.To(doubleType(42.13)),
	j:      `42.13`,
	sErr:   `tjson.Unmarshal: unhandled type "invalid"`,
}}

func TestDouble(t *testing.T) {
	t.Parallel()
	testJSON(t, doubleTestCases, func() tjsontype { return new(doubleType) })
}

func FuzzDouble(f *testing.F) {
	fuzzJSON(f, doubleTestCases)
}

func BenchmarkDouble(b *testing.B) {
	benchmark(b, doubleTestCases)
}
