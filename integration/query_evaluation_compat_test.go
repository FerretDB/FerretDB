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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestQueryEvaluationCompatRegexErrors(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"MissingClosingParen": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "g(-z]+ng  wrong regex"}}}}},
			resultType: emptyResult,
		},
		"MissingClosingBracket": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "g[-z+ng  wrong regex"}}}}},
			resultType: emptyResult,
		},
		"InvalidEscape": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "\\uZ"}}}}},
			resultType: emptyResult,
		},
		"NamedCapture": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "(?P<name)"}}}}},
			resultType: emptyResult,
		},
		"UnexpectedParen": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: ")"}}}}},
			resultType: emptyResult,
		},
		"TrailingBackslash": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `abc\`}}}}},
			resultType: emptyResult,
		},
		"InvalidRepetition": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `a**`}}}}},
			resultType: emptyResult,
		},
		"MissingRepetitionArgumentStar": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `*`}}}}},
			resultType: emptyResult,
		},
		"MissingRepetitionArgumentPlus": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `+`}}}}},
			resultType: emptyResult,
		},
		"MissingRepetitionArgumentQuestion": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `?`}}}}},
			resultType: emptyResult,
		},
		"InvalidClassRange": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `[z-a]`}}}}},
			resultType: emptyResult,
		},
		"InvalidNestedRepetitionOperatorStar": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: `a**`}}}}},
			resultType: emptyResult,
		},
		"InvalidPerlOp": {
			filter:     bson.D{{"v", bson.D{{"$regex", `(?z)`}}}},
			resultType: emptyResult,
		},
		"InvalidRepeatSize": {
			filter:     bson.D{{"v", bson.D{{"$regex", `(aa){3,10001}`}}}},
			resultType: emptyResult,
		},
		"RegexNoSuchField": {
			filter:     bson.D{{"no-such-field", bson.D{{"$regex", primitive.Regex{Pattern: "foo"}}}}},
			resultType: emptyResult,
		},
		"RegexNoSuchFieldString": {
			filter:     bson.D{{"no-such-field", bson.D{{"$regex", "foo"}}}},
			resultType: emptyResult,
		},
		"RegexBadOption": {
			filter:     bson.D{{"v", bson.D{{"$regex", primitive.Regex{Pattern: "foo", Options: "123"}}}}},
			resultType: emptyResult,
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryEvaluationCompatMod(t *testing.T) {
	if runtime.GOARCH == "arm64" {
		t.Skip("TODO https://github.com/FerretDB/FerretDB/issues/491")
	}

	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Int32": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{4000, 80}}}}},
		},
		"Int32_floatDivisor": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{float64(1048500.444), 60}}}}},
		},
		"Int32_floatRemainder": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{268435000, float64(440.555)}}}}},
		},
		"Int32_emptyAnswer": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{268435000, float64(400)}}}}},
			resultType: emptyResult,
		},
		"Int64": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{1099511620000, 8000}}}}},
		},
		"Int64_floatDivisor": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{float64(281474976000000.444), 700000}}}}},
		},
		"Int64_floatRemainder": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{72057594000000000, float64(40000000.555)}}}}},
		},
		"Int64_emptyAnswer": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1234567890, float64(111)}}}}},
			resultType: emptyResult,
		},
		"MaxInt64_Divisor": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{math.MaxInt64, 0}}}}},
		},
		"MaxInt64_Remainder": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1, math.MaxInt64}}}}},
			resultType: emptyResult,
		},
		"MaxInt64_floatDivisor": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{float64(math.MaxInt64), 0}}}}},
			resultType: emptyResult,
		},
		"MaxInt64_floatRemainder": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1, float64(math.MaxInt64)}}}}},
			resultType: emptyResult,
		},
		"MaxInt64_plus": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854775808e+18, 0}}}}},
			resultType: emptyResult,
		},
		"MaxInt64_1": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{922337203685477580, 7}}}}},
		},
		"MaxInt64_2": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854775807e+17, 7}}}}},
			resultType: emptyResult,
		},
		"MaxInt64_3": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854775800e+17, 7}}}}},
			resultType: emptyResult,
		},
		"MaxInt64_4": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{922337203, 6854775807}}}}},
			resultType: emptyResult,
		},
		"MaxInt64_overflowVerge": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854776832e+18, 0}}}}},
			resultType: emptyResult,
		},
		"MaxInt64_overflowDivisor": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854776833e+18, 0}}}}},
			resultType: emptyResult,
		},
		"MaxInt64_overflowBoth": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{9.223372036854776833e+18, 9.223372036854776833e+18}}}}},
			resultType: emptyResult,
		},
		"MinInt64_Divisor": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{math.MinInt64, 0}}}}},
		},
		"MinInt64_Remainder": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1, math.MinInt64}}}}},
			resultType: emptyResult,
		},
		"MinInt64_floatDivisor": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{float64(math.MinInt64), 0}}}}},
		},
		"MinInt64_floatRemainder": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1, float64(math.MinInt64)}}}}},
			resultType: emptyResult,
		},
		"MinInt64_minus": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854775809e+18, 0}}}}},
		},
		"MinInt64_1": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{-922337203685477580, -8}}}}},
		},
		"MinInt64_2": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854775808e+17, -8}}}}},
			resultType: emptyResult,
		},
		"MinInt64_3": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854775800e+17, -8}}}}},
			resultType: emptyResult,
		},
		"MinInt64_4": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{-922337203, -6854775808}}}}},
			resultType: emptyResult,
		},
		"MinInt64_overflowVerge": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854776832e+18, 0}}}}},
		},
		"MinInt64_overflowDivisor": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854776833e+18, 0}}}}},
			resultType: emptyResult,
		},
		"MinInt64_overflowBoth": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{-9.223372036854776833e+18, -9.223372036854776833e+18}}}}},
			resultType: emptyResult,
		},
		"Float64_1": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1.79769e+307, 0}}}}},
			resultType: emptyResult,
		},
		"Float64_2": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{math.MaxFloat64, 0}}}}},
			resultType: emptyResult,
		},
		"Float64_3": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{math.MaxFloat64, math.MaxFloat64}}}}},
			resultType: emptyResult,
		},
		"NegativeDivisor": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{-100, 89}}}}},
		},
		"NegativeRemainder": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{100, -89}}}}},
		},
		"NegativeBoth": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{-100, -89}}}}},
		},
		"NegativeDivisorFloat": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{-100.5, 89.5}}}}},
		},
		"NegativeRemainderFloat": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{100.5, -89.5}}}}},
		},
		"NegativeBothFloat": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{-100.5, -89.5}}}}},
		},
		"DivisorZero": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{0, 1}}}}},
			resultType: emptyResult,
		},
		"DivisorSmallestNonzeroFloat64": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{math.SmallestNonzeroFloat64, 1}}}}},
			resultType: emptyResult,
		},
		"RemainderSmallestNonzeroFloat64": {
			filter: bson.D{{"v", bson.D{{"$mod", bson.A{23456789, math.SmallestNonzeroFloat64}}}}},
		},
		"EmptyArray": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{}}}}},
			resultType: emptyResult,
		},
		"NotEnoughElements": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1}}}}},
			resultType: emptyResult,
		},
		"TooManyElements": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1, 2, 3}}}}},
			resultType: emptyResult,
		},
		"DivisorNotNumber": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{"1", 2}}}}},
			resultType: emptyResult,
		},
		"RemainderNotNumber": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1, "2"}}}}},
			resultType: emptyResult,
		},
		"Nil": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{nil, 3}}}}},
			resultType: emptyResult,
		},
		"InfinityNegative": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1, math.Inf(-1)}}}}},
			resultType: emptyResult,
		},
		"Infinity": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{1, math.Inf(0)}}}}},
			resultType: emptyResult,
		},
		"InfinityPositive": {
			filter:     bson.D{{"v", bson.D{{"$mod", bson.A{math.Inf(+1), 0}}}}},
			resultType: emptyResult,
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryEvaluationCompatExpr(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Expression": {
			filter: bson.D{{"$expr", "$v"}},
		},
		"ExpressionDotNotation": {
			filter: bson.D{{"$expr", "$v.foo"}},
		},
		"ExpressionIndexDotNotation": {
			filter: bson.D{{"$expr", "$v.0.foo"}},
		},
		"Document": {
			filter: bson.D{{"$expr", bson.D{{"v", "foo"}}}},
		},
		"DocumentExpression": {
			filter: bson.D{{"$expr", bson.D{{"v", "$v"}}}},
		},
		"DocumentNestedExpr": {
			filter:     bson.D{{"$expr", bson.D{{"v", bson.D{{"$expr", int32(1)}}}}}},
			resultType: emptyResult,
		},
		"DocumentInvalid": {
			filter:     bson.D{{"$expr", bson.D{{"v", "$"}}}},
			resultType: emptyResult,
		},
		"Array": {
			filter: bson.D{{"$expr", bson.A{"$v"}}},
		},
		"ArrayMany": {
			filter: bson.D{{"$expr", bson.A{nil, "foo", int32(42)}}},
		},
		"ArrayInvalid": {
			filter:     bson.D{{"$expr", bson.A{"$"}}},
			resultType: emptyResult,
		},
		"String": {
			filter: bson.D{{"$expr", "v"}},
		},
		"True": {
			filter: bson.D{{"$expr", true}},
		},
		"False": {
			filter:     bson.D{{"$expr", false}},
			resultType: emptyResult,
		},
		"IntZero": {
			filter:     bson.D{{"$expr", int32(0)}},
			resultType: emptyResult,
		},
		"Int": {
			filter: bson.D{{"$expr", int32(1)}},
		},
		"LongZero": {
			filter:     bson.D{{"$expr", int64(0)}},
			resultType: emptyResult,
		},
		"Long": {
			filter: bson.D{{"$expr", int64(42)}},
		},
		"DoubleZero": {
			filter:     bson.D{{"$expr", float64(0)}},
			resultType: emptyResult,
		},
		"Double": {
			filter: bson.D{{"$expr", float64(-1)}},
		},
		"NonExistent": {
			filter:     bson.D{{"$expr", "$non-existent"}},
			resultType: emptyResult,
		},
		"Type": {
			filter: bson.D{{"$expr", bson.D{{"$type", "$v"}}}},
		},
		"Sum": {
			filter: bson.D{{"$expr", bson.D{{"$sum", "$v"}}}},
		},
		"SumType": {
			filter: bson.D{{"$expr", bson.D{{"$type", bson.D{{"$sum", "$v"}}}}}},
		},
		"Gt": {
			filter: bson.D{{"$expr", bson.D{{"$gt", bson.A{"$v", 2}}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1456",
		},
	}

	testQueryCompat(t, testCases)
}
