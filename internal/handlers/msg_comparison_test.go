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

package handlers

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

const (
	DoubleDataType                 = "double"
	DoubleZeroDataType             = "double-zero"
	DoubleNegativeInfinityDataType = "double-negative-infinity"
	DoublePositiveInfinityDataType = "double-positive-infinity"
	ArrayDataType                  = "array"
	EmptyArrayDataType             = "array-empty"
	DoubleNanDataType              = "double-nan"
	DoubleMaxDataType              = "double-max"
	DoubleSmallestDataType         = "double-smallest"
	StringDataType                 = "string"
	EmptyStringDataType            = "string-empty"
	Int32DataType                  = "int32"
	Int32ZeroDataType              = "int32-zero"
	Int32MaxDataType               = "int32-max"
	Int32MinDataType               = "int32-min"
	Int64DataType                  = "int64"
	Int64ZeroDataType              = "int64-zero"
	Int64MaxDataType               = "int64-max"
	Int64MinDataType               = "int64-min"
	TimestampDataType              = "timestamp"
	SchemaWithAllTypes             = "values"
	CollectionWithAllTypes         = "values"
	BinaryDataType                 = "binary"
	EmptyBinaryDataType            = "binary-empty"
	BoolFalseDataType              = "bool-false"
	BoolTrueDataType               = "bool-true"
	DateTimeDataType               = "datetime"
	DateTimeEpochDataType          = "datetime-epoch"
	DateTimeMinYearDataType        = "datetime-year-min"
	DateTimeMaxYearDataType        = "datetime-year-max"
)

