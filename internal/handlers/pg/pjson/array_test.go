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

package pjson

import (
	"testing"
	"time"

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
			must.NotFail(types.NewArray()),
			types.Binary{Subtype: types.BinaryUser, B: []byte{0x42}},
			true,
			time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC).Local(),
			must.NotFail(types.NewDocument()),
			42.13,
			int32(42),
			int64(42),
			"foo",
			types.Null,
		))),
		sch: &elem{
			Type: elemTypeArray,
			Items: []*elem{
				{Type: elemTypeArray, Items: []*elem{}},
				binDataSchema(types.BinaryUser),
				boolSchema,
				dateSchema,
				{Type: elemTypeObject, Schema: &schema{Properties: map[string]*elem{}, Keys: []string{}}},
				doubleSchema,
				intSchema,
				longSchema,
				stringSchema,
				nullSchema,
			},
		},
		j: `[[],"Qg==",true,1627378542123,{},42.13,42,42,"foo",null]`,
	}, {
		name: "EOF",
		sch:  &elem{Type: elemTypeArray, Items: []*elem{}},
		j:    `[`,
		jErr: `unexpected EOF`,
	}, {
		name: "SchemaIsNil",
		sch:  new(elem),
		j:    `["foo"]`,
		jErr: `pjson.arrayType.UnmarshalJSON: array schema is nil for non-empty array`,
	}, {
		name: "ExtraElemsInSchema",
		sch:  &elem{Type: elemTypeArray, Items: []*elem{stringSchema, stringSchema}},
		j:    `["foo"]`,
		jErr: `pjson.arrayType.UnmarshalJSON: 2 elements in schema, 1 in total`,
	}, {
		name: "ExtraElemsInArray",
		sch:  &elem{Type: elemTypeArray, Items: []*elem{stringSchema}},
		j:    `["foo", "bar"]`,
		jErr: `pjson.arrayType.UnmarshalJSON: 1 elements in schema, 2 in total`,
	},
}

func TestArray(t *testing.T) {
	t.Parallel()
	testJSON(t, arrayTestCases, func() pjsontype { return new(arrayType) })
}

func FuzzArray(f *testing.F) {
	fuzzJSON(f, arrayTestCases, func() pjsontype { return new(arrayType) })
}

func BenchmarkArray(b *testing.B) {
	benchmark(b, arrayTestCases, func() pjsontype { return new(arrayType) })
}
