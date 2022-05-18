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
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func convertDocument(d *types.Document) *documentType {
	res := documentType(*d)
	return &res
}

var documentTestCases []testCase

func loadVals() {
	documentTestCases = []testCase{
		{
			name: "timestampZero",
			v: convertDocument(
				must.NotFail(types.NewDocument("timestampZero", types.Timestamp(0))),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["timestampZero"],"timestampZero":{"$t":"0"}}` +
				`}`,
		},
		{
			name: "timestamp",
			v: convertDocument(
				must.NotFail(types.NewDocument("timestamp", types.Timestamp(42))),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["timestamp"],"timestamp":{"$t":"42"}}` +
				`}`,
		},
		{
			name: "stringEmpty",
			v: convertDocument(
				must.NotFail(types.NewDocument("stringEmpty", "")),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["stringEmpty"],"stringEmpty":""}` +
				`}`,
		},
		{
			name: "stringFoo",
			v: convertDocument(
				must.NotFail(types.NewDocument("stringFoo", "foo")),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["stringFoo"],"stringFoo":"foo"}` +
				`}`,
		},
		{
			name: "objectEmpty",
			v: convertDocument(
				must.NotFail(types.NewDocument("objectEmpty", types.ObjectID{})),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["objectEmpty"],"objectEmpty":{"$o":"000000000000000000000000"}}` +
				`}`,
		},
		{
			name: "objectID42",
			v: convertDocument(
				must.NotFail(types.NewDocument("objectID42", types.ObjectID{0x42})),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["objectID42"],"objectID42":{"$o":"420000000000000000000000"}}` +
				`}`,
		},
		{
			name: "doubleZero",
			v: convertDocument(
				must.NotFail(types.NewDocument("doubleZero", float64(0.0))),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["doubleZero"],"doubleZero":0}` +
				`}`,
		},
		{
			name: "double",
			v: convertDocument(
				must.NotFail(types.NewDocument("double", float64(42.13))),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["double"],"double":42.13}` +
				`}`,
		},
		{
			name: "dateTimeZero",
			v: convertDocument(
				must.NotFail(types.NewDocument("dateTime", time.Time{})),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["dateTime"],"dateTime":"0001-01-01T00:00:00Z"}` +
				`}`,
		},
		{
			name: "datetime123",
			v: convertDocument(
				must.NotFail(types.NewDocument("datetime123", time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC).Local())),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["datetime123"],"datetime123":"2021-07-27T12:35:42.123+03:00"}` +
				`}`,
		},
		{
			name: "true",
			v: convertDocument(
				must.NotFail(types.NewDocument("bool", true)),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["bool"],"bool":true}` +
				`}`,
		},
		{
			name: "false",
			v: convertDocument(
				must.NotFail(types.NewDocument("bool", false)),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["bool"],"bool":false}` +
				`}`,
		},
		{
			name: "binaryEmpty",
			v: convertDocument(
				must.NotFail(types.NewDocument("binaryEmpty", types.Binary{Subtype: types.BinaryGeneric, B: []byte{}})),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["binaryEmpty"],"binaryEmpty":{"$b":"","s":0}}` +
				`}`,
		},
		{
			name: "binary",
			v: convertDocument(
				must.NotFail(types.NewDocument("binaryEmpty", types.Binary{Subtype: types.BinaryGeneric, B: []byte{0x42}})),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":` +
				`{"$k":["binaryEmpty"],"binaryEmpty":{"$b":"Qg==","s":0}}` +
				`}`,
		},
		{
			name: "doc",
			v: convertDocument(
				must.NotFail(types.NewDocument(
					"lsid", must.NotFail(types.NewDocument(
						"id", types.Binary{
							Subtype: types.BinaryUUID,
							B: []byte{
								0xa3, 0x19, 0xf2, 0xb4, 0xa1, 0x75, 0x40, 0xc7,
								0xb8, 0xe7, 0xa3, 0xa3, 0x2e, 0xc2, 0x56, 0xbe,
							},
						},
					)),
				)),
			),
			j: `{"$k":["type","properties"],"type":"object","properties":{` +
				`"$k":["lsid"],"lsid":{"$k":["type","properties"],"type":"object","properties":{` +
				`"$k":["id"],"id":{"$b":"oxnytKF1QMe456OjLsJWvg==","s":4}}}` +
				`}}`,
		},
	}

	for i := range documentTestCases {
		v := fromTJSON(documentTestCases[i].v).(*types.Document)
		schema := must.NotFail(DocumentSchema(v))
		documentTestCases[i].s = schema
	}
}

func TestDocument(t *testing.T) {
	t.Parallel()
	loadVals()
	testJSON(t, documentTestCases, func() tjsontype { return new(documentType) })
}

func FuzzDocument(f *testing.F) {
	fuzzJSON(f, documentTestCases, func() tjsontype { return new(documentType) })
}

func BenchmarkDocument(b *testing.B) {
	benchmark(b, documentTestCases, func() tjsontype { return new(documentType) })
}
