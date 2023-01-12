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
	"math"
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
		name:   "array_bool",
		v:      convertArray(must.NotFail(types.NewArray(true, false, true))),
		schema: &Schema{Type: Array, Items: boolSchema},
		j:      `[true,false,true]`,
	}, {
		name:   "array_string",
		v:      convertArray(must.NotFail(types.NewArray("foo", "bar"))),
		schema: &Schema{Type: Array, Items: stringSchema},
		j:      `["foo","bar"]`,
	}, {
		name:   "array_int64",
		v:      convertArray(must.NotFail(types.NewArray(int64(math.MinInt64), int64(math.MaxInt64)))),
		schema: &Schema{Type: Array, Items: int64Schema},
		j:      `[-9223372036854775808,9223372036854775807]`,
	}, {
		name:   "array_null",
		v:      new(nullType),
		schema: &Schema{Type: Array, Items: stringSchema},
		j:      `null`,
	}, {
		name:   "array_empty",
		v:      convertArray(must.NotFail(types.NewArray())),
		schema: &Schema{Type: Array, Items: stringSchema},
		j:      `[]`,
	}, {
		name: "array_object",
		v: convertArray(must.NotFail(types.NewArray(
			must.NotFail(types.NewDocument(
				"foo", must.NotFail(types.NewArray(
					types.Regex{Pattern: "foo", Options: "i"},
					types.Regex{Pattern: "bar", Options: ""},
				)),
				"bar", "baz",
			)),

			must.NotFail(types.NewDocument(
				"foo", must.NotFail(types.NewArray(
					types.Regex{Pattern: "fizz", Options: ""},
					types.Regex{Pattern: "buzz", Options: "i"},
				)),
				"bar", "cat",
			)),
		))),
		schema: &Schema{Type: Array, Items: &Schema{
			Type: Object,
			Properties: map[string]*Schema{
				"foo": {Type: Array, Items: regexSchema},
				"bar": stringSchema,
			},
		}},
		j: `[{"$k":["foo","bar"],"foo":[{"$r":"foo","o":"i"},{"$r":"bar","o":""}],"bar":"baz"},` +
			`{"$k":["foo","bar"],"foo":[{"$r":"fizz","o":""},{"$r":"buzz","o":"i"}],"bar":"cat"}]`,
	}, {
		name:   "EOF",
		schema: &Schema{Type: Array, Items: boolSchema},
		j:      `[`,
		jErr:   `unexpected EOF`,
	}, {
		name:   "SchemaIsNil",
		v:      convertArray(must.NotFail(types.NewArray(true, false))),
		schema: &Schema{Type: Array, Items: nil},
		j:      `[true,false]`,
		jErr:   `tjson.arrayType.UnmarshalJSON: array schema is nil for non-empty array`,
	}, {
		name:   "VariousTypesInArray",
		v:      convertArray(must.NotFail(types.NewArray("foo", true, int64(42)))),
		schema: &Schema{Type: Array, Items: stringSchema},
		j:      `["foo",true,42]`, // we don't validate marshaling of array items
		jErr:   `json: cannot unmarshal bool into Go value of type string`,
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
