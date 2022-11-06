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

var regexTestCases = []testCase{{
	name:   "normal",
	v:      pointer.To(regexType{Pattern: "hoffman", Options: "i"}),
	j:      `{"$r":"hoffman","o":"i"}`,
	schema: regexSchema,
}, {
	name:   "empty",
	v:      pointer.To(regexType{Pattern: "", Options: ""}),
	j:      `{"$r":"","o":""}`,
	schema: regexSchema,
}, {
	name:   "EOF",
	j:      `{`,
	jErr:   `unexpected EOF`,
	schema: regexSchema,
}, {
	name:   "schema mismatch",
	schema: boolSchema,
	v:      pointer.To(regexType{Pattern: "hoffman", Options: "i"}),
	j:      `{"$r":"hoffman","o":"i"}`,
	sErr:   "json: cannot unmarshal object into Go value of type bool",
}, {
	name:   "invalid schema",
	schema: &Schema{Type: "invalid"},
	v:      pointer.To(regexType{Pattern: "hoffman", Options: "i"}),
	j:      `{"$r":"hoffman","o":"i"}`,
	sErr:   `tjson.Unmarshal: unhandled type "invalid"`,
}}

func TestRegex(t *testing.T) {
	t.Parallel()
	testJSON(t, regexTestCases, func() tjsontype { return new(regexType) })
}

func FuzzRegex(f *testing.F) {
	fuzzJSON(f, regexTestCases)
}

func BenchmarkRegex(b *testing.B) {
	benchmark(b, regexTestCases)
}
