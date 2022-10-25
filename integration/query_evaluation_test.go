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
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestQueryEvaluationMod(t *testing.T) {
	setup.SkipForTigris(t)

	if runtime.GOARCH == "arm64" {
		t.Skip("TODO https://github.com/FerretDB/FerretDB/issues/491")
	}

	t.Parallel()
	ctx, collection := setup.Setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "Zero"}, {"v", 0}},
		bson.D{{"_id", "NegativeZero"}, {"v", math.Copysign(0, -1)}},
		bson.D{{"_id", "Int32_1"}, {"v", int32(4080)}},
		bson.D{{"_id", "Int32_2"}, {"v", int32(1048560)}},
		bson.D{{"_id", "Int32_3"}, {"v", int32(268435440)}},
		bson.D{{"_id", "Int64_1"}, {"v", int64(1099511628000)}},
		bson.D{{"_id", "Int64_2"}, {"v", int64(281474976700000)}},
		bson.D{{"_id", "Int64_3"}, {"v", int64(72057594040000000)}},
		bson.D{{"_id", "Nil"}, {"v", nil}},
		bson.D{{"_id", "String"}, {"v", "12"}},
		bson.D{{"_id", "NaN"}, {"v", math.NaN()}},
		bson.D{{"_id", "SmallestNonzeroFloat64"}, {"v", math.SmallestNonzeroFloat64}},
		bson.D{{"_id", "PositiveNumber"}, {"v", 123456789}},
		bson.D{{"_id", "NegativeNumber"}, {"v", -123456789}},
		bson.D{{"_id", "MaxInt64"}, {"v", math.MaxInt64}},
		bson.D{{"_id", "MaxInt64_float"}, {"v", float64(math.MaxInt64)}},
		bson.D{{"_id", "MaxInt64_plus"}, {"v", float64(math.MaxInt64 + 1)}},
		bson.D{{"_id", "MaxInt64_overflowVerge"}, {"v", 9.223372036854776832e+18}},
		bson.D{{"_id", "MaxInt64_overflow"}, {"v", 9.223372036854776833e+18}},
		bson.D{{"_id", "MaxFloat64_minus"}, {"v", 1.79769e+307}},
		bson.D{{"_id", "MaxFloat64"}, {"v", math.MaxFloat64}},
		bson.D{{"_id", "MinInt64"}, {"v", math.MinInt64}},
		bson.D{{"_id", "MinInt64_float"}, {"v", float64(math.MinInt64)}},
		bson.D{{"_id", "MinInt64_minus"}, {"v", float64(math.MinInt64 - 1)}},
		bson.D{{"_id", "MinInt64_overflowVerge"}, {"v", -9.223372036854776832e+18}},
		bson.D{{"_id", "MinInt64_overflow"}, {"v", -9.223372036854776833e+18}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      bson.D
		expectedIDs []any
		err         *mongo.CommandError
	}{
		"Int32": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{4000, 80}}}}},
			expectedIDs: []any{"Int32_1"},
		},
		"Int32_floatDivisor": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{float64(1048500.444), 60}}}}},
			expectedIDs: []any{"Int32_2"},
		},
		"Int32_floatRemainder": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{268435000, float64(440.555)}}}}},
			expectedIDs: []any{"Int32_3"},
		},
		"Int32_emptyAnswer": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{268435000, float64(400)}}}}},
			expectedIDs: []any{},
		},
		"Int64": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{1099511620000, 8000}}}}},
			expectedIDs: []any{"Int64_1"},
		},
		"Int64_floatDivisor": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{float64(281474976000000.444), 700000}}}}},
			expectedIDs: []any{"Int64_2"},
		},
		"Int64_floatRemainder": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{72057594000000000, float64(40000000.555)}}}}},
			expectedIDs: []any{"Int64_3"},
		},
		"Int64_emptyAnswer": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{1234567890, float64(111)}}}}},
			expectedIDs: []any{},
		},
		"MaxInt64_Divisor": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{math.MaxInt64, 0}}}}},
			expectedIDs: []any{"MaxInt64", "NegativeZero", "SmallestNonzeroFloat64", "Zero"},
		},
		"MaxInt64_Remainder": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{1, math.MaxInt64}}}}},
			expectedIDs: []any{},
		},
		"MaxInt64_floatDivisor": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{float64(math.MaxInt64), 0}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"MaxInt64_floatRemainder": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{1, float64(math.MaxInt64)}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, remainder value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"MaxInt64_plus": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854775808e+18, 0}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"MaxInt64_1": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{922337203685477580, 7}}}}},
			expectedIDs: []any{"MaxInt64"},
		},
		"MaxInt64_2": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854775807e+17, 7}}}}},
			expectedIDs: []any{},
		},
		"MaxInt64_3": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854775800e+17, 7}}}}},
			expectedIDs: []any{},
		},
		"MaxInt64_4": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{922337203, 6854775807}}}}},
			expectedIDs: []any{},
		},
		"MaxInt64_overflowVerge": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854776832e+18, 0}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value",
			},
		},
		"MaxInt64_overflowDivisor": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854776833e+18, 0}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"MaxInt64_overflowBoth": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854776833e+18, 9.223372036854776833e+18}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"MinInt64_Divisor": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{math.MinInt64, 0}}}}},
			expectedIDs: []any{"MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge", "NegativeZero", "SmallestNonzeroFloat64", "Zero"},
		},
		"MinInt64_Remainder": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{1, math.MinInt64}}}}},
			expectedIDs: []any{},
		},
		"MinInt64_floatDivisor": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{float64(math.MinInt64), 0}}}}},
			expectedIDs: []any{"MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge", "NegativeZero", "SmallestNonzeroFloat64", "Zero"},
		},
		"MinInt64_floatRemainder": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{1, float64(math.MinInt64)}}}}},
			expectedIDs: []any{},
		},
		"MinInt64_minus": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854775809e+18, 0}}}}},
			expectedIDs: []any{"MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge", "NegativeZero", "SmallestNonzeroFloat64", "Zero"},
		},
		"MinInt64_1": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{-922337203685477580, -8}}}}},
			expectedIDs: []any{"MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge"},
		},
		"MinInt64_2": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854775808e+17, -8}}}}},
			expectedIDs: []any{},
		},
		"MinInt64_3": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854775800e+17, -8}}}}},
			expectedIDs: []any{},
		},
		"MinInt64_4": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{-922337203, -6854775808}}}}},
			expectedIDs: []any{},
		},
		"MinInt64_overflowVerge": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854776832e+18, 0}}}}},
			expectedIDs: []any{"MinInt64", "MinInt64_float", "MinInt64_minus", "MinInt64_overflowVerge", "NegativeZero", "SmallestNonzeroFloat64", "Zero"},
		},
		"MinInt64_overflowDivisor": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854776833e+18, 0}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"MinInt64_overflowBoth": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854776833e+18, -9.223372036854776833e+18}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"Float64_1": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{1.79769e+307, 0}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"Float64_2": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{math.MaxFloat64, 0}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"Float64_3": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{math.MaxFloat64, math.MaxFloat64}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: Out of bounds coercing to integral value`,
			},
		},
		"NegativeDivisor": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{-100, 89}}}}},
			expectedIDs: []any{"PositiveNumber"},
		},
		"NegativeRemainder": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{100, -89}}}}},
			expectedIDs: []any{"NegativeNumber"},
		},
		"NegativeBoth": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{-100, -89}}}}},
			expectedIDs: []any{"NegativeNumber"},
		},
		"NegativeDivisorFloat": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{-100.5, 89.5}}}}},
			expectedIDs: []any{"PositiveNumber"},
		},
		"NegativeRemainderFloat": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{100.5, -89.5}}}}},
			expectedIDs: []any{"NegativeNumber"},
		},
		"NegativeBothFloat": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{-100.5, -89.5}}}}},
			expectedIDs: []any{"NegativeNumber"},
		},
		"DivisorZero": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{0, 1}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `divisor cannot be 0`,
			},
		},
		"ZeroNegativeDivisor": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{math.Copysign(0, -1), 1}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `divisor cannot be 0`,
			},
		},
		"DivisorSmallestNonzeroFloat64": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{math.SmallestNonzeroFloat64, 1}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `divisor cannot be 0`,
			},
		},
		"RemainderSmallestNonzeroFloat64": {
			filter:      bson.D{{"v", bson.D{{"$mod", bson.A{23456789, math.SmallestNonzeroFloat64}}}}},
			expectedIDs: []any{"NegativeZero", "SmallestNonzeroFloat64", "Zero"},
		},
		"EmptyArray": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, not enough elements`,
			},
		},
		"NotEnoughElements": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{1}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, not enough elements`,
			},
		},
		"TooManyElements": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{1, 2, 3}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, too many elements`,
			},
		},
		"DivisorNotNumber": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{"1", 2}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor not a number`,
			},
		},
		"RemainderNotNumber": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{1, "2"}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, remainder not a number`,
			},
		},
		"Nil": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{nil, 3}}}}},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: `malformed mod, divisor not a number`,
			},
		},
		"DivisorNaN": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{math.NaN(), 1}}}}},
			err: &mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, divisor value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
		"RemainderNaN": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{1, math.NaN()}}}}},
			err: &mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, remainder value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
		"InfinityNegative": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{1, math.Inf(-1)}}}}},
			err: &mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, remainder value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
		"Infinity": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{1, math.Inf(0)}}}}},
			err: &mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: `malformed mod, remainder value is invalid :: caused by :: ` +
					`Unable to coerce NaN/Inf to integral type`,
			},
		},
		"InfinityPositive": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{math.Inf(+1), 0}}}}},
			err: &mongo.CommandError{
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

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

func TestQueryEvaluationRegex(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "multiline-string"}, {"v", "bar\nfoo"}},
		bson.D{
			{"_id", "document-nested-strings"},
			{"v", bson.D{{"foo", bson.D{{"bar", "quz"}}}}},
		},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter      any
		expectedIDs []any
		err         *mongo.CommandError
		altMessage  string
	}{
		"Regex": {
			filter:      bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "foo"}}}}},
			expectedIDs: []any{"multiline-string", "string"},
		},
		"RegexNested": {
			filter:      bson.D{{"v.foo.bar", bson.D{{"$regex", primitive.Regex{Pattern: "quz"}}}}},
			expectedIDs: []any{"document-nested-strings"},
		},
		"RegexWithOption": {
			filter:      bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "42", Options: "i"}}}}},
			expectedIDs: []any{"string-double", "string-whole"},
		},
		"RegexStringOptionMatchCaseInsensitive": {
			filter:      bson.D{{"v", bson.D{{"$regex", "foo"}, {"$options", "i"}}}},
			expectedIDs: []any{"multiline-string", "regex", "string"},
		},
		"RegexStringOptionMatchLineEnd": {
			filter:      bson.D{{"v", bson.D{{"$regex", "b.*foo"}, {"$options", "s"}}}},
			expectedIDs: []any{"multiline-string"},
		},
		"RegexStringOptionMatchMultiline": {
			filter:      bson.D{{"v", bson.D{{"$regex", "^foo"}, {"$options", "m"}}}},
			expectedIDs: []any{"multiline-string", "string"},
		},
		"RegexNoSuchField": {
			filter:      bson.D{{"no-such-field", bson.D{{"$regex", primitive.Regex{Pattern: "foo"}}}}},
			expectedIDs: []any{},
		},
		"RegexNoSuchFieldString": {
			filter:      bson.D{{"no-such-field", bson.D{{"$regex", "foo"}}}},
			expectedIDs: []any{},
		},
		"RegexBadOption": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "foo", Options: "123"}}}}},
			err: &mongo.CommandError{
				Code:    51108,
				Name:    "Location51108",
				Message: ` invalid flag in regex options: 1`,
			},
			altMessage: `invalid flag in regex options: 1`,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				require.Nil(t, tc.expectedIDs)
				AssertEqualAltError(t, *tc.err, tc.altMessage, err)
				return
			}
			require.NoError(t, err)

			var actual []bson.D
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIDs, CollectIDs(t, actual))
		})
	}
}

