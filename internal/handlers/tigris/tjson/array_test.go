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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func convertArray(a *types.Array) *arrayType {
	res := arrayType(*a)
	return &res
}

var arrayTestCases = []testCase{
	{
		name: "array_all",
		v: convertArray(must.NotFail(types.NewArray(
			"foo",
			"bar",
		))),
		schema: &Schema{
			Type:  Array,
			Items: stringSchema,
		},
		j: `["foo","bar"]`,
	}, {
		name:   "EOF",
		v:      convertArray(must.NotFail(types.NewArray())),
		schema: &Schema{},
		j:      `[`,
		jErr:   `unexpected EOF`,
	}, {
		name:   "SchemaIsNil",
		v:      convertArray(must.NotFail(types.NewArray())),
		schema: new(Schema),
		j:      `["foo"]`,
		jErr:   `tjson.arrayType.UnmarshalJSON: array schema is nil for non-empty array`,
	}, {
		name:   "VariousTypesInArray",
		v:      convertArray(must.NotFail(types.NewArray())),
		schema: &Schema{Type: Array, Items: stringSchema},
		j:      `["foo",1,"bar"]`,
		jErr:   `tjson.arrayType.UnmarshalJSON: 1 elements in schema, 2 in total`,
	},
}

func TestArray(t *testing.T) {
	t.Parallel()
	testJSON(t, arrayTestCases, func() tjsontype { return new(arrayType) })
}

func FuzzArray(f *testing.F) {
	fuzzJSON(f, arrayTestCases)
}

func BenchmarkArray(b *testing.B) {
	benchmark(b, arrayTestCases)
}
