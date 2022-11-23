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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestQueryArrayCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"AllString": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{"foo"}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"AllStringRepeated": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{"foo", "foo", "foo"}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"AllStringEmpty": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{""}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"AllWhole": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{int32(42)}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"AllWholeNotFound": {
			filter:     bson.D{{"v", bson.D{{"$all", bson.A{int32(44)}}}}},
			resultType: emptyResult,
		},
		"AllZero": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{math.Copysign(0, +1)}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"AllDouble": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{42.13}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"AllDoubleMax": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{math.MaxFloat64}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"AllDoubleMin": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{math.SmallestNonzeroFloat64}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"AllMultiAll": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{"foo", 42}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"AllMultiAllWithNil": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{"foo", nil}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"AllEmpty": {
			filter:     bson.D{{"v", bson.D{{"$all", bson.A{}}}}},
			resultType: emptyResult,
		},
		"AllNotFound": {
			filter:     bson.D{{"v", bson.D{{"$all", bson.A{"hello"}}}}},
			resultType: emptyResult,
		},
		"$allNeedsAnArrayInt": {
			filter:     bson.D{{"v", bson.D{{"$all", 1}}}},
			resultType: emptyResult,
		},
		"$allNeedsAnArrayNil": {
			filter:     bson.D{{"v", bson.D{{"$all", nil}}}},
			resultType: emptyResult,
		},
		"DotNotationPositionIndexGreaterThanArrayLength": {
			filter:     bson.D{{"v.5", bson.D{{"$type", "double"}}}},
			resultType: emptyResult,
		},
		"DotNotationPositionIndexAtTheEndOfArray": {
			filter:        bson.D{{"v.1", bson.D{{"$type", "double"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"DotNotationPositionTypeNull": {
			filter:        bson.D{{"v.1", bson.D{{"$type", "double"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"DotNotationPositionRegex": {
			filter:        bson.D{{"v.1", primitive.Regex{Pattern: "foo"}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"DotNotationNoSuchFieldPosition": {
			filter:     bson.D{{"v.some.0", bson.A{42}}},
			resultType: emptyResult,
		},
		"DotNotationField": {
			filter:        bson.D{{"v.array", int32(42)}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"DotNotationFieldPosition": {
			filter:        bson.D{{"v.array.0", int32(42)}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"DotNotationFieldPositionQuery": {
			filter:        bson.D{{"v.array.0", bson.D{{"$gte", int32(42)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"DotNotationFieldPositionQueryNonArray": {
			filter:     bson.D{{"v.document.0", bson.D{{"$lt", int32(42)}}}},
			resultType: emptyResult,
		},
		"DotNotationFieldPositionField": {
			filter:     bson.D{{"v.array.2.foo", "bar"}},
			resultType: emptyResult,
		},
		"ElemMatchDoubleTarget": {
			filter: bson.D{
				{"_id", "double"},
				{"v", bson.D{{"$elemMatch", bson.D{{"$gt", int32(0)}}}}},
			},
			resultType: emptyResult,
		},
		"ElemMatchGtZero": {
			filter:        bson.D{{"v", bson.D{{"$elemMatch", bson.D{{"$gt", int32(0)}}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"ElemMatchGtZeroWithTypeArray": {
			filter: bson.D{
				{"v", bson.D{
					{"$elemMatch", bson.D{
						{"$gt", int32(0)},
					}},
					{"$type", "array"},
				}},
			},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"ElemMatchGtZeroWithTypeString": {
			filter: bson.D{
				{"v", bson.D{
					{"$elemMatch", bson.D{
						{"$gt", int32(0)},
					}},
					{"$type", "string"},
				}},
			},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"ElemMatchGtLt": {
			filter: bson.D{
				{"v", bson.D{
					{"$elemMatch", bson.D{
						{"$gt", int32(0)},
						{"$lt", int32(43)},
					}},
				}},
			},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"EqualityOne": {
			filter:        bson.D{{"v", bson.A{int32(42)}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"EqualityTwo": {
			filter:     bson.D{{"v", bson.A{42, "foo"}}},
			resultType: emptyResult,
		},
		"EqualityThree": {
			filter:        bson.D{{"v", bson.A{int32(42), "foo", nil}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"EqualityThree-reverse": {
			filter:        bson.D{{"v", bson.A{nil, "foo", int32(42)}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"EqualityEmpty": {
			filter:        bson.D{{"v", bson.A{}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"EqualityNull": {
			filter:        bson.D{{"v", bson.A{nil}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"SizeFloat64": {
			filter:        bson.D{{"v", bson.D{{"$size", float64(2)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"SizeInt32": {
			filter:        bson.D{{"v", bson.D{{"$size", int32(2)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"SizeInt64": {
			filter:        bson.D{{"v", bson.D{{"$size", int64(2)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"SizeInvalidUse": {
			filter:     bson.D{{"$size", 2}},
			resultType: emptyResult,
		},
		"SizeNotFound": {
			filter:     bson.D{{"v", bson.D{{"$size", 4}}}},
			resultType: emptyResult,
		},
		"SizeZero": {
			filter:        bson.D{{"v", bson.D{{"$size", 0}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
	}

	testQueryCompat(t, testCases)
}
