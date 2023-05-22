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

package sjson

import (
	"testing"

	"github.com/AlekSi/pointer"
)

var regexTestCases = []testCase{{
	name: "normal",
	v:    pointer.To(regexType{Pattern: "hoffman", Options: "i"}),
	sch:  regexSchema("i"),
	j:    `"hoffman"`,
}, {
	name: "empty",
	v:    pointer.To(regexType{Pattern: "", Options: ""}),
	sch:  regexSchema(""),
	j:    `""`,
}, {
	name: "EOF",
	j:    `{`,
	jErr: `unexpected EOF`,
}, {
	name: "NilOptions",
	v:    pointer.To(regexType{Pattern: "hoffman", Options: ""}),
	sch: &elem{
		Type:    elemTypeRegex,
		Options: nil,
	},
	j:    `"hoffman"`,
	jErr: `regex options is nil`,
}}

func TestRegex(t *testing.T) {
	t.Parallel()
	testJSON(t, regexTestCases, func() sjsontype { return new(regexType) })
}

func FuzzRegex(f *testing.F) {
	fuzzJSON(f, regexTestCases, func() sjsontype { return new(regexType) })
}

func BenchmarkRegex(b *testing.B) {
	benchmark(b, regexTestCases, func() sjsontype { return new(regexType) })
}
