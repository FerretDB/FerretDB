package handlers

import (
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
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
	DataDataType                   = "data"
	BoolDataType                   = "bool"
	DatetimeDataType               = "datetime"
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
		BoolDataType:                   true,
		DataDataType:                   time.Date(2020, 2, 15, 9, 34, 33, 0, time.UTC).Local(),
		DatetimeDataType:               time.Date(2021, 7, 27, 9, 35, 42, 123000000, time.UTC).Local(),
		Int32DataType:                  int32(42),
		Int32ZeroDataType:              int32(0),
		Int32MaxDataType:               int32(2147483647),
		Int32MinDataType:               int32(-2147483648),
		Int64DataType:                  int64(42),
		Int64ZeroDataType:              int64(0),
		Int64MaxDataType:               int64(9223372036854775807),
		Int64MinDataType:               int64(-9223372036854775808),
		TimestampDataType:              types.Timestamp(42),
		ArrayDataType:                  must.NotFail(types.NewArray("array", int32(42))),
		EmptyArrayDataType:             must.NotFail(types.NewArray()),
		BinaryDataType:                 types.Binary{Subtype: types.BinaryUser, B: []byte{42, 0, 13}},
	}

	t.Parallel()

	ctx, handler, _ := setup(t, &testutil.PoolOpts{
		ReadOnly: true,
	})

	type testCase struct {
		req  *types.Document
		resp *types.Array
	}

	var testCases = map[string]testCase{

		"StringEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[StringDataType]),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x01, 0x00, 0x00, 0x00, 0x08},
					"name", StringDataType,
					"value", dataValues[StringDataType],
				)),
			)),
		},

		"EmptyStringEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[EmptyStringDataType]),
			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x02, 0x02, 0x00, 0x00, 0x00, 0x09},
					"name", EmptyStringDataType,
					"value", dataValues[EmptyStringDataType],
				)),
			)),
		},

		"DoubleEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x01},
					"name", DoubleDataType,
					"value", dataValues[DoubleDataType],
				)),
			)),
		},

		"DoubleNegativeInfinityEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleNegativeInfinityDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x06, 0x00, 0x00, 0x00, 0x06},
					"name", DoubleNegativeInfinityDataType,
					"value", dataValues[DoubleNegativeInfinityDataType],
				)),
			)),
		},

		"DoublePositiveInfinityEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoublePositiveInfinityDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x05, 0x00, 0x00, 0x00, 0x05},
					"name", DoublePositiveInfinityDataType,
					"value", dataValues[DoublePositiveInfinityDataType],
				)),
			)),
		},

		"DoubleZeroEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleZeroDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00, 0x00, 0x02},
					"name", DoubleZeroDataType,
					"value", dataValues[DoubleZeroDataType],
				)),
			)),
		},

		"DoubleMaxEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleMaxDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x03, 0x00, 0x00, 0x00, 0x03},
					"name", DoubleMaxDataType,
					"value", dataValues[DoubleMaxDataType],
				)),
			)),
		},

		"DoubleSmallestEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleSmallestDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x04, 0x00, 0x00, 0x00, 0x04},
					"name", DoubleSmallestDataType,
					"value", dataValues[DoubleSmallestDataType],
				)),
			)),
		},

		"DoubleNanDataTypeEq": {
			// TODO: write properly NaN test
			// Nan eq works

			//be, _ := fjson.Marshal(expected)
			//ba, _ := fjson.Marshal(actual)

			//fmt.Println("actyeal NanKEY", string(ba))
			//fmt.Println("expedal NanKEY", string(be))

			//assert.Equal(t, be, ba)

			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[DoubleNanDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x01, 0x07, 0x00, 0x00, 0x00, 0x07},
					"name", DoubleNanDataType,
					"value", dataValues[DoubleNanDataType],
				)),
			)),
		},

		"ArrayEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[ArrayDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x01, 0x00, 0x00, 0x00, 0x0c},
					"name", ArrayDataType,
					"value", dataValues[ArrayDataType],
				)),
			)),
		},

		"EmptyArrayEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[EmptyArrayDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x04, 0x02, 0x00, 0x00, 0x00, 0x0d},
					"name", EmptyArrayDataType,
					"value", dataValues[EmptyArrayDataType],
				)),
			)),
		},

		"BinaryEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[BinaryDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x05, 0x01, 0x00, 0x00, 0x00, 0x0e},
					"name", BinaryDataType,
					"value", dataValues[BinaryDataType],
				)),
			)),
		},

		"Int32Eq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int32DataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x01, 0x00, 0x00, 0x00, 0x19},
					"name", Int32DataType,
					"value", dataValues[Int32DataType],
				)),
			)),
		},
		"Int32ZeroEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int32ZeroDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x02, 0x00, 0x00, 0x00, 0x1a},
					"name", Int32ZeroDataType,
					"value", dataValues[Int32ZeroDataType],
				)),
			)),
		},
		"Int32MaxEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int32MaxDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x03, 0x00, 0x00, 0x00, 0x1b},
					"name", Int32MaxDataType,
					"value", dataValues[Int32MaxDataType],
				)),
			)),
		},
		"Int32MinEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int32MinDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x10, 0x04, 0x00, 0x00, 0x00, 0x1c},
					"name", Int32MinDataType,
					"value", dataValues[Int32MinDataType],
				)),
			)),
		},

		"Int64ZeroEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int64ZeroDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x03, 0x00, 0x00, 0x00, 0x20},
					"name", Int64ZeroDataType,
					"value", dataValues[Int64ZeroDataType],
				)),
			)),
		},
		"Int64MaxEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int64MaxDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x03, 0x00, 0x00, 0x00, 0x21},
					"name", Int64MaxDataType,
					"value", dataValues[Int64MaxDataType],
				)),
			)),
		},
		"Int64MinEq": {
			req: findValueRequestByComparisonOperator("$eq", CollectionWithAllTypes, dataValues[Int64MinDataType]),

			resp: must.NotFail(types.NewArray(
				must.NotFail(types.NewDocument(
					"_id", types.ObjectID{0x61, 0x2e, 0xc2, 0x80, 0x00, 0x00, 0x12, 0x04, 0x00, 0x00, 0x00, 0x22},
					"name", Int64MinDataType,
					"value", dataValues[Int64MinDataType],
				)),
			)),
		},
	}
	for name, tc := range testCases {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
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

func findValueRequestByComparisonOperator(operator, collection string, value any) *types.Document {
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
