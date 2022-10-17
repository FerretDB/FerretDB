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

package fjson

import (
	"math"
	"testing"

	"github.com/AlekSi/pointer"
)

var doubleTestCases = []testCase{{
	name: "42.13",
	v:    pointer.To(doubleType(42.13)),
	j:    `{"$f":42.13}`,
}, {
	name: "zero",
	v:    pointer.To(doubleType(math.Copysign(0, +1))),
	j:    `{"$f":0}`,
}, {
	name: "negative zero",
	v:    pointer.To(doubleType(math.Copysign(0, -1))),
	j:    `{"$f":"-0"}`,
}, {
	name: "max float64",
	v:    pointer.To(doubleType(math.MaxFloat64)),
	j:    `{"$f":1.7976931348623157e+308}`,
}, {
	name: "smallest positive float64",
	v:    pointer.To(doubleType(math.SmallestNonzeroFloat64)),
	j:    `{"$f":5e-324}`,
}, {
	name: "+Infinity",
	v:    pointer.To(doubleType(math.Inf(+1))),
	j:    `{"$f":"Infinity"}`,
}, {
	name: "-Infinity",
	v:    pointer.To(doubleType(math.Inf(-1))),
	j:    `{"$f":"-Infinity"}`,
}, {
	name: "NaN",
	v:    pointer.To(doubleType(math.NaN())),
	j:    `{"$f":"NaN"}`,
}}

func TestDouble(t *testing.T) {
	t.Parallel()
	testJSON(t, doubleTestCases, func() fjsontype { return new(doubleType) })
}
