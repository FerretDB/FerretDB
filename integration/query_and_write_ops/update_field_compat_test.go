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

package query_and_write_ops

import (
	"math"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestUpdateFieldCompatCurrentDate(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1669")

	t.Parallel()

	testCases := map[string]updateCurrentDateCompatTestCase{
		"DocumentEmpty": {
			update:     bson.D{{"$currentDate", bson.D{}}},
			resultType: integration.EmptyResult,
		},
		"ArrayEmpty": {
			update:     bson.D{{"$currentDate", bson.A{}}},
			resultType: integration.EmptyResult,
		},
		"Int32Wrong": {
			update:     bson.D{{"$currentDate", int32(1)}},
			resultType: integration.EmptyResult,
		},
		"Nil": {
			update:     bson.D{{"$currentDate", nil}},
			resultType: integration.EmptyResult,
		},
		"BoolTrue": {
			update: bson.D{{"$currentDate", bson.D{{"v", true}}}},
			paths:  []types.Path{types.NewStaticPath("v")},
		},
		"BoolTwoTrue": {
			update: bson.D{{"$currentDate", bson.D{{"v", true}, {"nonexistent", true}}}},
			paths: []types.Path{
				types.NewStaticPath("v"),
				types.NewStaticPath("nonexistent"),
			},
		},
		"BoolFalse": {
			update: bson.D{{"$currentDate", bson.D{{"v", false}}}},
			paths:  []types.Path{types.NewStaticPath("v")},
		},
		"Int32": {
			update:     bson.D{{"$currentDate", bson.D{{"v", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"Timestamp": {
			update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "timestamp"}}}}}},
			paths:  []types.Path{types.NewStaticPath("v")},
		},
		"TimestampCapitalised": {
			update:     bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "Timestamp"}}}}}},
			resultType: integration.EmptyResult,
		},
		"Date": {
			update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "date"}}}}}},
			paths:  []types.Path{types.NewStaticPath("v")},
		},
		"WrongType": {
			update:     bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", bson.D{{"abcd", int32(1)}}}}}}}},
			resultType: integration.EmptyResult,
		},
		"NoField": {
			update: bson.D{{"$currentDate", bson.D{{"nonexistent", bson.D{{"$type", "date"}}}}}},
			paths: []types.Path{
				types.NewStaticPath("nonexistent"),
			},
		},
		"UnrecognizedOption": {
			update: bson.D{{
				"$currentDate",
				bson.D{{"v", bson.D{{"array", bson.D{{"unexsistent", bson.D{}}}}}}},
			}},
			resultType: integration.EmptyResult,
		},
		"DuplicateKeys": {
			update: bson.D{{"$currentDate", bson.D{
				{"v", bson.D{{"$type", "timestamp"}}},
				{"v", bson.D{{"$type", "timestamp"}}},
			}}},
			resultType: integration.EmptyResult,
		},
	}

	testUpdateCurrentDateCompat(t, testCases)
}

