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

func TestEvalutionMod(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "Zero"}, {"value", 0}},
		bson.D{{"_id", "NegativeZero"}, {"value", math.Copysign(0, -1)}},
		bson.D{{"_id", "Int32_1"}, {"value", int32(4080)}},
		bson.D{{"_id", "Int32_2"}, {"value", int32(1048560)}},
		bson.D{{"_id", "Int32_3"}, {"value", int32(268435440)}},
		bson.D{{"_id", "Int64_1"}, {"value", int64(1099511628000)}},
		bson.D{{"_id", "Int64_2"}, {"value", int64(281474976700000)}},
		bson.D{{"_id", "Int64_3"}, {"value", int64(72057594040000000)}},
		bson.D{{"_id", "Nil"}, {"value", nil}},
		bson.D{{"_id", "String"}, {"value", "12"}},
		bson.D{{"_id", "NaN"}, {"value", math.NaN()}},
		bson.D{{"_id", "Infinity"}, {"value", math.Inf(0)}},
		bson.D{{"_id", "InfinityNegative"}, {"value", math.Inf(-1)}},
		bson.D{{"_id", "InfinityPositive"}, {"value", math.Inf(+1)}},
		bson.D{{"_id", "SmallestNonzeroFloat64"}, {"value", math.SmallestNonzeroFloat64}},
		bson.D{{"_id", "PositiveNumber"}, {"value", 123456789}},
		bson.D{{"_id", "NegativeNumber"}, {"value", -123456789}},
		bson.D{{"_id", "MaxInt64"}, {"value", math.MaxInt64}},
		bson.D{{"_id", "MaxInt64_float"}, {"value", float64(math.MaxInt64)}},
		bson.D{{"_id", "MaxInt64_plus"}, {"value", float64(math.MaxInt64 + 1)}},
		bson.D{{"_id", "MaxInt64_overflowVerge"}, {"value", 9.223372036854776832e+18}},
		bson.D{{"_id", "MaxInt64_overflow"}, {"value", 9.223372036854776833e+18}},
		bson.D{{"_id", "MaxFloat64_minus"}, {"value", 1.79769e+307}},
		bson.D{{"_id", "MaxFloat64"}, {"value", math.MaxFloat64}},
		bson.D{{"_id", "MinInt64"}, {"value", math.MinInt64}},
		bson.D{{"_id", "MinInt64_float"}, {"value", float64(math.MinInt64)}},
		bson.D{{"_id", "MinInt64_minus"}, {"value", float64(math.MinInt64 - 1)}},
		bson.D{{"_id", "MinInt64_overflowVerge"}, {"value", -9.223372036854776832e+18}},
		bson.D{{"_id", "MinInt64_overflow"}, {"value", -9.223372036854776833e+18}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		reqData bson.D
		waitIDs []any
		err     mongo.CommandError
	}{
		"Int32": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{4000, 80}}}}},
			waitIDs: []any{"Int32_1"},
		},
		"Int32_floatDivisor": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{float64(1048500.444), 60}}}}},
			waitIDs: []any{"Int32_2"},
		},
		"Int32_floatRemainder": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{268435000, float64(440.555)}}}}},
			waitIDs: []any{"Int32_3"},
		},
		"Int32_emptyAnswer": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{268435000, float64(400)}}}}},
			waitIDs: []any{},
		},
		"Int64": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1099511620000, 8000}}}}},
			waitIDs: []any{"Int64_1"},
		},
		"Int64_floatDivisor": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{float64(281474976000000.444), 700000}}}}},
			waitIDs: []any{"Int64_2"},
		},
		"Int64_floatRemainder": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{72057594000000000, float64(40000000.555)}}}}},
			waitIDs: []any{"Int64_3"},
		},
		"Int64_emptyAnswer": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1234567890, float64(111)}}}}},
			waitIDs: []any{},
		},
		"MaxInt64_Divisor": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{math.MaxInt64, 0}}}}},
			waitIDs: []any{"Zero", "NegativeZero", "SmallestNonzeroFloat64", "MaxInt64"},
		},
		"MaxInt64_Remainder": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1, math.MaxInt64}}}}},
			waitIDs: []any{},
		},
		"MaxInt64_floatDivisor": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{float64(math.MaxInt64), 0}}}}},
			waitIDs: []any{"Zero", "NegativeZero", "SmallestNonzeroFloat64", "MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge"},
		},
		"MaxInt64_floatRemainder": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1, float64(math.MaxInt64)}}}}},
			waitIDs: []any{},
		},
		"MaxInt64_plus": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{9.223372036854775808e+18, 0}}}}},
			waitIDs: []any{"Zero", "NegativeZero", "SmallestNonzeroFloat64", "MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge"},
		},
		"MaxInt64_1": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{922337203685477580, 7}}}}},
			waitIDs: []any{"MaxInt64"},
		},
		"MaxInt64_2": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{9.223372036854775807e+17, 7}}}}},
			waitIDs: []any{},
		},
		"MaxInt64_3": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{9.223372036854775800e+17, 7}}}}},
			waitIDs: []any{},
		},
		"MaxInt64_4": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{922337203, 6854775807}}}}},
			waitIDs: []any{},
		},
		"MaxInt64_overflowVerge": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{9.223372036854776832e+18, 0}}}}},
			waitIDs: []any{"Zero", "NegativeZero", "SmallestNonzeroFloat64", "MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge"},
		},
		"MaxInt64_overflowDivisor": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{9.223372036854776833e+18, 0}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"MaxInt64_overflowBoth": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{9.223372036854776833e+18, 9.223372036854776833e+18}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"MinInt64_Divisor": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{math.MinInt64, 0}}}}},
			waitIDs: []any{"Zero", "NegativeZero", "SmallestNonzeroFloat64", "MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge"},
		},
		"MinInt64_Remainder": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1, math.MinInt64}}}}},
			waitIDs: []any{},
		},
		"MinInt64_floatDivisor": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{float64(math.MinInt64), 0}}}}},
			waitIDs: []any{"Zero", "NegativeZero", "SmallestNonzeroFloat64", "MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge"},
		},
		"MinInt64_floatRemainder": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1, float64(math.MinInt64)}}}}},
			waitIDs: []any{},
		},
		"MinInt64_minus": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-9.223372036854775809e+18, 0}}}}},
			waitIDs: []any{"Zero", "NegativeZero", "SmallestNonzeroFloat64", "MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge"},
		},
		"MinInt64_1": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-922337203685477580, -8}}}}},
			waitIDs: []any{"MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge"},
		},
		"MinInt64_2": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-9.223372036854775808e+17, -8}}}}},
			waitIDs: []any{},
		},
		"MinInt64_3": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-9.223372036854775800e+17, -8}}}}},
			waitIDs: []any{},
		},
		"MinInt64_4": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-922337203, -6854775808}}}}},
			waitIDs: []any{},
		},
		"MinInt64_overflowVerge": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-9.223372036854776832e+18, 0}}}}},
			waitIDs: []any{"Zero", "NegativeZero", "SmallestNonzeroFloat64", "MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge"},
		},
		"MinInt64_overflowDivisor": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-9.223372036854776833e+18, 0}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"MinInt64_overflowBoth": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-9.223372036854776833e+18, -9.223372036854776833e+18}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"Float64_1": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1.79769e+307, 0}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"Float64_2": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{math.MaxFloat64, 0}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"Float64_3": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{math.MaxFloat64, math.MaxFloat64}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"NegativeDivisor": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-100, 89}}}}},
			waitIDs: []any{"PositiveNumber"},
		},
		"NegativeRemainder": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{100, -89}}}}},
			waitIDs: []any{"NegativeNumber"},
		},
		"NegativeBoth": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-100, -89}}}}},
			waitIDs: []any{"NegativeNumber"},
		},
		"NegativeDivisorFloat": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-100.5, 89.5}}}}},
			waitIDs: []any{"PositiveNumber"},
		},
		"NegativeRemainderFloat": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{100.5, -89.5}}}}},
			waitIDs: []any{"NegativeNumber"},
		},
		"NegativeBothFloat": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{-100.5, -89.5}}}}},
			waitIDs: []any{"NegativeNumber"},
		},
		"DivisorZero": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{0, 1}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `divisor cannot be 0`,
			},
		},
		"ZeroNegativeDevisor": {
			reqData: bson.D{{"$mod", bson.A{math.Copysign(0, -1), 1}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `unknown top level operator: $mod. If you have a field name that starts with a '$' symbol, consider using $getField or $setField.`,
			},
		},
		"DivisorSmallestNonzeroFloat64": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{math.SmallestNonzeroFloat64, 1}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `divisor cannot be 0`,
			},
		},
		"RemainderSmallestNonzeroFloat64": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{23456789, math.SmallestNonzeroFloat64}}}}},
			waitIDs: []any{"Zero", "NegativeZero", "SmallestNonzeroFloat64"},
		},
		"EmptyArray": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, not enough elements`,
			},
		},
		"NotEnoughElements": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, not enough elements`,
			},
		},
		"TooManyElements": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1, 2, 3}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, too many elements`,
			},
		},
		"DivisorNotNumber": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{"1", 2}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor not a number`,
			},
		},
		"RemainderNotNumber": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1, "2"}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, remainder not a number`,
			},
		},
		"Nil": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{nil, 3}}}}},
			err: mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor not a number`,
			},
		},
		"DivisorNaN": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{math.NaN(), 1}}}}},
			err: mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
		"RemainderNaN": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1, math.NaN()}}}}},
			err: mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, remainder value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
		"InfinityNegative": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1, math.Inf(-1)}}}}},
			err: mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, remainder value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
		"Infinity": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{1, math.Inf(0)}}}}},
			err: mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, remainder value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
		"InfinityPositive": {
			reqData: bson.D{{"value", bson.D{{"$mod", bson.A{math.Inf(+1), 0}}}}},
			err: mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.reqData)
			if tc.err.Code != 0 {
				require.Nil(t, tc.waitIDs)
				assertEqualError(t, tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.waitIDs, collectIDs(t, actual))
		})
	}
}
