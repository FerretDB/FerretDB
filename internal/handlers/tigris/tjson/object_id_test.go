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

var objectIDTestCases = []testCase{{
	name:   "normal",
	v:      pointer.To(objectIDType{0x62, 0xea, 0x6a, 0x94, 0x3d, 0x44, 0xb1, 0x0e, 0x1b, 0x6b, 0x87, 0x97}),
	schema: objectIDSchema,
	j:      `"YupqlD1EsQ4ba4eX"`,
}, {
	name:   "EOF",
	schema: objectIDSchema,
	j:      `{`,
	jErr:   `unexpected EOF`,
}, {
	name:   "schema mismatch",
	schema: boolSchema,
	v:      pointer.To(objectIDType{0x62, 0xea, 0x6a, 0x94, 0x3d, 0x44, 0xb1, 0x0e, 0x1b, 0x6b, 0x87, 0x97}),
	j:      `"YupqlD1EsQ4ba4eX"`,
	sErr:   "json: cannot unmarshal string into Go value of type bool",
}, {
	name:   "invalid schema",
	schema: &Schema{Type: "invalid"},
	v:      pointer.To(objectIDType{0x62, 0xea, 0x6a, 0x94, 0x3d, 0x44, 0xb1, 0x0e, 0x1b, 0x6b, 0x87, 0x97}),
	j:      `"YupqlD1EsQ4ba4eX"`,
	sErr:   `tjson.Unmarshal: unhandled type "invalid"`,
}}

func TestObjectID(t *testing.T) {
	t.Parallel()
	testJSON(t, objectIDTestCases, func() tjsontype { return new(objectIDType) })
}

func FuzzObjectID(f *testing.F) {
	fuzzJSON(f, objectIDTestCases)
}

func BenchmarkObjectID(b *testing.B) {
	benchmark(b, objectIDTestCases)
}