func TestUpdateFieldCompatInc(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Int32": {
			update: bson.D{{"$inc", bson.D{{"v", int32(42)}}}},
		},
		"Int32Negative": {
			update: bson.D{{"$inc", bson.D{{"v", int32(-42)}}}},
		},
		"Int64Max": {
			update: bson.D{{"$inc", bson.D{{"v", math.MaxInt64}}}},
		},
		"Int64Min": {
			update: bson.D{{"$inc", bson.D{{"v", math.MinInt64}}}},
		},
		"EmptyUpdatePath": {
			update: bson.D{{"$inc", bson.D{{}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/673",
		},
		"WrongIncTypeArray": {
			update:     bson.D{{"$inc", bson.A{}}},
			resultType: integration.EmptyResult,
		},
		"DuplicateKeys": {
			update:     bson.D{{"$inc", bson.D{{"v", int32(42)}, {"v", int32(43)}}}},
			resultType: integration.EmptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

// TestUpdateFieldCompatIncComplex are test that do not work on tigris.
func TestUpdateFieldCompatIncComplex(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1668")

	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"IntNegativeIncrement": {
			update: bson.D{{"$inc", bson.D{{"v", int32(-1)}}}},
		},
		"DoubleIncrement": {
			update: bson.D{{"$inc", bson.D{{"v", float64(42.13)}}}},
		},
		"LongNegativeIncrement": {
			update: bson.D{{"$inc", bson.D{{"v", int64(-1)}}}},
		},
		"IncTwoFields": {
			update: bson.D{{"$inc", bson.D{{"foo", int32(12)}, {"v", int32(1)}}}},
		},
		"DoubleBigDoubleIncrement": {
			update: bson.D{{"$inc", bson.D{{"v", 42.13}}}},
		},
		"DoubleIntIncrement": {
			update: bson.D{{"$inc", bson.D{{"v", int32(1)}}}},
		},
		"IntOverflow": {
			update: bson.D{{"$inc", bson.D{{"v", math.MaxInt64}}}},
		},
		"DoubleIncrementIntField": {
			update: bson.D{{"$inc", bson.D{{"v", float64(1.13)}}}},
		},
		"DoubleLongIncrement": {
			update: bson.D{{"$inc", bson.D{{"v", int64(1)}}}},
		},
		"DoubleNegativeIncrement": {
			update: bson.D{{"$inc", bson.D{{"v", float64(-42.13)}}}},
		},
		"DoubleDoubleBigIncrement": {
			update: bson.D{{"$inc", bson.D{{"v", float64(1 << 61)}}}},
		},
		"DoubleIncOnNullValue": {
			update: bson.D{{"$inc", bson.D{{"v", float64(1)}}}},
		},
		"FieldNotExist": {
			update:        bson.D{{"$inc", bson.D{{"foo", int32(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1676",
		},
		"IncOnString": {
			update:     bson.D{{"$inc", "string"}},
			resultType: integration.EmptyResult,
		},
		"IncWithStringValue": {
			update:     bson.D{{"$inc", bson.D{{"v", "bad value"}}}},
			resultType: integration.EmptyResult,
		},
		"NotExistStringValue": {
			update:     bson.D{{"$inc", bson.D{{"foo.bar", "bad value"}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationFieldExist": {
			update: bson.D{{"$inc", bson.D{{"v.foo", int32(1)}}}},
		},
		"DotNotationArrayValue": {
			update: bson.D{{"$inc", bson.D{{"v.0", int32(1)}}}},
		},
		"DotNotationFieldNotExist": {
			update: bson.D{{"$inc", bson.D{{"not.existent.path", int32(1)}}}},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$inc", bson.D{{"v..", int32(42)}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationNegativeIndex": {
			update: bson.D{{"$inc", bson.D{{"v.-1", int32(42)}}}},
		},
		"DotNotationIndexOutsideArray": {
			update: bson.D{{"$inc", bson.D{{"v.100", int32(42)}}}},
		},
		"DotNotationArrayFieldNotExist": {
			update: bson.D{{"$inc", bson.D{{"v.array.foo", int32(1)}}}},
			skip:   "TODO: fix namespace error",
		},
		"DotNotationArrayFieldExist": {
			update: bson.D{{"$inc", bson.D{{"v.array.0", int32(1)}}}},
		},
		"DotNotationArrayFieldValue": {
			update: bson.D{{"$inc", bson.D{{"v.0.foo", int32(1)}}}},
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatMax(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Int32Lower": {
			update:        bson.D{{"$max", bson.D{{"v", int32(30)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"Int32Higher": {
			update:        bson.D{{"$max", bson.D{{"v", int32(60)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"Int32Negative": {
			update:        bson.D{{"$max", bson.D{{"v", int32(-22)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"Document": {
			update: bson.D{{"$max", bson.D{{"v", bson.D{{"foo", "bar"}}}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/457",
		},
		"EmptyDocument": {
			update: bson.D{{"$max", bson.D{{"v", bson.D{{}}}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/457",
		},
		"Double": {
			update: bson.D{{"$max", bson.D{{"v", 54.32}}}},
		},
		"DoubleNegative": {
			update:        bson.D{{"$max", bson.D{{"v", -54.32}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"NotExisting": {
			update:        bson.D{{"$max", bson.D{{"v", int32(60)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},

		"MultipleQueries": {
			update:        bson.D{{"$max", bson.D{{"a", int32(30)}, {"v", int32(39)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"MultipleQueriesSorted": {
			update:        bson.D{{"$max", bson.D{{"v", int32(39)}, {"a", int32(30)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"DuplicateKeys": {
			update:     bson.D{{"$max", bson.D{{"v", int32(39)}, {"v", int32(30)}}}},
			resultType: integration.EmptyResult,
		},

		// Strings are not converted to numbers (except for Tigris with int64 fields)
		"StringIntegerHigher": {
			update:        bson.D{{"$max", bson.D{{"v", "60"}}}},
			skipForTigris: "In compat collection `v` will be a string, in Tigris - a number.",
		},
		"StringIntegerLower": {
			update:        bson.D{{"$max", bson.D{{"v", "30"}}}},
			skipForTigris: "In compat collection `v` will be a string, in Tigris - a number.",
		},
		"StringDouble": {
			update: bson.D{{"$max", bson.D{{"v", "54.32"}}}},
		},
		"StringDoubleNegative": {
			update: bson.D{{"$max", bson.D{{"v", "-54.32"}}}},
		},
		"StringLexicographicHigher": {
			update: bson.D{{"$max", bson.D{{"v", "goo"}}}},
		},
		"StringLexicographicLower": {
			update: bson.D{{"$max", bson.D{{"v", "eoo"}}}},
		},
		"StringLexicographicUpperCase": {
			update: bson.D{{"$max", bson.D{{"v", "Foo"}}}},
		},
		"BoolTrue": {
			update: bson.D{{"$max", bson.D{{"v", true}}}},
		},
		"BoolFalse": {
			update:        bson.D{{"$max", bson.D{{"v", false}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"EmptyOperand": {
			update:     bson.D{{"$max", bson.D{}}},
			resultType: integration.EmptyResult,
		},
		"DateTime": {
			update:        bson.D{{"$max", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 12, 18, 42, 123000000, time.UTC))}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"DateTimeLower": {
			update:        bson.D{{"$max", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 3, 18, 42, 123000000, time.UTC))}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayEmpty": {
			update:        bson.D{{"$max", bson.D{{"v", bson.A{}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayOne": {
			update:        bson.D{{"$max", bson.D{{"v", bson.A{int32(42)}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"Array": {
			update:        bson.D{{"$max", bson.D{{"v", bson.A{int32(42), "foo", nil}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayReverse": {
			update:        bson.D{{"$max", bson.D{{"v", bson.A{nil, "foo", int32(42)}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayNull": {
			update:        bson.D{{"$max", bson.D{{"v", bson.A{nil}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArraySlice": {
			update:        bson.D{{"$max", bson.D{{"v", bson.A{int32(42), "foo"}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayShuffledValues": {
			update:        bson.D{{"$max", bson.D{{"v", bson.A{"foo", nil, int32(42)}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayDocuments": {
			update:        bson.D{{"$max", bson.D{{"v", bson.A{bson.D{{"foo", int32(42)}}, bson.D{{"foo", nil}}}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"DotNotation": {
			update: bson.D{{"$max", bson.D{{"v.foo", int32(42)}}}},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$max", bson.D{{"v..", int32(42)}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationNegativeIndex": {
			update:        bson.D{{"$max", bson.D{{"v.-1", int32(42)}}}},
			skipForTigris: "Schema validation would fail as we will create a new field with name '-1'",
		},
		"DotNotationIndexOutsideArray": {
			update:        bson.D{{"$max", bson.D{{"v.100", int32(42)}}}},
			skipForTigris: "Will fail because of type conversion.",
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatMin(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Int32Lower": {
			update:        bson.D{{"$min", bson.D{{"v", int32(30)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"Int32Higher": {
			update:        bson.D{{"$min", bson.D{{"v", int32(60)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"Int32Negative": {
			update:        bson.D{{"$min", bson.D{{"v", int32(-22)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"Document": {
			update: bson.D{{"$min", bson.D{{"v", bson.D{{"foo", "bar"}}}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/457",
		},
		"EmptyDocument": {
			update: bson.D{{"$min", bson.D{{"v", bson.D{{}}}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/457",
		},
		"Double": {
			update: bson.D{{"$min", bson.D{{"v", 54.32}}}},
		},
		"DoubleNegative": {
			update:        bson.D{{"$min", bson.D{{"v", -54.32}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"NotExisting": {
			update:        bson.D{{"$min", bson.D{{"v", int32(60)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"MultipleQueries": {
			update:        bson.D{{"$min", bson.D{{"a", int32(30)}, {"v", int32(39)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"MultipleQueriesSorted": {
			update:        bson.D{{"$min", bson.D{{"v", int32(39)}, {"a", int32(30)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"DuplicateKeys": {
			update:     bson.D{{"$min", bson.D{{"v", int32(39)}, {"v", int32(30)}}}},
			resultType: integration.EmptyResult,
		},
		"StringIntegerHigher": {
			update:        bson.D{{"$min", bson.D{{"v", "60"}}}},
			skipForTigris: "In compat collection `v` will be a string, in Tigris - a number.",
		},
		"StringIntegerLower": {
			update:        bson.D{{"$min", bson.D{{"v", "30"}}}},
			skipForTigris: "In compat collection `v` will be a string, in Tigris - a number.",
		},
		"StringDouble": {
			update: bson.D{{"$min", bson.D{{"v", "54.32"}}}},
		},
		"StringDoubleNegative": {
			update: bson.D{{"$min", bson.D{{"v", "-54.32"}}}},
		},
		"StringLexicographicHigher": {
			update: bson.D{{"$min", bson.D{{"v", "goo"}}}},
		},
		"StringLexicographicLower": {
			update: bson.D{{"$min", bson.D{{"v", "eoo"}}}},
		},
		"StringLexicographicUpperCase": {
			update: bson.D{{"$min", bson.D{{"v", "Foo"}}}},
		},
		"BoolTrue": {
			update:        bson.D{{"$min", bson.D{{"v", true}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"BoolFalse": {
			update:        bson.D{{"$min", bson.D{{"v", false}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"EmptyOperand": {
			update:     bson.D{{"$min", bson.D{}}},
			resultType: integration.EmptyResult,
		},
		"DateTime": {
			update:        bson.D{{"$min", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 12, 18, 42, 123000000, time.UTC))}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"DateTimeLower": {
			update:        bson.D{{"$min", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 3, 18, 42, 123000000, time.UTC))}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayEmpty": {
			update:        bson.D{{"$min", bson.D{{"v", bson.A{}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayOne": {
			update:        bson.D{{"$min", bson.D{{"v", bson.A{int32(42)}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"Array": {
			update:        bson.D{{"$min", bson.D{{"v", bson.A{int32(42), "foo", nil}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayReverse": {
			update:        bson.D{{"$min", bson.D{{"v", bson.A{nil, "foo", int32(42)}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayNull": {
			update:        bson.D{{"$min", bson.D{{"v", bson.A{nil}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArraySlice": {
			update:        bson.D{{"$min", bson.D{{"v", bson.A{int32(42), "foo"}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayShuffledValues": {
			update:        bson.D{{"$min", bson.D{{"v", bson.A{"foo", nil, int32(42)}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ArrayDocuments": {
			update:        bson.D{{"$min", bson.D{{"v", bson.A{bson.D{{"foo", int32(42)}}, bson.D{{"foo", nil}}}}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"DotNotation": {
			update: bson.D{{"$min", bson.D{{"v.foo", int32(42)}}}},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$min", bson.D{{"v..", int32(42)}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationNegativeIndex": {
			update:        bson.D{{"$min", bson.D{{"v.-1", int32(42)}}}},
			skipForTigris: "Schema validation would fail as we will create a new field with name '-1'",
		},
		"DotNotationIndexOutOfArray": {
			update:        bson.D{{"$min", bson.D{{"v.100", int32(42)}}}},
			skipForTigris: "Will fail because of type conversion.",
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatRename(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Simple": {
			update: bson.D{{"$rename", bson.D{{"v", "foo"}}}},
		},
		"DuplicateField": {
			update:     bson.D{{"$rename", bson.D{{"v", "v"}}}},
			resultType: integration.EmptyResult,
		},
		"NonExistingField": {
			update:     bson.D{{"$rename", bson.D{{"foo", "bar"}}}},
			resultType: integration.EmptyResult,
		},
		"EmptyField": {
			update:     bson.D{{"$rename", bson.D{{"", "v"}}}},
			resultType: integration.EmptyResult,
		},
		"EmptyDest": {
			update:     bson.D{{"$rename", bson.D{{"v", ""}}}},
			resultType: integration.EmptyResult,
		},
		"MultipleConflictDestSource": {
			update:     bson.D{{"$rename", bson.D{{"v", "foo"}, {"foo", "bar"}}}},
			resultType: integration.EmptyResult,
		},
		"MultipleConflictSourceDest": {
			update:     bson.D{{"$rename", bson.D{{"v", "foo"}, {"bar", "v"}}}},
			resultType: integration.EmptyResult,
		},
		"MultipleConflictDestFields": {
			update:     bson.D{{"$rename", bson.D{{"v", "foo"}, {"v", "bar"}}}},
			resultType: integration.EmptyResult,
		},
		"MultipleSecondInvalid": {
			update:     bson.D{{"$rename", bson.D{{"v.foo", "boo"}, {"v.array", 1}}}},
			resultType: integration.EmptyResult,
		},
		"FieldEmpty": {
			update:     bson.D{{"$rename", bson.D{}}},
			resultType: integration.EmptyResult,
		},
		"InvalidString": {
			update:     bson.D{{"$rename", "string"}},
			resultType: integration.EmptyResult,
		},
		"InvalidDoc": {
			update:     bson.D{{"$rename", primitive.D{}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationDocumentMove": {
			update:        bson.D{{"$rename", bson.D{{"v.foo", "boo"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1776",
		},
		"DotNotationDocumentDuplicate": {
			update:        bson.D{{"$rename", bson.D{{"v.foo", "v.array"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1776",
		},
		"DotNotationDocNonExistent": {
			update:     bson.D{{"$rename", bson.D{{"not.existent.path", ""}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationArrayField": {
			update:     bson.D{{"$rename", bson.D{{"v.array.0", ""}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationArrayNonExisting": {
			update:     bson.D{{"$rename", bson.D{{"foo.0.baz", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationMultipleFields": {
			update:        bson.D{{"$rename", bson.D{{"v.foo", "v.bar"}, {"v.42", "v.43"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1776",
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$rename", bson.D{{"v..", "v.bar"}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationNegativeIndex": {
			update:     bson.D{{"$rename", bson.D{{"v.-1.bar", "v.-1.baz"}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationIndexOutOfArray": {
			update:     bson.D{{"$rename", bson.D{{"v.100.bar", "v.100.baz"}}}},
			resultType: integration.EmptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatUnset(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Simple": {
			update: bson.D{{"$unset", bson.D{{"v", ""}}}},
		},
		"NonExisting": {
			update:     bson.D{{"$unset", bson.D{{"foo", ""}}}},
			resultType: integration.EmptyResult,
		},
		"Nested": {
			update: bson.D{{"$unset", bson.D{{"v", bson.D{{"array", ""}}}}}},
		},
		"DuplicateKeys": {
			update:     bson.D{{"$unset", bson.D{{"v", ""}, {"v", ""}}}},
			resultType: integration.EmptyResult,
		},
		"Empty": {
			update:     bson.D{{"$unset", bson.D{}}},
			resultType: integration.EmptyResult,
		},
		"DocumentField": {
			update:     bson.D{{"$unset", bson.D{{"foo", ""}}}},
			resultType: integration.EmptyResult,
		},
		"EmptyArray": {
			update:     bson.D{{"$unset", bson.A{}}},
			resultType: integration.EmptyResult,
		},
		"DotNotation": {
			update: bson.D{{"$unset", bson.D{{"v.foo", ""}}}},
		},
		"DotNotationNonExistentPath": {
			update:     bson.D{{"$unset", bson.D{{"not.existent.path", ""}}}},
			resultType: integration.EmptyResult,
		},
		"DotArrayField": {
			update: bson.D{{"$unset", bson.D{{"v.array.0", ""}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1242",
		},
		"DotNotationArrNonExistentPath": {
			update:     bson.D{{"$unset", bson.D{{"non.0.existent", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$unset", bson.D{{"v..", ""}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationNegativeIndex": {
			update:     bson.D{{"$unset", bson.D{{"v.-1.bar", ""}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationIndexOutOfArray": {
			update:     bson.D{{"$unset", bson.D{{"v.100.bar", ""}}}},
			resultType: integration.EmptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatSet(t *testing.T) {
	t.Parallel()

	// Tigris does not update number type upon set due to schema.
	// Hence $set is tested on the same number type for tigris using
	// following providers.
	int32sProvider := []shareddata.Provider{shareddata.Int32s}
	int64sProvider := []shareddata.Provider{shareddata.Int64s}
	doublesProvider := []shareddata.Provider{shareddata.Doubles}

	testCases := map[string]updateCompatTestCase{
		"SetNullInExisingField": {
			update: bson.D{{"$set", bson.D{{"v", nil}}}},
		},
		"DuplicateKeys": {
			update:     bson.D{{"$set", bson.D{{"v", 42}, {"v", "hello"}}}},
			resultType: integration.EmptyResult,
		},
		"NilOperand": {
			update:     bson.D{{"$set", nil}},
			resultType: integration.EmptyResult,
		},
		"String": {
			update:     bson.D{{"$set", "string"}},
			resultType: integration.EmptyResult,
		},
		"EmptyDoc": {
			update:     bson.D{{"$set", bson.D{}}},
			resultType: integration.EmptyResult,
		},
		"OkSetString": {
			update: bson.D{{"$set", bson.D{{"v", "ok value"}}}},
		},
		"FieldNotExist": {
			update:        bson.D{{"$set", bson.D{{"foo", int32(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1676",
		},
		"Double": {
			update:        bson.D{{"$set", bson.D{{"v", float64(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1668",
		},
		"Null": {
			update: bson.D{{"$set", bson.D{{"v", nil}}}},
		},
		"Int32": {
			update:        bson.D{{"$set", bson.D{{"v", int32(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1668",
		},
		"SetTwoFields": {
			update:        bson.D{{"$set", bson.D{{"foo", int32(12)}, {"v", nil}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1676",
		},
		"Int32Type": {
			update:        bson.D{{"$set", bson.D{{"v", int32(42)}}}},
			skipForTigris: "tested in Int32TypeOnly without int64 and double shareddata",
		},
		"Int32TypeOnly": {
			update:    bson.D{{"$set", bson.D{{"v", int32(42)}}}},
			providers: int32sProvider,
		},
		"Int64Type": {
			update:        bson.D{{"$set", bson.D{{"v", int64(42)}}}},
			skipForTigris: "tested in Int64TypeOnly without int32 and double shareddata",
		},
		"Int64TypeOnly": {
			update:    bson.D{{"$set", bson.D{{"v", int64(42)}}}},
			providers: int64sProvider,
		},
		"DoubleType": {
			update:        bson.D{{"$set", bson.D{{"v", 42.0}}}},
			skipForTigris: "tested in DoubleTypeOnly without int32 and int64 shareddata",
		},
		"DoubleTypeOnly": {
			update:    bson.D{{"$set", bson.D{{"v", 42.0}}}},
			providers: doublesProvider,
		},
		"DocSameNumberType": {
			update: bson.D{{"$set", bson.D{{"v", bson.D{{"foo", int32(42)}}}}}},
		},
		"DocDifferentNumberType": {
			update:        bson.D{{"$set", bson.D{{"v", bson.D{{"foo", int64(42)}}}}}},
			skipForTigris: "Tigris cannot set different number type",
		},

		"DocumentField": {
			update:        bson.D{{"$set", bson.D{{"foo", int32(42)}, {"bar", "baz"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1676",
		},
		"Binary": {
			update: bson.D{{"$set", bson.D{{"v", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}},
		},
		"BinaryGenericSubtype": {
			update: bson.D{{"$set", bson.D{{"v", primitive.Binary{Subtype: 0x00, Data: []byte{42, 0, 13}}}}}},
		},
		"BinaryEmpty": {
			update: bson.D{{"$set", bson.D{{"v", primitive.Binary{Data: []byte{}}}}}},
		},
		"ObjectID": {
			update:        bson.D{{"$set", bson.D{{"v", must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011"))}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1830",
		},
		"ObjectIDEmpty": {
			update:        bson.D{{"$set", bson.D{{"v", primitive.NilObjectID}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1830",
		},
		"Bool": {
			update: bson.D{{"$set", bson.D{{"v", true}}}},
		},
		"Datetime": {
			update:        bson.D{{"$set", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1830",
		},
		"DatetimeNanoSecDiff": {
			update:        bson.D{{"$set", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000001, time.UTC))}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1830",
		},
		"DatetimeEpoch": {
			update:        bson.D{{"$set", bson.D{{"v", primitive.NewDateTimeFromTime(time.Unix(0, 0))}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1830",
		},
		"Regex": {
			update: bson.D{{"$set", bson.D{{"v", primitive.Regex{Pattern: "foo"}}}}},
		},
		"RegexOption": {
			update: bson.D{{"$set", bson.D{{"v", primitive.Regex{Pattern: "foo", Options: "i"}}}}},
		},
		"RegexEmpty": {
			update: bson.D{{"$set", bson.D{{"v", primitive.Regex{}}}}},
		},
		"Timestamp": {
			update: bson.D{{"$set", bson.D{{"v", primitive.Timestamp{T: 41, I: 12}}}}},
		},
		"TimestampNoI": {
			update: bson.D{{"$set", bson.D{{"v", primitive.Timestamp{T: 41}}}}},
		},
		"TimestampNoT": {
			update: bson.D{{"$set", bson.D{{"v", primitive.Timestamp{I: 12}}}}},
		},
		"DocFieldExist": {
			update: bson.D{{"$set", bson.D{{"v.foo", int32(1)}}}},
		},
		"DocumentFieldNotExist": {
			update:        bson.D{{"$set", bson.D{{"foo.bar", int32(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1676",
		},
		"ArrayFieldExist": {
			update:        bson.D{{"$set", bson.D{{"v.array.0", int32(1)}}}},
			skipForTigris: "Tigris does not support language keyword 'array' as field name",
		},
		"ArrayFieldNotExist": {
			update:        bson.D{{"$set", bson.D{{"foo.0.baz", int32(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1676",
		},
		"DocArrFieldNotExists_0": {
			update:        bson.D{{"$set", bson.D{{"v.0.foo", int32(1)}}}},
			skipForTigris: "Tigris needs a special data set: https://github.com/FerretDB/FerretDB/issues/1507",
		},
		"DocArrFieldNotExists_1": {
			update:        bson.D{{"$set", bson.D{{"v.1.foo", int32(1)}}}},
			skipForTigris: "Tigris needs a special data set: https://github.com/FerretDB/FerretDB/issues/1507",
		},
		"DocArrFieldNotExists_2": {
			update:        bson.D{{"$set", bson.D{{"v.2", int32(1)}}}},
			skipForTigris: "Tigris needs a special data set: https://github.com/FerretDB/FerretDB/issues/1507",
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$set", bson.D{{"v..", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationNegativeIndex": {
			update:        bson.D{{"$set", bson.D{{"v.-1.bar", int32(1)}}}},
			skipForTigris: "Schema validation would fail as we will create a new field with name '-1'",
		},
		"DotNotationIndexOutOfArray": {
			update:        bson.D{{"$set", bson.D{{"v.100.bar", int32(1)}}}},
			skipForTigris: "Schema validation would fail for this case. ",
		},
		"ID": {
			update:     bson.D{{"$set", bson.D{{"_id", "non-existent"}}}},
			resultType: integration.EmptyResult,
		},
		"ConflictKey": {
			update: bson.D{
				{"$set", bson.D{{"v", "val"}}},
				{"$min", bson.D{{"v.foo", "val"}}},
			},
			resultType: integration.EmptyResult,
		},
		"ConflictKeyPrefix": {
			update: bson.D{
				{"$set", bson.D{{"v.foo", "val"}}},
				{"$min", bson.D{{"v", "val"}}},
			},
			resultType: integration.EmptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatSetArray(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Many": {
			update: bson.D{{"$set", bson.D{{"foo", int32(1)}, {"bar", bson.A{}}}}},
		},
		"Array": {
			update:     bson.D{{"$set", bson.A{}}},
			resultType: integration.EmptyResult,
		},
		"ArrayNil": {
			update:        bson.D{{"$set", bson.D{{"v", bson.A{nil}}}}},
			skipForTigris: "TODO: tigris produce empty result because composites dataset is not applicable",
		},
		"EmptyArray": {
			update: bson.D{{"$set", bson.D{{"v", bson.A{}}}}},
		},
		"ArrayStringsDesc": {
			update: bson.D{{"$set", bson.D{{"v", bson.A{"c", "b", "a"}}}}},
		},
		"ArrayChangedNumberType": {
			update:        bson.D{{"$set", bson.D{{"v", bson.A{int64(42), int64(43), 45.5}}}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"ArrayUnchangedNumberType": {
			update:        bson.D{{"$set", bson.D{{"v", bson.A{int32(42), int64(43), 45.5}}}}},
			skipForTigris: "Tigris does not support mixed types in arrays",
		},
		"DocSameNumberType": {
			update:        bson.D{{"$set", bson.D{{"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}}}},
			skipForTigris: "Tigris does not support field names started from numbers (`42`)",
		},
		"DocDifferentNumberType": {
			update:        bson.D{{"$set", bson.D{{"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int64(42), "foo", nil}}}}}}},
			skipForTigris: "Tigris does not support field names started from numbers (`42`)",
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatSetOnInsert(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Nil": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v", nil}}}},
			resultType: integration.EmptyResult,
		},
		"EmptyDoc": {
			update:     bson.D{{"$setOnInsert", bson.D{}}},
			resultType: integration.EmptyResult,
		},
		"DoubleDouble": {
			update:     bson.D{{"$setOnInsert", 43.13}},
			resultType: integration.EmptyResult,
		},
		"ErrString": {
			update:     bson.D{{"$setOnInsert", "any string"}},
			resultType: integration.EmptyResult,
		},
		"ErrNil": {
			update:     bson.D{{"$setOnInsert", nil}},
			resultType: integration.EmptyResult,
		},
		"DuplicateKeys": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v", 1}, {"v", 2}}}},
			resultType: integration.EmptyResult,
		},
		"DocumentFieldExist": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v.foo", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"DocumentFieldNotExist": {
			update:     bson.D{{"$setOnInsert", bson.D{{"foo.bar", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"ArrayFieldExist": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v.array.0", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"ArrFieldNotExist": {
			update:     bson.D{{"$setOnInsert", bson.D{{"foo.0.baz", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"DocArrFieldNotExist": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v.0.foo", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v..", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationNegativeIdx": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v.-1.bar", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationIndexOutOfArr": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v.100.bar", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatSetOnInsertArray(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Array": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v", bson.A{}}}}},
			resultType: integration.EmptyResult,
		},
		"EmptyArray": {
			update:     bson.D{{"$setOnInsert", bson.A{}}},
			resultType: integration.EmptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatMixed(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"SetSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$setOnInsert", bson.D{{"v", nil}}},
			},
			resultType: integration.EmptyResult,
		},
		"SetIncSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$inc", bson.D{{"foo", int32(1)}}},
				{"$setOnInsert", bson.D{{"v", nil}}},
			},
			resultType: integration.EmptyResult,
		},
		"UnknownOperator": {
			filter:     bson.D{{"_id", "test"}},
			update:     bson.D{{"$foo", bson.D{{"foo", int32(1)}}}},
			resultType: integration.EmptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatMul(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1668")

	t.Parallel()

	providers := shareddata.AllProviders().
		// OverflowVergeDoubles and Scalars contain numbers that produces +INF on compat,
		// validation error on target upon $mul operation.
		Remove("OverflowVergeDoubles", "Scalars")

	testCases := map[string]updateCompatTestCase{
		"Int32": {
			update:    bson.D{{"$mul", bson.D{{"v", int32(42)}}}},
			providers: providers,
		},
		"Int32Negative": {
			update:    bson.D{{"$mul", bson.D{{"v", int32(-42)}}}},
			providers: providers,
		},
		"Int32Min": {
			update:    bson.D{{"$mul", bson.D{{"v", math.MinInt32}}}},
			providers: providers,
		},
		"Int32Max": {
			update:    bson.D{{"$mul", bson.D{{"v", math.MaxInt32}}}},
			providers: providers,
		},
		"Int64": {
			update:    bson.D{{"$mul", bson.D{{"v", int64(42)}}}},
			providers: providers,
		},
		"Int64Negative": {
			update:    bson.D{{"$mul", bson.D{{"v", int32(-42)}}}},
			providers: providers,
		},
		"Int64Min": {
			update:    bson.D{{"$mul", bson.D{{"v", math.MinInt64}}}},
			providers: providers,
		},
		"Int64Max": {
			update:    bson.D{{"$mul", bson.D{{"v", math.MaxInt64}}}},
			providers: providers,
		},
		"Double": {
			update:    bson.D{{"$mul", bson.D{{"v", 42.13}}}},
			providers: providers,
		},
		"DoubleNegative": {
			update:    bson.D{{"$mul", bson.D{{"v", int32(-42)}}}},
			providers: providers,
		},
		"DoubleSmallestNonZero": {
			update:    bson.D{{"$mul", bson.D{{"v", math.SmallestNonzeroFloat64}}}},
			providers: providers,
		},
		"DoubleBig": {
			update:    bson.D{{"$mul", bson.D{{"v", float64(1 << 61)}}}},
			providers: providers,
		},
		"Empty": {
			update:     bson.D{{"$mul", bson.D{}}},
			resultType: integration.EmptyResult,
		},
		"Null": {
			update:     bson.D{{"$mul", bson.D{{"v", nil}}}},
			resultType: integration.EmptyResult,
		},
		"String": {
			update:     bson.D{{"$mul", bson.D{{"v", "string"}}}},
			resultType: integration.EmptyResult,
		},
		"MissingField": {
			update:     bson.D{{"$mul", "invalid"}},
			resultType: integration.EmptyResult,
		},
		"StringFieldNotExist": {
			update:     bson.D{{"$mul", bson.D{{"foo.bar", "bad value"}}}},
			resultType: integration.EmptyResult,
		},
		"FieldNotExist": {
			update: bson.D{{"$mul", bson.D{{"foo", int32(45)}}}},
		},
		"TwoFields": {
			update: bson.D{{"$mul", bson.D{{"foo", int32(12)}, {"v", int32(1)}}}},
		},
		"DuplicateKeys": {
			update:     bson.D{{"$mul", bson.D{{"v", int32(42)}, {"v", int32(43)}}}},
			resultType: integration.EmptyResult,
		},
		"InvalidLastField": {
			update:     bson.D{{"$mul", bson.D{{"foo", int32(12)}, {"v", "string"}}}},
			resultType: integration.EmptyResult,
		},
		"MultipleOperator": {
			update: bson.D{
				{"$set", bson.D{{"foo", int32(43)}}},
				{"$mul", bson.D{{"v", int32(42)}}},
			},
			providers: providers,
		},
		"ConflictPop": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$pop", bson.D{{"v", -1}}},
			},
			providers:  providers,
			resultType: integration.EmptyResult,
		},
		"ConflictSet": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$set", bson.D{{"v", int32(43)}}},
			},
			providers:  providers,
			resultType: integration.EmptyResult,
		},
		"ConflictInc": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$inc", bson.D{{"v", int32(43)}}},
			},
			providers:  providers,
			resultType: integration.EmptyResult,
		},
		"ConflictMin": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$min", bson.D{{"v", int32(30)}}},
			},
			providers:  providers,
			resultType: integration.EmptyResult,
		},
		"ConflictMax": {
			update: bson.D{
				{"$max", bson.D{{"v", int32(30)}}},
				{"$mul", bson.D{{"v", int32(42)}}},
			},
			providers:  providers,
			resultType: integration.EmptyResult,
		},
		"ConflictSetOnInsert": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$setOnInsert", 43.13},
			},
			providers:  providers,
			resultType: integration.EmptyResult,
		},
		"ConflictUnset": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$unset", bson.D{{"v", ""}}},
			},
			providers:  providers,
			resultType: integration.EmptyResult,
		},
		"ConflictCurrentDate": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$currentDate", bson.D{{"v", bson.D{{"$type", "date"}}}}},
			},
			providers:  providers,
			resultType: integration.EmptyResult,
		},
		"DotNotation": {
			update: bson.D{{"$mul", bson.D{{"v.foo", int32(45)}}}},
		},
		"DotNotationNotExistentPath": {
			update: bson.D{{"$mul", bson.D{{"not.existent.path", int32(45)}}}},
		},
		"DotNotationArrayFieldExist": {
			update: bson.D{{"$mul", bson.D{{"v.array.0", int32(45)}}}},
		},
		"DotNotationArrayFieldNotExist": {
			update: bson.D{{"$mul", bson.D{{"v.array.0.foo", int32(45)}}}},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$mul", bson.D{{"v..", int32(45)}}}},
			resultType: integration.EmptyResult,
		},
		"DotNotationIndexExceedsArrayLength": {
			update: bson.D{{"$mul", bson.D{{"v.100.bar", int32(45)}}}},
		},
		"DotNotationFieldNumericName": {
			update: bson.D{{"$mul", bson.D{{"v.array.42", int32(42)}}}},
		},
		"DotNotationNegativeIndex": {
			update: bson.D{{"$mul", bson.D{{"v.array.-1", int32(42)}}}},
		},
	}

	testUpdateCompat(t, testCases)
}
