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
	"testing"

	"github.com/AlekSi/pointer"
)

var cstringTestCases = []testCase{{
	name: "foo",
	v:    pointer.To(fjsonCString("foo")),
	j:    `{"$c":"foo"}`,
}, {
	name: "empty",
	v:    pointer.To(fjsonCString("")),
	j:    `{"$c":""}`,
}, {
	name: "EOF",
	j:    `{`,
	jErr: `unexpected EOF`,
}}

func TestCString(t *testing.T) {
	t.Parallel()
	testJSON(t, cstringTestCases, func() fjsontype { return new(fjsonCString) })
}

func FuzzCString(f *testing.F) {
	fuzzJSON(f, cstringTestCases, func() fjsontype { return new(fjsonCString) })
}

func BenchmarkCString(b *testing.B) {
	benchmark(b, cstringTestCases, func() fjsontype { return new(fjsonCString) })
}