func TestQueryEvaluationRegexErrors(t *testing.T) {
	setup.SkipForTigris(t)

	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "multiline-string"}, {"v", "bar\nfoo"}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		filter any
		err    *mongo.CommandError
	}{
		"MissingClosingParen": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "g(-z]+ng  wrong regex"}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: missing )",
			},
		},
		"MissingClosingBracket": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "g[-z+ng  wrong regex"}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: missing terminating ] for character class",
			},
		},
		"InvalidEscape": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "\\uZ"}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: PCRE does not support \\L, \\l, \\N{name}, \\U, or \\u",
			},
		},
		"NamedCapture": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "(?P<name)"}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: syntax error in subpattern name (missing terminator)",
			},
		},
		"UnexpectedParen": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: ")"}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: unmatched parentheses",
			},
		},
		"TrailingBackslash": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `abc\`}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: \\ at end of pattern",
			},
		},
		"InvalidRepetition": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `a**`}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: nothing to repeat",
			},
		},
		"MissingRepetitionArgumentStar": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `*`}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: nothing to repeat",
			},
		},
		"MissingRepetitionArgumentPlus": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `+`}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: nothing to repeat",
			},
		},
		"MissingRepetitionArgumentQuestion": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `?`}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: nothing to repeat",
			},
		},
		"InvalidClassRange": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `[z-a]`}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: range out of order in character class",
			},
		},
		"InvalidNestedRepetitionOperatorStar": {
			filter: bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `a**`}}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: nothing to repeat",
			},
		},
		"InvalidPerlOp": {
			filter: bson.D{{"v", bson.D{{"$regex", `(?z)`}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: unrecognized character after (? or (?-",
			},
		},
		"InvalidRepeatSize": {
			filter: bson.D{{"v", bson.D{{"$regex", `(aa){3,10001}`}}}},
			err: &mongo.CommandError{
				Code:    51091,
				Name:    "Location51091",
				Message: "Regular expression is invalid: regular expression is too large",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := collection.Find(ctx, tc.filter, options.Find().SetSort(bson.D{{"_id", 1}}))
			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
