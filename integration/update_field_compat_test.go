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
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
)

func TestUpdateFieldCompatCurrentDate(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1669")

	t.Parallel()

	testCases := map[string]updateCurrentDateCompatTestCase{
		"DocumentEmpty": {
			update:     bson.D{{"$currentDate", bson.D{}}},
			resultType: emptyResult,
		},
		"ArrayEmpty": {
			update:     bson.D{{"$currentDate", bson.A{}}},
			resultType: emptyResult,
		},
		"Int32Wrong": {
			update:     bson.D{{"$currentDate", int32(1)}},
			resultType: emptyResult,
		},
		"Nil": {
			update:     bson.D{{"$currentDate", nil}},
			resultType: emptyResult,
		},
		"BoolTrue": {
			update: bson.D{{"$currentDate", bson.D{{"v", true}}}},
			paths:  []types.Path{types.NewPathFromString("v")},
		},
		"BoolTwoTrue": {
			update: bson.D{{"$currentDate", bson.D{{"v", true}, {"nonexistent", true}}}},
			paths: []types.Path{
				types.NewPathFromString("v"),
				types.NewPathFromString("nonexistent"),
			},
		},
		"BoolFalse": {
			update: bson.D{{"$currentDate", bson.D{{"v", false}}}},
			paths:  []types.Path{types.NewPathFromString("v")},
		},
		"Int32": {
			update:     bson.D{{"$currentDate", bson.D{{"v", int32(1)}}}},
			resultType: emptyResult,
		},
		"Timestamp": {
			update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "timestamp"}}}}}},
			paths:  []types.Path{types.NewPathFromString("v")},
		},
		"TimestampCapitalised": {
			update:     bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "Timestamp"}}}}}},
			resultType: emptyResult,
		},
		"Date": {
			update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "date"}}}}}},
			paths:  []types.Path{types.NewPathFromString("v")},
		},
		"WrongType": {
			update:     bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", bson.D{{"abcd", int32(1)}}}}}}}},
			resultType: emptyResult,
		},
		"NoField": {
			update: bson.D{{"$currentDate", bson.D{{"nonexistent", bson.D{{"$type", "date"}}}}}},
			paths: []types.Path{
				types.NewPathFromString("nonexistent"),
			},
		},
		"UnrecognizedOption": {
			update: bson.D{{
				"$currentDate",
				bson.D{{"v", bson.D{{"array", bson.D{{"unexsistent", bson.D{}}}}}}},
			}},
			resultType: emptyResult,
		},
		"DuplicateKeys": {
			update: bson.D{{"$currentDate", bson.D{
				{"v", bson.D{{"$type", "timestamp"}}},
				{"v", bson.D{{"$type", "timestamp"}}},
			}}},
			resultType: emptyResult,
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
		"DotNotationFieldExist": {
			update:        bson.D{{"$inc", bson.D{{"v.foo", int32(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1088",
		},
		"DotNotationFieldNotExist": {
			update:        bson.D{{"$inc", bson.D{{"foo.bar", int32(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1088",
		},
		"DuplicateKeys": {
			update:     bson.D{{"$inc", bson.D{{"v", int32(42)}, {"v", int32(43)}}}},
			resultType: emptyResult,
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
			update: bson.D{{"$inc", bson.D{{"v", float64(2 << 60)}}}},
		},
		"DoubleIncOnNullValue": {
			update: bson.D{{"$inc", bson.D{{"v", float64(1)}}}},
		},
		"ArrayFieldExist": {
			update: bson.D{{"$inc", bson.D{{"v.array.0", int32(1)}}}},
		},
		"DocFieldNotExist": {
			update: bson.D{{"$inc", bson.D{{"foo.bar", int32(1)}}}},
		},
		"ArrayFieldNotExist": {
			update: bson.D{{"$inc", bson.D{{"v.array.foo", int32(1)}}}},
		},
		"FieldNotExist": {
			update:        bson.D{{"$inc", bson.D{{"foo", int32(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1676",
		},
		"DocFieldExist": {
			update: bson.D{{"$inc", bson.D{{"v.foo", int32(1)}}}},
		},
		"DocArrayFieldNotExist": {
			update: bson.D{{"$inc", bson.D{{"foo.0.baz", int32(1)}}}},
		},
		"ArrayFieldValueNotExist": {
			update: bson.D{{"$inc", bson.D{{"v.0.foo", int32(1)}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1658",
		},
		"IncOnString": {
			update:     bson.D{{"$inc", "string"}},
			resultType: emptyResult,
		},
		"IncWithStringValue": {
			update:     bson.D{{"$inc", bson.D{{"v", "bad value"}}}},
			resultType: emptyResult,
		},
		"NotExistStringValue": {
			update:     bson.D{{"$inc", bson.D{{"foo.bar", "bad value"}}}},
			resultType: emptyResult,
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
			resultType: emptyResult,
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
			resultType: emptyResult,
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
			resultType: emptyResult,
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
			resultType: emptyResult,
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
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatRename(t *testing.T) {
	testCases := map[string]updateCompatTestCase{
		"Simple": {
			update: bson.D{{"$rename", bson.D{{"v", "foo"}}}},
		},
		"DuplicateField": {
			update:     bson.D{{"$rename", bson.D{{"v", "v"}}}},
			resultType: emptyResult,
		},
		"NonExistingField": {
			update:     bson.D{{"$rename", bson.D{{"foo", "bar"}}}},
			resultType: emptyResult,
		},
		"EmptyField": {
			update:     bson.D{{"$rename", bson.D{{"", "v"}}}},
			resultType: emptyResult,
		},
		"EmptyDest": {
			update:     bson.D{{"$rename", bson.D{{"v", ""}}}},
			resultType: emptyResult,
		},
		"DotDocumentMove": {
			update:        bson.D{{"$rename", bson.D{{"v.foo", "boo"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1776",
		},
		"DotDocumentDuplicate": {
			update:        bson.D{{"$rename", bson.D{{"v.foo", "v.array"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1776",
		},
		"DotDocumentNonExisting": {
			update:     bson.D{{"$rename", bson.D{{"foo.bar", ""}}}},
			resultType: emptyResult,
		},
		"DotArrayField": {
			update:     bson.D{{"$rename", bson.D{{"v.array.0", ""}}}},
			resultType: emptyResult,
		},
		"DotArrayNonExisting": {
			update:     bson.D{{"$rename", bson.D{{"foo.0.baz", int32(1)}}}},
			resultType: emptyResult,
		},
		"Multiple": {
			update:        bson.D{{"$rename", bson.D{{"v.foo", "v.bar"}, {"v.42", "v.43"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1776",
		},
		"MultipleConflictDestSource": {
			update:     bson.D{{"$rename", bson.D{{"v", "foo"}, {"foo", "bar"}}}},
			resultType: emptyResult,
		},
		"MultipleConflictSourceDest": {
			update:     bson.D{{"$rename", bson.D{{"v", "foo"}, {"bar", "v"}}}},
			resultType: emptyResult,
		},
		"MultipleConflictDestFields": {
			update:     bson.D{{"$rename", bson.D{{"v", "foo"}, {"v", "bar"}}}},
			resultType: emptyResult,
		},
		"MultipleSecondInvalid": {
			update:     bson.D{{"$rename", bson.D{{"v.foo", "boo"}, {"v.array", 1}}}},
			resultType: emptyResult,
		},
		"FieldEmpty": {
			update:     bson.D{{"$rename", bson.D{}}},
			resultType: emptyResult,
		},
		"InvalidString": {
			update:     bson.D{{"$rename", "string"}},
			resultType: emptyResult,
		},
		"InvalidDoc": {
			update:     bson.D{{"$rename", primitive.D{}}},
			resultType: emptyResult,
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
			resultType: emptyResult,
		},
		"Nested": {
			update: bson.D{{"$unset", bson.D{{"v", bson.D{{"array", ""}}}}}},
		},
		"DotDocument": {
			update: bson.D{{"$unset", bson.D{{"v.foo", ""}}}},
		},
		"DotDocumentNonExisting": {
			update:     bson.D{{"$unset", bson.D{{"foo.bar", ""}}}},
			resultType: emptyResult,
		},
		"DotArrayField": {
			update: bson.D{{"$unset", bson.D{{"v.array.0", ""}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1242",
		},
		"DotArrayNonExisting": {
			update:     bson.D{{"$unset", bson.D{{"foo.0.baz", int32(1)}}}},
			resultType: emptyResult,
		},
		"DuplicateKeys": {
			update:     bson.D{{"$unset", bson.D{{"v", ""}, {"v", ""}}}},
			resultType: emptyResult,
		},
		"Empty": {
			update:     bson.D{{"$unset", bson.D{}}},
			resultType: emptyResult,
		},
		"DocumentField": {
			update:     bson.D{{"$unset", bson.D{{"foo", ""}}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatUnsetArray(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"EmptyArray": {
			update:     bson.D{{"$unset", bson.A{}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"SetNullInExisingField": {
			update: bson.D{{"$set", bson.D{{"v", nil}}}},
		},
		"DuplicateKeys": {
			update:     bson.D{{"$set", bson.D{{"v", 42}, {"v", "hello"}}}},
			resultType: emptyResult,
		},
		"NilOperand": {
			update:     bson.D{{"$set", nil}},
			resultType: emptyResult,
		},
		"String": {
			update:     bson.D{{"$set", "string"}},
			resultType: emptyResult,
		},
		"EmptyDoc": {
			update:     bson.D{{"$set", bson.D{}}},
			resultType: emptyResult,
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
		"SetSameValueInt": {
			update: bson.D{{"$set", bson.D{{"v", int32(42)}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1662",
		},
		"DocFieldExist": {
			update: bson.D{{"$set", bson.D{{"v.foo", int32(1)}}}},
		},
		"DocumentFieldNotExist": {
			update:        bson.D{{"$set", bson.D{{"foo.bar", int32(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1676",
		},
		"ArrayFieldExist": {
			update: bson.D{{"$set", bson.D{{"v.array.0", int32(1)}}}},
		},
		"ArrayFieldNotExist": {
			update:        bson.D{{"$set", bson.D{{"foo.0.baz", int32(1)}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1676",
		},
		"DocArrFieldNotExist": {
			update: bson.D{{"$set", bson.D{{"v.0.foo", int32(1)}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1661",
		},
		"DocumentField": {
			update:        bson.D{{"$set", bson.D{{"foo", int32(42)}, {"bar", "baz"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1676",
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
			resultType: emptyResult,
		},
		"ArrayNil": {
			update: bson.D{{"$set", bson.D{{"v", bson.A{nil}}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1662",
		},
		"EmptyArray": {
			update:        bson.D{{"$set", bson.D{{"v", bson.A{}}}}},
			skipForTigris: `Internal error when set "v":[] https://github.com/FerretDB/FerretDB/issues/1704`,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatSetOnInsert(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Nil": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v", nil}}}},
			resultType: emptyResult,
		},
		"EmptyDoc": {
			update:     bson.D{{"$setOnInsert", bson.D{}}},
			resultType: emptyResult,
		},
		"DoubleDouble": {
			update:     bson.D{{"$setOnInsert", 43.13}},
			resultType: emptyResult,
		},
		"ErrString": {
			update:     bson.D{{"$setOnInsert", "any string"}},
			resultType: emptyResult,
		},
		"ErrNil": {
			update:     bson.D{{"$setOnInsert", nil}},
			resultType: emptyResult,
		},
		"DocumentFieldExist": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v.foo", int32(1)}}}},
			resultType: emptyResult,
		},
		"DocumentFieldNotExist": {
			update:     bson.D{{"$setOnInsert", bson.D{{"foo.bar", int32(1)}}}},
			resultType: emptyResult,
		},
		"ArrayFieldExist": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v.array.0", int32(1)}}}},
			resultType: emptyResult,
		},
		"ArrFieldNotExist": {
			update:     bson.D{{"$setOnInsert", bson.D{{"foo.0.baz", int32(1)}}}},
			resultType: emptyResult,
		},
		"DocArrFieldNotExist": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v.0.foo", int32(1)}}}},
			resultType: emptyResult,
		},
		"DuplicateKeys": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v", 1}, {"v", 2}}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatSetOnInsertArray(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Array": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v", bson.A{}}}}},
			resultType: emptyResult,
		},
		"EmptyArray": {
			update:     bson.D{{"$setOnInsert", bson.A{}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatPop(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"DuplicateKeys": {
			update:     bson.D{{"$pop", bson.D{{"v", 1}, {"v", 1}}}},
			resultType: emptyResult,
		},
		"Pop": {
			update:        bson.D{{"$pop", bson.D{{"v", 1}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1677",
		},
		"PopFirst": {
			update:        bson.D{{"$pop", bson.D{{"v", -1}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1677",
		},
		"PopDotNotation": {
			update: bson.D{{"$pop", bson.D{{"v.array", 1}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1663",
		},
		"PopNoSuchKey": {
			update:     bson.D{{"$pop", bson.D{{"foo", 1}}}},
			resultType: emptyResult,
		},
		"PopEmptyValue": {
			update:     bson.D{{"$pop", bson.D{}}},
			resultType: emptyResult,
		},
		"PopNotValidValueString": {
			update:     bson.D{{"$pop", bson.D{{"v", "foo"}}}},
			resultType: emptyResult,
		},
		"PopNotValidValueInt": {
			update:     bson.D{{"$pop", bson.D{{"v", int32(42)}}}},
			resultType: emptyResult,
		},
		"PopLastAndFirst": {
			update: bson.D{{"$pop", bson.D{{"v", 1}, {"v", -1}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/666",
		},
		"PopDotNotationNonArray": {
			update: bson.D{{"$pop", bson.D{{"v.foo", 1}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1663",
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
			resultType: emptyResult,
		},
		"SetIncSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$inc", bson.D{{"foo", int32(1)}}},
				{"$setOnInsert", bson.D{{"v", nil}}},
			},
			resultType: emptyResult,
		},
		"UnknownOperator": {
			filter:     bson.D{{"_id", "test"}},
			update:     bson.D{{"$foo", bson.D{{"foo", int32(1)}}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}
