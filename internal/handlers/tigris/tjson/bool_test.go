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
	"testing"

	"github.com/AlekSi/pointer"
)

var boolTestCases = []testCase{{
	name:   "false",
	v:      pointer.To(boolType(false)),
	schema: boolSchema,
	j:      `false`,
}, {
	name:   "true",
	v:      pointer.To(boolType(true)),
	schema: boolSchema,
	j:      `true`,
}, {
	name:   "schema mismatch",
	schema: binarySchema,
	v:      pointer.To(boolType(true)),
	j:      `true`,
	sErr:   "json: cannot unmarshal bool into Go value of type map[string]json.RawMessage",
}, {
	name:   "invalid schema",
	schema: &Schema{Type: "invalid"},
	v:      pointer.To(boolType(true)),
	j:      `true`,
	sErr:   `tjson.Unmarshal: unhandled type "invalid"`,
}}

func TestBool(t *testing.T) {
	t.Parallel()
	testJSON(t, boolTestCases, func() tjsontype { return new(boolType) })
}

func FuzzBool(f *testing.F) {
	fuzzJSON(f, boolTestCases)
}

func BenchmarkBool(b *testing.B) {
	benchmark(b, boolTestCases)
}