func TestComparison(t *testing.T) {
	dataValues := map[string]any{
		DoubleDataType:                 42.13,
		DoubleNegativeInfinityDataType: math.Inf(-1),
		DoublePositiveInfinityDataType: math.Inf(+1),
		DoubleZeroDataType:             0.0,
		DoubleNanDataType:              math.NaN(),
		DoubleMaxDataType:              math.MaxFloat64,
		DoubleSmallestDataType:         math.SmallestNonzeroFloat64,
		StringDataType:                 "foo",
		EmptyStringDataType:            "",
		Int32DataType:                  int32(42),
		Int32ZeroDataType:              int32(0),
		Int32MaxDataType:               int32(2147483647),
		Int32MinDataType:               int32(-2147483648),
		Int64DataType:                  int64(42),
		Int64ZeroDataType:              int64(0),
		Int64MaxDataType:               int64(9223372036854775807),
		Int64MinDataType:               int64(-9223372036854775808),
		DateTimeDataType:               time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC),
		DateTimeEpochDataType:          time.Unix(0, 0),
		DateTimeMinYearDataType:        time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTimeMaxYearDataType:        time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC),
		TimestampDataType:              types.Timestamp(42),
		ArrayDataType:                  must.NotFail(types.NewArray("array", int32(42))),
		EmptyArrayDataType:             must.NotFail(types.NewArray()),
		BinaryDataType:                 types.Binary{Subtype: types.BinaryUser, B: []byte{42, 0, 13}},
		EmptyBinaryDataType:            types.Binary{Subtype: 0, B: []byte{}},
		BoolFalseDataType:              false,
		BoolTrueDataType:               true,
	}

	t.Parallel()

	ctx, handler, _ := setup(t, &testutil.PoolOpts{
		ReadOnly: true,
	})

	type testCase struct {
		req  *types.Document
		resp *types.Array
	}

	testCases := map[string]testCase{
		"StringEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[StringDataType]),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x02, 0x01},
					"name", StringDataType,
					"value", dataValues[StringDataType],
				)),
			)),
		},
		"EmptyStringEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[EmptyStringDataType]),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x02, 0x00, 0x00, 0x02, 0x02},
					"name", EmptyStringDataType,
					"value", dataValues[EmptyStringDataType],
				)),
			)),
		},

		"DoubleEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x01},
					"name", DoubleDataType,
					"value", dataValues[DoubleDataType],
				)),
			)),
		},
		"DoubleNegativeInfinityEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleNegativeInfinityDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x06, 0x00, 0x00, 0x01, 0x06},
					"name", DoubleNegativeInfinityDataType,
					"value", dataValues[DoubleNegativeInfinityDataType],
				)),
			)),
		},
		"DoublePositiveInfinityEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoublePositiveInfinityDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x05, 0x00, 0x00, 0x01, 0x05},
					"name", DoublePositiveInfinityDataType,
					"value", dataValues[DoublePositiveInfinityDataType],
				)),
			)),
		},
		"DoubleZeroEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleZeroDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x01, 0x02},
					"name", DoubleZeroDataType,
					"value", dataValues[DoubleZeroDataType],
				)),
			)),
		},
		"DoubleMaxEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleMaxDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x03, 0x00, 0x00, 0x01, 0x03},
					"name", DoubleMaxDataType,
					"value", dataValues[DoubleMaxDataType],
				)),
			)),
		},
		"DoubleSmallestEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleSmallestDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x04, 0x00, 0x00, 0x01, 0x04},
					"name", DoubleSmallestDataType,
					"value", dataValues[DoubleSmallestDataType],
				)),
			)),
		},
		//"DoubleNanDataTypeEq": {
		//	// TODO: write properly NaN test
		//	// Nan eq works
		//
		//	//be, _ := fjson.Marshal(expected)
		//	//ba, _ := fjson.Marshal(actual)
		//
		//	//fmt.Println("actyeal NanKEY", string(ba))
		//	//fmt.Println("expedal NanKEY", string(be))
		//
		//	//assert.Equal(t, be, ba)
		//
		//	req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleNanDataType]),
		//
		//	resp: must.NotFail(types.NewArray(
		//		must.NotFail(types.NewDocument(
		//			"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x07, 0x00, 0x00, 0x00, 0x07},
		//			"name", DoubleNanDataType,
		//			"value", dataValues[DoubleNanDataType],
		//		)),
		//	)),
		//},

		"ArrayEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[ArrayDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x04, 0x01},
					"name", ArrayDataType,
					"value", dataValues[ArrayDataType],
				)),
			)),
		},
		"EmptyArrayEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[EmptyArrayDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x02, 0x00, 0x00, 0x04, 0x02},
					"name", EmptyArrayDataType,
					"value", dataValues[EmptyArrayDataType],
				)),
			)),
		},

		"BinaryEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[BinaryDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x05, 0x01, 0x00, 0x00, 0x05, 0x01},
					"name", BinaryDataType,
					"value", dataValues[BinaryDataType],
				)),
			)),
		},
		"EmptyBinaryEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[EmptyBinaryDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x05, 0x02, 0x00, 0x00, 0x05, 0x02},
					"name", EmptyBinaryDataType,
					"value", dataValues[EmptyBinaryDataType],
				)),
			)),
		},

		"BoolFalseEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[BoolFalseDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x08, 0x01},
					"name", BoolFalseDataType,
					"value", dataValues[BoolFalseDataType],
				)),
			)),
		},
		"BoolTrueEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[BoolTrueDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x08, 0x02, 0x00, 0x00, 0x08, 0x02},
					"name", BoolTrueDataType,
					"value", dataValues[BoolTrueDataType],
				)),
			)),
		},

		"Int32Eq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int32DataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x10, 0x01},
					"name", Int32DataType,
					"value", dataValues[Int32DataType],
				)),
			)),
		},
		"Int32ZeroEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int32ZeroDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x10, 0x02},
					"name", Int32ZeroDataType,
					"value", dataValues[Int32ZeroDataType],
				)),
			)),
		},
		"Int32MaxEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int32MaxDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x03, 0x00, 0x00, 0x10, 0x03},
					"name", Int32MaxDataType,
					"value", dataValues[Int32MaxDataType],
				)),
			)),
		},
		"Int32MinEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int32MinDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x04, 0x00, 0x00, 0x10, 0x04},
					"name", Int32MinDataType,
					"value", dataValues[Int32MinDataType],
				)),
			)),
		},

		"Int64Eq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int64DataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x01, 0x00, 0x00, 0x12, 0x01},
					"name", Int64DataType,
					"value", dataValues[Int64DataType],
				)),
			)),
		},
		"Int64ZeroEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int64ZeroDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x02, 0x00, 0x00, 0x12, 0x02},
					"name", Int64ZeroDataType,
					"value", dataValues[Int64ZeroDataType],
				)),
			)),
		},
		"Int64MaxEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int64MaxDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x03, 0x00, 0x00, 0x12, 0x03},
					"name", Int64MaxDataType,
					"value", dataValues[Int64MaxDataType],
				)),
			)),
		},
		"Int64MinEq": {
			req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int64MinDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x04, 0x00, 0x00, 0x12, 0x04},
					"name", Int64MinDataType,
					"value", dataValues[Int64MinDataType],
				)),
			)),
		},

		//"DateTimeEq": {
		//	req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DateTimeDataType]),
		//
		//	resp: must.NotFail(types.NewArray(
		//		must.NotFail(types.NewDocument(
		//			"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x9, 0x01, 0x00, 0x00, 0x00, 0x12},
		//			"name", DateTimeDataType,
		//			"value", dataValues[DateTimeDataType],
		//		)),
		//	)),
		//},
		//
		//"TimestampEq": {
		//	req: findValueByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[TimestampDataType]),
		//
		//	resp: must.NotFail(types.NewArray(
		//		must.NotFail(types.NewDocument(
		//			"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x11, 0x01, 0x00, 0x00, 0x00, 0x1d},
		//			"name", TimestampDataType,
		//			"value", dataValues[TimestampDataType],
		//		)),
		//	)),
		//},
	}

	for name, tc := range testCases { //nolint:paralleltest // false positive
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual := handle(ctx, t, handler, tc.req)
			expected := must.NotFail(types.NewDocument(
				"cursor", must.NotFail(types.NewDocument(
					"firstBatch", tc.resp,
					"id", int64(0),
					"ns", SchemaWithAllTypes+"."+CollectionWithAllTypes,
				)),
				"ok", float64(1),
			))

			assert.Equal(t, expected, actual)
		})
	}
}

func findValueByComparisonOperator(operator, collection string, value any) *types.Document {
	req := must.NotFail(types.NewDocument(
		"find", collection,
		"filter", must.NotFail(types.NewDocument(
			"value", must.NotFail(types.NewDocument(
				operator, value,
			)),
		)),
		"$db", SchemaWithAllTypes,
	))
	return req
}
