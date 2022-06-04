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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
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
		bson.D{{"_id", "zero_NaN"}, {"value", 0}},
		bson.D{{"_id", "zero_infinity"}, {"value", 0}},
		bson.D{{"_id", "zero_positiveInfinity"}, {"value", 0}},
		bson.D{{"_id", "zero_negativeInfinity"}, {"value", 0}},
		bson.D{{"_id", "zero_negativeZero"}, {"value", 0}},

		bson.D{{"_id", "int32_zero"}, {"value", int32(1000)}},
		bson.D{{"_id", "int64_zero"}, {"value", int64(1000)}},
		bson.D{{"_id", "float64_zero"}, {"value", float64(1000.11)}},

		bson.D{{"_id", "int32_int32"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_int64"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_float64"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_maxInt32"}, {"value", int32(10)}},
		bson.D{{"_id", "int32_maxInt64"}, {"value", int32(2)}},

		bson.D{{"_id", "int64_int32"}, {"value", int64(10)}},
		bson.D{{"_id", "int64_int64"}, {"value", int64(10)}},
		bson.D{{"_id", "int64_float64"}, {"value", int64(10)}},
		bson.D{{"_id", "int64_maxInt32"}, {"value", int64(2)}},
		bson.D{{"_id", "int64_maxInt64"}, {"value", int64(2)}},

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

		bson.D{{"_id", "SmallestNonzeroFloat64"}, {"value", math.SmallestNonzeroFloat64}},
		bson.D{{"_id", "NegativeNumber"}, {"value", -123456789}},

		bson.D{{"_id", "Nil"}, {"value", nil}},
		bson.D{{"_id", "NaN"}, {"value", math.NaN()}},
		bson.D{{"_id", "Infinity"}, {"value", math.Inf(0)}},
		bson.D{{"_id", "InfinityNegative"}, {"value", math.Inf(-1)}},
		bson.D{{"_id", "InfinityPositive"}, {"value", math.Inf(+1)}},
		bson.D{{"_id", "MinInt64_minus"}, {"value", float64(math.MinInt64 - 1)}},
		bson.D{{"_id", "MinInt64_overflowVerge"}, {"value", -9.223372036854776832e+18}},
		bson.D{{"_id", "NegativeZero"}, {"value", math.Copysign(0, -1)}},
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
		// "Zero_Float64": {
		// 	filter:   bson.D{{"_id", "zero_float64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", float64(10.11)}}}},
		// 	expected: map[string]any{"_id": "zero_float64", "value": float64(0)},
		// },
		// "Zero_NegativeInt64": {
		// 	filter:   bson.D{{"_id", "zero_negativeInt64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", int64(-1)}}}},
		// 	expected: map[string]any{"_id": "zero_negativeInt64", "value": int64(0)},
		// },
		// "Zero_SmallestNonzeroFloat64": {
		// 	filter:   bson.D{{"_id", "zero_smallestNonzeroFloat64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.SmallestNonzeroFloat64}}}},
		// 	expected: map[string]any{"_id": "zero_smallestNonzeroFloat64", "value": float64(0)},
		// },
		// "Zero_NaN": {
		// 	filter:   bson.D{{"_id", "zero_NaN"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.NaN()}}}},
		// 	expected: map[string]any{"_id": "zero_NaN", "value": math.NaN()},
		// },
		// "Zero_Infinity": {
		// 	filter:   bson.D{{"_id", "zero_infinity"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.Inf(0)}}}},
		// 	expected: map[string]any{"_id": "zero_infinity", "value": math.NaN()},
		// },
		// "Zero_PositiveInfinity": {
		// 	filter:   bson.D{{"_id", "zero_positiveInfinity"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.Inf(+1)}}}},
		// 	expected: map[string]any{"_id": "zero_positiveInfinity", "value": math.NaN()},
		// },
		"Zero_NegativeInfinity": {
			filter:   bson.D{{"_id", "zero_negativeInfinity"}},
			update:   bson.D{{"$mul", bson.D{{"value", math.Inf(-1)}}}},
			expected: bson.D{{"_id", "zero_negativeInfinity"}, {"value", math.NaN()}},
		},
		// "Zero_NegativeZero": {
		// 	filter:   bson.D{{"_id", "zero_negativeZero"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.Copysign(0, -1)}}}},
		// 	expected: map[string]any{"_id": "zero_negativeZero", "value": float64(0)},
		// },

		// "Int32_Zero": {
		// 	filter:   bson.D{{"_id", "int32_zero"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
		// 	expected: map[string]any{"_id": "int32_zero", "value": int32(0)},
		// },
		// "Int64_Zero": {
		// 	filter:   bson.D{{"_id", "int64_zero"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
		// 	expected: map[string]any{"_id": "int64_zero", "value": int64(0)},
		// },
		// "Float64_Zero": {
		// 	filter:   bson.D{{"_id", "float64_zero"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", 0}}}},
		// 	expected: map[string]any{"_id": "float64_zero", "value": float64(0)},
		// },
		// "Int32_Int32": {
		// 	filter:   bson.D{{"_id", "int32_int32"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", int32(10)}}}},
		// 	expected: map[string]any{"_id": "int32_int32", "value": int32(100)},
		// },
		// "Int32_Int64": {
		// 	filter:   bson.D{{"_id", "int32_int64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", int64(10)}}}},
		// 	expected: map[string]any{"_id": "int32_int64", "value": int64(100)},
		// },
		// "Int32_Float64": {
		// 	filter:   bson.D{{"_id", "int32_float64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", float64(10.11)}}}},
		// 	expected: map[string]any{"_id": "int32_float64", "value": float64(101.1)},
		// },
		// "Int32_MaxInt32": {
		// 	filter:   bson.D{{"_id", "int32_maxInt32"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt32}}}},
		// 	expected: map[string]any{"_id": "int32_maxInt32", "value": int64(21474836470)},
		// },
		// "Int32_MaxInt64": {
		// 	filter:   bson.D{{"_id", "int32_maxInt64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
		// 	expected: map[string]any{"_id": "int32_maxInt64", "value": int32(2)},
		// },
		// "Int64_Int32": {
		// 	filter:   bson.D{{"_id", "int64_int32"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", int32(10)}}}},
		// 	expected: map[string]any{"_id": "int64_int32", "value": int64(100)},
		// },
		// "Int64_Int64": {
		// 	filter:   bson.D{{"_id", "int64_int64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", int64(10)}}}},
		// 	expected: map[string]any{"_id": "int64_int64", "value": int64(100)},
		// },
		// "Int64_Float64": {
		// 	filter:   bson.D{{"_id", "int64_float64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", float64(10.11)}}}},
		// 	expected: map[string]any{"_id": "int64_float64", "value": float64(101.1)},
		// },
		// "Int64_MaxInt32": {
		// 	filter:   bson.D{{"_id", "int64_maxInt32"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt32}}}},
		// 	expected: map[string]any{"_id": "int64_maxInt32", "value": int64(4294967294)},
		// },
		// "Int64_MaxInt64": {
		// 	filter:   bson.D{{"_id", "int64_maxInt64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
		// 	expected: map[string]any{"_id": "int64_maxInt64", "value": int64(2)},
		// },
		// "MaxInt32_MaxInt32": {
		// 	filter:   bson.D{{"_id", "maxInt32_maxInt32"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt32}}}},
		// 	expected: map[string]any{"_id": "maxInt32_maxInt32", "value": int64(4611686014132420609)},
		// },
		// "MaxInt64_MaxInt64": {
		// 	filter:   bson.D{{"_id", "maxInt64_maxInt64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
		// 	expected: map[string]any{"_id": "maxInt64_maxInt64", "value": int64(math.MaxInt64)},
		// },
		// "MaxInt64_Int32": {
		// 	filter:   bson.D{{"_id", "maxInt64_int32"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", int32(2)}}}},
		// 	expected: map[string]any{"_id": "maxInt64_int32", "value": int64(math.MaxInt64)},
		// },
		// "MaxInt64_Int64": {
		// 	filter:   bson.D{{"_id", "maxInt64_int64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", int64(2)}}}},
		// 	expected: map[string]any{"_id": "maxInt64_int64", "value": int64(math.MaxInt64)},
		// },

		// "Float64_Int32": {
		// 	filter:   bson.D{{"_id", "float64_int32"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", int32(20)}}}},
		// 	expected: map[string]any{"_id": "float64_int32", "value": float64(2002.2)},
		// },
		// "Float64_Int64": {
		// 	filter:   bson.D{{"_id", "float64_int64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", int64(20)}}}},
		// 	expected: map[string]any{"_id": "float64_int64", "value": float64(2002.2)},
		// },
		// "Float64_MaxInt32": {
		// 	filter:   bson.D{{"_id", "float64_maxInt32"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt32}}}},
		// 	expected: map[string]any{"_id": "float64_maxInt32", "value": float64(2.1498458790117e+11)},
		// },
		// "Float64_MaxInt64": {
		// 	filter:   bson.D{{"_id", "float64_maxInt64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.MaxInt64}}}},
		// 	expected: map[string]any{"_id": "float64_maxInt64", "value": float64(9.233517746095316e+20)},
		// },

		// "Float64_Float64": {
		// 	filter:   bson.D{{"_id", "float64_float64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", float64(20.202)}}}},
		// 	expected: map[string]any{"_id": "float64_float64", "value": float64(2020.2)},
		// },

		// "Float64_MaxFloat64": {
		// 	filter:   bson.D{{"_id", "float64_maxFloat64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.MaxFloat64}}}},
		// 	expected: map[string]any{"_id": "float64_maxFloat64", "value": math.Inf(+1)},
		// },

		// "MaxFloat64_Float64": {
		// 	filter:   bson.D{{"_id", "maxFloat64_float64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", float64(1.11)}}}},
		// 	expected: map[string]any{"_id": "maxFloat64_float64", "value": math.Inf(+1)},
		// },

		// "MaxFloat64_MaxFloat64": {
		// 	filter:   bson.D{{"_id", "maxFloat64_maxFloat64"}},
		// 	update:   bson.D{{"$mul", bson.D{{"value", math.MaxFloat64}}}},
		// 	expected: map[string]any{"_id": "maxFloat64_maxFloat64", "value": math.Inf(+1)},
		// },

		// "FieldDoc": {
		// 	filter: bson.D{{"_id", "1"}},
		// 	update: bson.D{{"$rename", bson.D{{"name", primitive.D{}}}}},
		// 	err: &mongo.WriteError{
		// 		Code:    2,
		// 		Message: `The 'to' field for $rename must be a string: name: {}`,
		// 	},
		// 	altMessage: `The 'to' field for $rename must be a string: name: object`,
		// },

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

		// TODO issues #673
		/* "RenameDoc_1": {
			filter: bson.D{{"_id", "1"}},
			update: bson.D{{"$rename", bson.D{{}}}},
			err: &mongo.WriteError{
				Code:    56,
				Message: `An empty update path is not valid.`,
			},
			altMessage: `An empty update path is not valid.`,
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

// assertEqual is assert.Equal that also can compare NaNs and ±0.
func assertEqual(tb testing.TB, expected, actual any) bool {
	// 	tb.Helper()

	// 	switch expected := expected.(type) {
	// 	// should not be possible, check just in case
	// 	case doubleType, float64:
	// 		tb.Fatalf("unexpected type %[1]T: %[1]v", expected)

	// 	case *doubleType:
	// 		require.IsType(tb, expected, actual, msgAndArgs...)
	// 		e := float64(*expected)
	// 		a := float64(*actual.(*doubleType))
	// 		if math.IsNaN(e) || math.IsNaN(a) {
	// 			return assert.Equal(tb, math.IsNaN(e), math.IsNaN(a), msgAndArgs...)
	// 		}
	// 		if e == 0 && a == 0 {
	// 			return assert.Equal(tb, math.Signbit(e), math.Signbit(a), msgAndArgs...)
	// 		}
	// 		// fallthrough to regular assert.Equal below
	// 	}

	return assert.Equal(tb, expected, actual)
}
