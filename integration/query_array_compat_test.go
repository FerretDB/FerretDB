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

func TestQueryArrayCompatSize(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"float64": {
			filter:        bson.D{{"v", bson.D{{"$size", float64(2)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"int32": {
			filter:        bson.D{{"v", bson.D{{"$size", int32(2)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"int64": {
			filter:        bson.D{{"v", bson.D{{"$size", int64(2)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"InvalidUse": {
			filter:     bson.D{{"$size", 2}},
			resultType: emptyResult,
		},
		"NotFound": {
			filter:     bson.D{{"v", bson.D{{"$size", 4}}}},
			resultType: emptyResult,
		},
		"Zero": {
			filter:        bson.D{{"v", bson.D{{"$size", 0}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryArrayCompatDotNotation(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"PositionIndexGreaterThanArrayLength": {
			filter:     bson.D{{"v.5", bson.D{{"$type", "double"}}}},
			resultType: emptyResult,
		},
		"PositionIndexAtTheEndOfArray": {
			filter:        bson.D{{"v.1", bson.D{{"$type", "double"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"PositionTypeNull": {
			filter:        bson.D{{"v.1", bson.D{{"$type", "double"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"PositionRegex": {
			filter:        bson.D{{"v.1", primitive.Regex{Pattern: "foo"}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"NoSuchFieldPosition": {
			filter:     bson.D{{"v.some.0", bson.A{42}}},
			resultType: emptyResult,
		},
		"Field": {
			filter:        bson.D{{"v.array", int32(42)}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"FieldPosition": {
			filter:        bson.D{{"v.array.0", int32(42)}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"FieldPositionQuery": {
			filter:        bson.D{{"v.array.0", bson.D{{"$gte", int32(42)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"FieldPositionQueryNonArray": {
			filter:     bson.D{{"v.document.0", bson.D{{"$lt", int32(42)}}}},
			resultType: emptyResult,
		},
		"FieldPositionField": {
			filter:     bson.D{{"v.array.2.foo", "bar"}},
			resultType: emptyResult,
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryArrayCompatElemMatch(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"DoubleTarget": {
			filter: bson.D{
				{"_id", "double"},
				{"v", bson.D{{"$elemMatch", bson.D{{"$gt", int32(0)}}}}},
			},
			resultType: emptyResult,
		},
		"GtZero": {
			filter:        bson.D{{"v", bson.D{{"$elemMatch", bson.D{{"$gt", int32(0)}}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"GtZeroWithTypeArray": {
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
		"GtZeroWithTypeString": {
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
		"GtLt": {
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
	}

	testQueryCompat(t, testCases)
}

func TestQueryArrayCompatEquality(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"One": {
			filter:        bson.D{{"v", bson.A{int32(42)}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"Two": {
			filter:     bson.D{{"v", bson.A{42, "foo"}}},
			resultType: emptyResult,
		},
		"Three": {
			filter:        bson.D{{"v", bson.A{int32(42), "foo", nil}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"Three-reverse": {
			filter:        bson.D{{"v", bson.A{nil, "foo", int32(42)}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"Empty": {
			filter:        bson.D{{"v", bson.A{}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"Null": {
			filter:        bson.D{{"v", bson.A{nil}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryArrayCompatAll(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"String": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{"foo"}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"StringRepeated": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{"foo", "foo", "foo"}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"StringEmpty": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{""}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"Whole": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{int32(42)}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"WholeNotFound": {
			filter:     bson.D{{"v", bson.D{{"$all", bson.A{int32(44)}}}}},
			resultType: emptyResult,
		},
		"Zero": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{math.Copysign(0, +1)}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"Double": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{42.13}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"DoubleMax": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{math.MaxFloat64}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"DoubleMin": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{math.SmallestNonzeroFloat64}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"MultiAll": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{"foo", 42}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"MultiAllWithNil": {
			filter:        bson.D{{"v", bson.D{{"$all", bson.A{"foo", nil}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"Empty": {
			filter:     bson.D{{"v", bson.D{{"$all", bson.A{}}}}},
			resultType: emptyResult,
		},
		"NotFound": {
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
	}

	testQueryCompat(t, testCases)
}
