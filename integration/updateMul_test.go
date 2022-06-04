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

package integration

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestUpdateMul(t *testing.T) {
	//	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "zero_zero"}, {"value", 0}},
		bson.D{{"_id", "zero_int32"}, {"value", 0}},
		bson.D{{"_id", "zero_int64"}, {"value", 0}},
		bson.D{{"_id", "zero_float64"}, {"value", 0}},
		bson.D{{"_id", "zero_negativeInt64"}, {"value", 0}},
		bson.D{{"_id", "zero_smallestNonzeroFloat64"}, {"value", 0}},
		bson.D{{"_id", "zero_nan"}, {"value", 0}},
		bson.D{{"_id", "zero_infinity"}, {"value", 0}},
		bson.D{{"_id", "zero_positiveInfinity"}, {"value", 0}},
		bson.D{{"_id", "zero_negativeInfinity"}, {"value", 0}},
		bson.D{{"_id", "zero_negativeZero"}, {"value", 0}},
		bson.D{{"_id", "int32_zero"}, {"value", int32(1000)}},
		bson.D{{"_id", "int64_zero"}, {"value", int64(1000)}},
		bson.D{{"_id", "float64_zero"}, {"value", float64(1000.11)}},
		bson.D{{"_id", "negativeInt64_zero"}, {"value", int64(-1)}},
		bson.D{{"_id", "smallestNonzeroFloat64_zero"}, {"value", math.SmallestNonzeroFloat64}},
		bson.D{{"_id", "nan_zero"}, {"value", math.NaN()}},
		bson.D{{"_id", "infinity_zero"}, {"value", math.Inf(0)}},
		bson.D{{"_id", "positiveInfinity_zero"}, {"value", math.Inf(+1)}},
		bson.D{{"_id", "negativeInfinity_zero"}, {"value", math.Inf(-1)}},
		bson.D{{"_id", "negativeZero_zero"}, {"value", math.Copysign(0, -1)}},
		bson.D{{"_id", "int32_int32"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_int64"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_float64"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_maxInt32"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_maxInt64"}, {"value", int32(2)}},
		bson.D{{"_id", "int32_negativeZero"}, {"value", int32(2)}},
		bson.D{{"_id", "int64_int32"}, {"value", int64(10)}},
		bson.D{{"_id", "int64_int64"}, {"value", int64(10)}},
		bson.D{{"_id", "int64_float64"}, {"value", int64(10)}},
		bson.D{{"_id", "int64_maxInt32"}, {"value", int64(2)}},
		bson.D{{"_id", "int64_maxInt64"}, {"value", int64(2)}},
		bson.D{{"_id", "int64_negativeZero"}, {"value", int64(2)}},
		bson.D{{"_id", "maxInt32_maxInt32"}, {"value", math.MaxInt32}},
		bson.D{{"_id", "maxInt64_maxInt64"}, {"value", math.MaxInt64}},
		bson.D{{"_id", "maxInt64_int32"}, {"value", math.MaxInt64}},
		bson.D{{"_id", "maxInt64_int64"}, {"value", math.MaxInt64}},
		bson.D{{"_id", "float64_int32"}, {"value", float64(100.11)}},
		bson.D{{"_id", "float64_int64"}, {"value", float64(100.11)}},
		bson.D{{"_id", "float64_float64"}, {"value", float64(100.11)}},
		bson.D{{"_id", "float64_maxInt32"}, {"value", float64(100.11)}},
		bson.D{{"_id", "float64_maxInt64"}, {"value", float64(100.11)}},
		bson.D{{"_id", "float64_maxFloat64"}, {"value", float64(1.11)}},
		bson.D{{"_id", "maxFloat64_float64"}, {"value", math.MaxFloat64}},
		bson.D{{"_id", "maxFloat64_maxFloat64"}, {"value", math.MaxFloat64}},
		bson.D{{"_id", "nil_int64"}, {"value", nil}},
		bson.D{{"_id", "int64_nil"}, {"value", int64(1000)}},
		bson.D{{"_id", "nan_int64"}, {"value", nil}},
		bson.D{{"_id", "int64_nan"}, {"value", int64(1000)}},
		bson.D{{"_id", "infinity_int64"}, {"value", math.Inf(0)}},
		bson.D{{"_id", "int64_infinity"}, {"value", int64(1000)}},
		bson.D{{"_id", "negativeZero_Int64"}, {"value", math.Copysign(0, -1)}},
		bson.D{{"_id", "smallestNonzeroFloat64_smallestNonzeroFloat64"}, {"value", math.SmallestNonzeroFloat64}},
		bson.D{{"_id", "maxInt64_minInt64"}, {"value", math.MaxInt64}},
		bson.D{{"_id", "minInt64_minInt64"}, {"value", math.MinInt64}},
		bson.D{{"_id", "minInt64_negativeInt64"}, {"value", math.MinInt64}},
		bson.D{{"_id", "int64_document"}, {"value", int64(300)}},
		bson.D{{"_id", "int64_array"}, {"value", int64(300)}},
		bson.D{{"_id", "document_int64"}, {"value", primitive.D{}}},
		bson.D{{"_id", "array_int64"}, {"value", primitive.A{}}},

		//		bson.D{{"_id", "tst_1"}, {"value", bson.D{{}}}}, // connection(127.0.0.1:40833[-7]) socket was unexpectedly closed: EOF
		//		bson.D{{"_id", "tst_2"}, {"value", int64(1)}},
		//		bson.D{{"_id", "tst_3"}, {"value", "string"}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter     bson.D
		update     bson.D
		expected   bson.D
		err        *mongo.WriteError
		altMessage string
	}{
		"Zero_Zero": {
			filter:   bson.D{{"_id", "zero_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "zero_zero"}, {"value", int32(0)}},
		},
		"Zero_Int32": {
			filter:   bson.D{{"_id", "zero_int32"}},
			update:   bson.D{{"$mul", bson.D{{"value", int32(10)}}}},
			expected: bson.D{{"_id", "zero_int32"}, {"value", int32(0)}},
		},
		"Zero_Int64": {
			filter:   bson.D{{"_id", "zero_int64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(10)}}}},
			expected: bson.D{{"_id", "zero_int64"}, {"value", int64(0)}},
		},
		"Zero_Float64": {
			filter:   bson.D{{"_id", "zero_float64"}},
			update:   bson.D{{"$mul", bson.D{{"value", float64(10.11)}}}},
			expected: bson.D{{"_id", "zero_float64"}, {"value", float64(0)}},
		},
		"Zero_NegativeInt64": {
			filter:   bson.D{{"_id", "zero_negativeInt64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(-1)}}}},
			expected: bson.D{{"_id", "zero_negativeInt64"}, {"value", int64(0)}},
		},
		"Zero_SmallestNonzeroFloat64": {
			filter:   bson.D{{"_id", "zero_smallestNonzeroFloat64"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.SmallestNonzeroFloat64}}}},
			expected: bson.D{{"_id", "zero_smallestNonzeroFloat64"}, {"value", float64(0)}},
		},
		"Zero_NaN": {
			filter:   bson.D{{"_id", "zero_nan"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.NaN()}}}},
			expected: bson.D{{"_id", "zero_nan"}, {"value", math.NaN()}},
		},
		"Zero_Infinity": {
			filter:   bson.D{{"_id", "zero_infinity"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.Inf(0)}}}},
			expected: bson.D{{"_id", "zero_infinity"}, {"value", math.NaN()}},
		},
		"Zero_PositiveInfinity": {
			filter:   bson.D{{"_id", "zero_positiveInfinity"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.Inf(+1)}}}},
			expected: bson.D{{"_id", "zero_positiveInfinity"}, {"value", math.NaN()}},
		},
		"Zero_NegativeInfinity": {
			filter:   bson.D{{"_id", "zero_negativeInfinity"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.Inf(-1)}}}},
			expected: bson.D{{"_id", "zero_negativeInfinity"}, {"value", math.NaN()}},
		},
		"Zero_NegativeZero": {
			filter:   bson.D{{"_id", "zero_negativeZero"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.Copysign(0, -1)}}}},
			expected: bson.D{{"_id", "zero_negativeZero"}, {"value", math.Copysign(0, -1)}},
		},
		"NegativeInt64_Zero": {
			filter:   bson.D{{"_id", "negativeInt64_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "negativeInt64_zero"}, {"value", int64(0)}},
		},
		"SmallestNonzeroFloat64_Zero": {
			filter:   bson.D{{"_id", "smallestNonzeroFloat64_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "smallestNonzeroFloat64_zero"}, {"value", float64(0)}},
		},
		"NaN_Zero": {
			filter:   bson.D{{"_id", "nan_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "nan_zero"}, {"value", math.NaN()}},
		},
		"Infinity_Zero": {
			filter:   bson.D{{"_id", "infinity_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "infinity_zero"}, {"value", math.NaN()}},
		},
		"PositiveInfinity_Zero": {
			filter:   bson.D{{"_id", "positiveInfinity_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "positiveInfinity_zero"}, {"value", math.NaN()}},
		},
		"NegativeInfinity_Zero": {
			filter:   bson.D{{"_id", "negativeInfinity_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "negativeInfinity_zero"}, {"value", math.NaN()}},
		},
		"NegativeZero_Zero": {
			filter:   bson.D{{"_id", "negativeZero_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "negativeZero_zero"}, {"value", math.Copysign(0, -1)}},
		},
		"Int32_Zero": {
			filter:   bson.D{{"_id", "int32_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "int32_zero"}, {"value", int32(0)}},
		},
		"Int64_Zero": {
			filter:   bson.D{{"_id", "int64_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "int64_zero"}, {"value", int64(0)}},
		},
		"Float64_Zero": {
			filter:   bson.D{{"_id", "float64_zero"}},
			update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
			expected: bson.D{{"_id", "float64_zero"}, {"value", float64(0)}},
		},
		"Int32_Int32": {
			filter:   bson.D{{"_id", "int32_int32"}},
			update:   bson.D{{"$mul", bson.D{{"value", int32(10)}}}},
			expected: bson.D{{"_id", "int32_int32"}, {"value", int32(100)}},
		},
		"Int32_Int64": {
			filter:   bson.D{{"_id", "int32_int64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(10)}}}},
			expected: bson.D{{"_id", "int32_int64"}, {"value", int64(100)}},
		},
		"Int32_Float64": {
			filter:   bson.D{{"_id", "int32_float64"}},
			update:   bson.D{{"$mul", bson.D{{"value", float64(10.11)}}}},
			expected: bson.D{{"_id", "int32_float64"}, {"value", float64(101.1)}},
		},
		"Int32_MaxInt32": {
			filter:   bson.D{{"_id", "int32_maxInt32"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt32}}}},
			expected: bson.D{{"_id", "int32_maxInt32"}, {"value", int64(21474836470)}},
		},
		"Int32_MaxInt64": {
			filter: bson.D{{"_id", "int32_maxInt64"}},
			update: bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `Failed to apply $mul operations to current value ((NumberInt)2) for document {_id: "int32_maxInt64"}`,
			},
			altMessage: `Failed to apply $mul operations to current value`,
		},
		"Int32_NegativeZero": {
			filter:   bson.D{{"_id", "int32_negativeZero"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.Copysign(0, -1)}}}},
			expected: bson.D{{"_id", "int32_negativeZero"}, {"value", math.Copysign(0, -1)}},
		},
		"Int64_Int32": {
			filter:   bson.D{{"_id", "int64_int32"}},
			update:   bson.D{{"$mul", bson.D{{"value", int32(10)}}}},
			expected: bson.D{{"_id", "int64_int32"}, {"value", int64(100)}},
		},
		"Int64_Int64": {
			filter:   bson.D{{"_id", "int64_int64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(10)}}}},
			expected: bson.D{{"_id", "int64_int64"}, {"value", int64(100)}},
		},
		"Int64_Float64": {
			filter:   bson.D{{"_id", "int64_float64"}},
			update:   bson.D{{"$mul", bson.D{{"value", float64(10.11)}}}},
			expected: bson.D{{"_id", "int64_float64"}, {"value", float64(101.1)}},
		},
		"Int64_MaxInt32": {
			filter:   bson.D{{"_id", "int64_maxInt32"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt32}}}},
			expected: bson.D{{"_id", "int64_maxInt32"}, {"value", int64(4294967294)}},
		},
		"Int64_MaxInt64": {
			filter: bson.D{{"_id", "int64_maxInt64"}},
			update: bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `Failed to apply $mul operations to current value ((NumberLong)2) for document {_id: "int64_maxInt64"}`,
			},
			altMessage: `Failed to apply $mul operations to current value`,
		},
		"Int64_NegativeZero": {
			filter:   bson.D{{"_id", "int64_negativeZero"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.Copysign(0, -1)}}}},
			expected: bson.D{{"_id", "int64_negativeZero"}, {"value", math.Copysign(0, -1)}},
		},
		"MaxInt32_MaxInt32": {
			filter:   bson.D{{"_id", "maxInt32_maxInt32"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt32}}}},
			expected: bson.D{{"_id", "maxInt32_maxInt32"}, {"value", int64(4611686014132420609)}},
		},
		"MaxInt64_MaxInt64": {
			filter: bson.D{{"_id", "maxInt64_maxInt64"}},
			update: bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `Failed to apply $mul operations to current value ((NumberLong)9223372036854775807) for document {_id: "maxInt64_maxInt64"}`,
			},
			altMessage: `Failed to apply $mul operations to current value`,
		},
		"MaxInt64_Int32": {
			filter: bson.D{{"_id", "maxInt64_int32"}},
			update: bson.D{{"$mul", bson.D{{"value", int32(2)}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `Failed to apply $mul operations to current value ((NumberLong)9223372036854775807) for document {_id: "maxInt64_int32"}`,
			},
			altMessage: `Failed to apply $mul operations to current value`,
		},
		"MaxInt64_Int64": {
			filter: bson.D{{"_id", "maxInt64_int64"}},
			update: bson.D{{"$mul", bson.D{{"value", int64(2)}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `Failed to apply $mul operations to current value ((NumberLong)9223372036854775807) for document {_id: "maxInt64_int64"}`,
			},
			altMessage: `Failed to apply $mul operations to current value`,
		},
		"Float64_Int32": {
			filter:   bson.D{{"_id", "float64_int32"}},
			update:   bson.D{{"$mul", bson.D{{"value", int32(20)}}}},
			expected: bson.D{{"_id", "float64_int32"}, {"value", float64(2002.2)}},
		},
		"Float64_Int64": {
			filter:   bson.D{{"_id", "float64_int64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(20)}}}},
			expected: bson.D{{"_id", "float64_int64"}, {"value", float64(2002.2)}},
		},
		"Float64_MaxInt32": {
			filter:   bson.D{{"_id", "float64_maxInt32"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt32}}}},
			expected: bson.D{{"_id", "float64_maxInt32"}, {"value", float64(2.1498458790117e+11)}},
		},
		"Float64_MaxInt64": {
			filter:   bson.D{{"_id", "float64_maxInt64"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
			expected: bson.D{{"_id", "float64_maxInt64"}, {"value", float64(9.233517746095316e+20)}},
		},

		// "Float64_Float64": {
		// 	filter:   bson.D{{"_id", "float64_float64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", float64(20.202)}}}},
		// 	expected: bson.D{{"_id", "float64_float64"}, {"value", float64(2022.42222)}},
		// },

		"Float64_MaxFloat64": {
			filter:   bson.D{{"_id", "float64_maxFloat64"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxFloat64}}}},
			expected: bson.D{{"_id", "float64_maxFloat64"}, {"value", math.Inf(+1)}},
		},
		"MaxFloat64_Float64": {
			filter:   bson.D{{"_id", "maxFloat64_float64"}},
			update:   bson.D{{"$mul", bson.D{{"value", float64(1.11)}}}},
			expected: bson.D{{"_id", "maxFloat64_float64"}, {"value", math.Inf(+1)}},
		},
		"MaxFloat64_MaxFloat64": {
			filter:   bson.D{{"_id", "maxFloat64_maxFloat64"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.MaxFloat64}}}},
			expected: bson.D{{"_id", "maxFloat64_maxFloat64"}, {"value", math.Inf(+1)}},
		},
		"Nil_Int64": {
			filter: bson.D{{"_id", "nil_int64"}},
			update: bson.D{{"$mul", bson.D{{"value", int64(1000)}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot apply $mul to a value of non-numeric type. {_id: "nil_int64"} has the field 'value' of non-numeric type null`,
			},
			altMessage: `Cannot apply $mul to a value of non-numeric type`,
		},
		"Int64_Nil": {
			filter: bson.D{{"_id", "int64_nil"}},
			update: bson.D{{"$mul", bson.D{{"value", nil}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot multiply with non-numeric argument: {value: null}`,
			},
			altMessage: `Cannot multiply with non-numeric argument`,
		},
		"NaN_Int64": {
			filter: bson.D{{"_id", "nan_int64"}},
			update: bson.D{{"$mul", bson.D{{"value", int64(1000)}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot apply $mul to a value of non-numeric type. {_id: "nan_int64"} has the field 'value' of non-numeric type null`,
			},
			altMessage: `Cannot apply $mul to a value of non-numeric type`,
		},
		"Int64_NaN": {
			filter:   bson.D{{"_id", "int64_nan"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.NaN()}}}},
			expected: bson.D{{"_id", "int64_nan"}, {"value", math.NaN()}},
		},
		"Infinity_Int64": {
			filter:   bson.D{{"_id", "infinity_int64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(1000)}}}},
			expected: bson.D{{"_id", "infinity_int64"}, {"value", math.Inf(0)}},
		},
		"Int64_Infinity": {
			filter:   bson.D{{"_id", "int64_infinity"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.Inf(0)}}}},
			expected: bson.D{{"_id", "int64_infinity"}, {"value", math.Inf(0)}},
		},
		"NegativeZero_Int64": {
			filter:   bson.D{{"_id", "negativeZero_Int64"}},
			update:   bson.D{{"$mul", bson.D{{"value", int64(1000)}}}},
			expected: bson.D{{"_id", "negativeZero_Int64"}, {"value", math.Copysign(0, -1)}},
		},
		"SmallestNonzeroFloat64_SmallestNonzeroFloat64": {
			filter:   bson.D{{"_id", "smallestNonzeroFloat64_smallestNonzeroFloat64"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.SmallestNonzeroFloat64}}}},
			expected: bson.D{{"_id", "smallestNonzeroFloat64_smallestNonzeroFloat64"}, {"value", float64(0)}},
		},
		"MaxInt64_MinInt64": {
			filter: bson.D{{"_id", "maxInt64_minInt64"}},
			update: bson.D{{"$mul", bson.D{{"value", math.MinInt64}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `Failed to apply $mul operations to current value ((NumberLong)9223372036854775807) for document {_id: "maxInt64_minInt64"}`,
			},
			altMessage: `Failed to apply $mul operations to current value`,
		},
		"MinInt64_MinInt64": {
			filter: bson.D{{"_id", "minInt64_minInt64"}},
			update: bson.D{{"$mul", bson.D{{"value", math.MinInt64}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `Failed to apply $mul operations to current value ((NumberLong)-9223372036854775808) for document {_id: "minInt64_minInt64"}`,
			},
			altMessage: `Failed to apply $mul operations to current value`,
		},
		"MinInt64_NegativeInt64": {
			filter: bson.D{{"_id", "minInt64_negativeInt64"}},
			update: bson.D{{"$mul", bson.D{{"value", int64(-1)}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `Failed to apply $mul operations to current value ((NumberLong)-9223372036854775808) for document {_id: "minInt64_negativeInt64"}`,
			},
			altMessage: `Failed to apply $mul operations to current value`,
		},
		"Int64_Document": {
			filter: bson.D{{"_id", "int64_document"}},
			update: bson.D{{"$mul", bson.D{{"value", primitive.D{}}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot multiply with non-numeric argument: {value: {}}`,
			},
			altMessage: `Cannot multiply with non-numeric argument`,
		},
		"Int64_Array": {
			filter: bson.D{{"_id", "int64_array"}},
			update: bson.D{{"$mul", bson.D{{"value", primitive.A{}}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot multiply with non-numeric argument: {value: []}`,
			},
			altMessage: `Cannot multiply with non-numeric argument`,
		},
		"Document_Int64": {
			filter: bson.D{{"_id", "document_int64"}},
			update: bson.D{{"$mul", bson.D{{"value", int64(300)}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot apply $mul to a value of non-numeric type. {_id: "document_int64"} has the field 'value' of non-numeric type object`,
			},
			altMessage: `Cannot apply $mul to a value of non-numeric type`,
		},
		"Array_Int64": {
			filter: bson.D{{"_id", "array_int64"}},
			update: bson.D{{"$mul", bson.D{{"value", int64(300)}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: `Cannot apply $mul to a value of non-numeric type. {_id: "array_int64"} has the field 'value' of non-numeric type array`,
			},
			altMessage: `Cannot apply $mul to a value of non-numeric type`,
		},

		//////////////////////////////////////////////////////////

		// "TST1": {
		// 	filter: bson.D{{"_id", "tst_1"}},
		// 	update: bson.D{{"$mul", bson.D{{"value", int64(300)}}}},
		// 	err: &mongo.WriteError{
		// 		Code:    14,
		// 		Message: `Cannot apply $mul to a value of non-numeric type. {_id: "tst_1"} has the field 'value' of non-numeric type object`,
		// 	},
		// 	altMessage: `Cannot apply $mul to a value of non-numeric type`,
		// },
		// "TST2": {
		// 	filter: bson.D{{"_id", "tst_2"}},
		// 	update: bson.D{{"$mul", bson.D{{"value", bson.D{{}}}}}},
		// 	err: &mongo.WriteError{
		// 		Code:    14,
		// 		Message: `Cannot multiply with non-numeric argument: {value: { : null }}`,
		// 	},
		// 	altMessage: `Cannot multiply with non-numeric argument`,
		// },
		// "TST3": {
		// 	filter: bson.D{{"_id", "tst_3"}},
		// 	update: bson.D{{"$mul", bson.D{{"value", int64(1)}}}},
		// 	err: &mongo.WriteError{
		// 		Code:    14,
		// 		Message: `Cannot apply $mul to a value of non-numeric type. {_id: "tst_3"} has the field 'value' of non-numeric type string`,
		// 	},
		// 	altMessage: `Cannot apply $mul to a value of non-numeric type`,
		// },
		// "TST4": {
		// 	filter: bson.D{{"_id", "tst_2"}},
		// 	update: bson.D{{"$mul", bson.D{{"value", "string"}}}},
		// 	err: &mongo.WriteError{
		// 		Code:    14,
		// 		Message: `Cannot multiply with non-numeric argument: {value: "string"}`,
		// 	},
		// 	altMessage: `Cannot multiply with non-numeric argument`,
		// },
		// "TST5": { // TODO issues #673
		// 	filter: bson.D{{"_id", "tst_2"}},
		// 	update: bson.D{{"$mul", bson.D{{}}}},
		// 	err: &mongo.WriteError{
		// 		Code:    56,
		// 		Message: `An empty update path is not valid.`,
		// 	},
		// 	altMessage: `An empty update path is not valid.`,
		// },

		///////////

		// TODO issues #673
		/* "FieldDoc": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{"name", bson.D{{}}}}}},
			err: &mongo.WriteError{
				Code:    2,
				Message: `The 'to' field for $rename must be a string: name: { : null }`,
			},
			altMessage: `The 'to' field for $rename must be a string: name: object`,
		}, */
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			//	t.Parallel()

			_, err = collection.UpdateOne(ctx, tc.filter, tc.update)
			if tc.err != nil {
				require.NotNil(t, tc.err)
				AssertEqualAltWriteError(t, *tc.err, tc.altMessage, err)
				return
			}
			require.NoError(t, err)

			var actual bson.D
			err = collection.FindOne(ctx, tc.filter).Decode(&actual)
			require.NoError(t, err)

			AssertEqualDocuments(t, tc.expected, actual)

			// m := actual.Map()
			// k := CollectKeys(t, actual)

			// for key, item := range tc.expected {
			// 	assert.Contains(t, k, key)
			// 	assertEqual(t, item, m[key])
			// }
		})
	}
}
