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
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestUpdateFieldCompatCurrentDate(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1669")

	t.Parallel()

	testCases := map[string]updateCurrentDateCompatTestCase{
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
		"Timestamp": {
			update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "timestamp"}}}}}},
			paths:  []types.Path{types.NewPathFromString("v")},
		},
		"Date": {
			update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "date"}}}}}},
			paths:  []types.Path{types.NewPathFromString("v")},
		},
		"NoField": {
			update: bson.D{{"$currentDate", bson.D{{"nonexistent", bson.D{{"$type", "date"}}}}}},
			paths: []types.Path{
				types.NewPathFromString("nonexistent"),
			},
		},
	}

	testUpdateCurrentDateCompat(t, testCases)
}

func testUpdateFieldCompatCurrentDate() map[string]updateCurrentDateCollectionParams {
	testCases := map[string]updateCurrentDateCollectionParams{
		"DocumentEmpty": {
			update: bson.D{{"$currentDate", bson.D{}}},
		},
		"ArrayEmpty": {
			update: bson.D{{"$currentDate", bson.A{}}},
		},
		"Int32Wrong": {
			update: bson.D{{"$currentDate", int32(1)}},
		},
		"Nil": {
			update: bson.D{{"$currentDate", nil}},
		},
		"Int32": {
			update: bson.D{{"$currentDate", bson.D{{"v", int32(1)}}}},
		},
		"TimestampCapitalised": {
			update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "Timestamp"}}}}}},
		},
		"WrongType": {
			update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", bson.D{{"abcd", int32(1)}}}}}}}},
		},
		"UnrecognizedOption": {
			update: bson.D{{
				"$currentDate",
				bson.D{{"v", bson.D{{"array", bson.D{{"unexsistent", bson.D{}}}}}}},
			}},
		},
		"DuplicateKeys": {
			update: bson.D{{"$currentDate", bson.D{
				{"v", bson.D{{"$type", "timestamp"}}},
				{"v", bson.D{{"$type", "timestamp"}}},
			}}},
		},
	}

	return testCases
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
		},
		"ArrayFieldIndexNotExist": {
			update: bson.D{{"$inc", bson.D{{"v.5.foo", int32(1)}}}},
		},
	}

	testUpdateCompat(t, testCases)
}

// TestUpdateFieldCompatIncComplex are test that do not work on tigris.
func testUpdateFieldCompatIncUnchanged() map[string]updateCollectionsParams {
	testCases := map[string]updateCollectionsParams{
		"DuplicateKeys": {
			update: bson.D{{"$inc", bson.D{{"v", int32(42)}, {"v", int32(43)}}}},
		},
		"IncOnString": {
			update: bson.D{{"$inc", "string"}},
		},
		"IncWithStringValue": {
			update: bson.D{{"$inc", bson.D{{"v", "bad value"}}}},
		},
		"NotExistStringValue": {
			update: bson.D{{"$inc", bson.D{{"foo.bar", "bad value"}}}},
		},
	}

	return testCases
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

func testUpdateFieldCompatMaxUnchanged() map[string]updateCollectionsParams {
	testCases := map[string]updateCollectionsParams{
		"DuplicateKeys": {
			update: bson.D{{"$max", bson.D{{"v", int32(39)}, {"v", int32(30)}}}},
		},
		"EmptyOperand": {
			update: bson.D{{"$max", bson.D{}}},
		},
	}

	return testCases
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

func testUpdateFieldCompatMinUnchanged() map[string]updateCollectionsParams {
	testCases := map[string]updateCollectionsParams{
		"DuplicateKeys": {
			update: bson.D{{"$min", bson.D{{"v", int32(39)}, {"v", int32(30)}}}},
		},
		"EmptyOperand": {
			update: bson.D{{"$min", bson.D{}}},
		},
	}

	return testCases
}

func TestUpdateFieldCompatRename(t *testing.T) {
	testCases := map[string]updateCompatTestCase{
		"Simple": {
			update: bson.D{{"$rename", bson.D{{"v", "foo"}}}},
		},
		"DotDocumentMove": {
			update:        bson.D{{"$rename", bson.D{{"v.foo", "boo"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1776",
		},
		"DotDocumentDuplicate": {
			update:        bson.D{{"$rename", bson.D{{"v.foo", "v.array"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1776",
		},
		"Multiple": {
			update:        bson.D{{"$rename", bson.D{{"v.foo", "v.bar"}, {"v.42", "v.43"}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1776",
		},
	}

	testUpdateCompat(t, testCases)
}

func testUpdateFieldCompatRenameUnchanged() map[string]updateCollectionsParams {
	testCases := map[string]updateCollectionsParams{
		"DuplicateField": {
			update: bson.D{{"$rename", bson.D{{"v", "v"}}}},
		},
		"NonExistingField": {
			update: bson.D{{"$rename", bson.D{{"foo", "bar"}}}},
		},
		"EmptyField": {
			update: bson.D{{"$rename", bson.D{{"", "v"}}}},
		},
		"EmptyDest": {
			update: bson.D{{"$rename", bson.D{{"v", ""}}}},
		},
		"DotDocumentNonExisting": {
			update: bson.D{{"$rename", bson.D{{"foo.bar", ""}}}},
		},
		"DotArrayField": {
			update: bson.D{{"$rename", bson.D{{"v.array.0", ""}}}},
		},
		"DotArrayNonExisting": {
			update: bson.D{{"$rename", bson.D{{"foo.0.baz", int32(1)}}}},
		},
		"MultipleConflictDestSource": {
			update: bson.D{{"$rename", bson.D{{"v", "foo"}, {"foo", "bar"}}}},
		},
		"MultipleConflictSourceDest": {
			update: bson.D{{"$rename", bson.D{{"v", "foo"}, {"bar", "v"}}}},
		},
		"MultipleConflictDestFields": {
			update: bson.D{{"$rename", bson.D{{"v", "foo"}, {"v", "bar"}}}},
		},
		"MultipleSecondInvalid": {
			update: bson.D{{"$rename", bson.D{{"v.foo", "boo"}, {"v.array", 1}}}},
		},
		"FieldEmpty": {
			update: bson.D{{"$rename", bson.D{}}},
		},
		"InvalidString": {
			update: bson.D{{"$rename", "string"}},
		},
		"InvalidDoc": {
			update: bson.D{{"$rename", primitive.D{}}},
		},
	}

	return testCases
}

func TestUpdateFieldCompatUnset(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Simple": {
			update: bson.D{{"$unset", bson.D{{"v", ""}}}},
		},
		"Nested": {
			update: bson.D{{"$unset", bson.D{{"v", bson.D{{"array", ""}}}}}},
		},
		"DotDocument": {
			update: bson.D{{"$unset", bson.D{{"v.foo", ""}}}},
		},
		"DotArrayField": {
			update: bson.D{{"$unset", bson.D{{"v.array.0", ""}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1242",
		},
	}

	testUpdateCompat(t, testCases)
}

func testUpdateFieldCompatUnsetUnchanged() map[string]updateCollectionsParams {
	testCases := map[string]updateCollectionsParams{
		"NonExisting": {
			update: bson.D{{"$unset", bson.D{{"foo", ""}}}},
		},
		"DotDocumentNonExisting": {
			update: bson.D{{"$unset", bson.D{{"foo.bar", ""}}}},
		},
		"DotArrayNonExisting": {
			update: bson.D{{"$unset", bson.D{{"foo.0.baz", int32(1)}}}},
		},
		"DuplicateKeys": {
			update: bson.D{{"$unset", bson.D{{"v", ""}, {"v", ""}}}},
		},
		"Empty": {
			update: bson.D{{"$unset", bson.D{}}},
		},
		"DocumentField": {
			update: bson.D{{"$unset", bson.D{{"foo", ""}}}},
		},
	}

	return testCases
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
	}

	testUpdateCompat(t, testCases)
}

func testUpdateFieldCompatSetUnchanged() map[string]updateCollectionsParams {
	testCases := map[string]updateCollectionsParams{
		"DuplicateKeys": {
			update: bson.D{{"$set", bson.D{{"v", 42}, {"v", "hello"}}}},
		},
		"NilOperand": {
			update: bson.D{{"$set", nil}},
		},
		"String": {
			update: bson.D{{"$set", "string"}},
		},
		"EmptyDoc": {
			update: bson.D{{"$set", bson.D{}}},
		},
		"Array": {
			update: bson.D{{"$set", bson.A{}}},
		},
	}

	return testCases
}

func TestUpdateFieldCompatSetArray(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Many": {
			update: bson.D{{"$set", bson.D{{"foo", int32(1)}, {"bar", bson.A{}}}}},
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

func testUpdateFieldCompatSetOnInsertUnchanged() map[string]updateCollectionsParams {
	testCases := map[string]updateCollectionsParams{
		"Nil": {
			update: bson.D{{"$setOnInsert", bson.D{{"v", nil}}}},
		},
		"EmptyDoc": {
			update: bson.D{{"$setOnInsert", bson.D{}}},
		},
		"DoubleDouble": {
			update: bson.D{{"$setOnInsert", 43.13}},
		},
		"ErrString": {
			update: bson.D{{"$setOnInsert", "any string"}},
		},
		"ErrNil": {
			update: bson.D{{"$setOnInsert", nil}},
		},
		"DocumentFieldExist": {
			update: bson.D{{"$setOnInsert", bson.D{{"v.foo", int32(1)}}}},
		},
		"DocumentFieldNotExist": {
			update: bson.D{{"$setOnInsert", bson.D{{"foo.bar", int32(1)}}}},
		},
		"ArrayFieldExist": {
			update: bson.D{{"$setOnInsert", bson.D{{"v.array.0", int32(1)}}}},
		},
		"ArrFieldNotExist": {
			update: bson.D{{"$setOnInsert", bson.D{{"foo.0.baz", int32(1)}}}},
		},
		"DocArrFieldNotExist": {
			update: bson.D{{"$setOnInsert", bson.D{{"v.0.foo", int32(1)}}}},
		},
		"DuplicateKeys": {
			update: bson.D{{"$setOnInsert", bson.D{{"v", 1}, {"v", 2}}}},
		},
	}

	return testCases
}

func testUpdateFieldCompatSetOnInsertArrayUnchaged() map[string]updateCollectionsParams {
	testCases := map[string]updateCollectionsParams{
		"Array": {
			update: bson.D{{"$setOnInsert", bson.D{{"v", bson.A{}}}}},
		},
		"EmptyArray": {
			update: bson.D{{"$setOnInsert", bson.A{}}},
		},
	}

	return testCases
}

func testUpdateFieldCompatMixedUnchanged() map[string]updateCollectionsParams {
	testCases := map[string]updateCollectionsParams{
		"SetSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$setOnInsert", bson.D{{"v", nil}}},
			},
		},
		"SetIncSetOnInsert": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{
				{"$set", bson.D{{"foo", int32(12)}}},
				{"$inc", bson.D{{"foo", int32(1)}}},
				{"$setOnInsert", bson.D{{"v", nil}}},
			},
		},
		"UnknownOperator": {
			filter: bson.D{{"_id", "test"}},
			update: bson.D{{"$foo", bson.D{{"foo", int32(1)}}}},
		},
	}

	return testCases
}

func TestUpdateFieldCompatMul(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1668")

	t.Parallel()

	providers := shareddata.AllProviders().
		// BigDoubles and Scalars contain numbers that produces +INF on compat,
		// validation error on target upon $mul operation.
		Remove("BigDoubles", "Scalars")

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
			update:    bson.D{{"$mul", bson.D{{"v", float64(2 << 60)}}}},
			providers: providers,
		},
		"FieldNotExist": {
			update: bson.D{{"$mul", bson.D{{"foo", int32(45)}}}},
		},
		"DocFieldExist": {
			update: bson.D{{"$mul", bson.D{{"v.foo", int32(45)}}}},
		},
		"DocFieldNotExist": {
			update: bson.D{{"$mul", bson.D{{"foo.bar", int32(45)}}}},
		},
		"ArrayFieldExist": {
			update: bson.D{{"$mul", bson.D{{"v.array.0", int32(45)}}}},
		},
		"ArrayFieldNotExist": {
			update: bson.D{{"$mul", bson.D{{"v.array.foo", int32(45)}}}},
		},
		"DocArrayFieldNotExist": {
			update: bson.D{{"$mul", bson.D{{"foo.0.baz", int32(45)}}}},
		},
		"TwoFields": {
			update: bson.D{{"$mul", bson.D{{"foo", int32(12)}, {"v", int32(1)}}}},
		},
		"MultipleOperator": {
			update: bson.D{
				{"$set", bson.D{{"foo", int32(43)}}},
				{"$mul", bson.D{{"v", int32(42)}}},
			},
			providers: providers,
		},
	}

	testUpdateCompat(t, testCases)
}

func testUpdateFieldCompatMulUnchanged() map[string]updateCollectionsParams {
	testCases := map[string]updateCollectionsParams{
		"Empty": {
			update: bson.D{{"$mul", bson.D{}}},
		},
		"Null": {
			update: bson.D{{"$mul", bson.D{{"v", nil}}}},
		},
		"String": {
			update: bson.D{{"$mul", bson.D{{"v", "string"}}}},
		},
		"MissingField": {
			update: bson.D{{"$mul", "invalid"}},
		},
		"StringFieldNotExist": {
			update: bson.D{{"$mul", bson.D{{"foo.bar", "bad value"}}}},
		},
		"DuplicateKeys": {
			update: bson.D{{"$mul", bson.D{{"v", int32(42)}, {"v", int32(43)}}}},
		},
		"InvalidLastField": {
			update: bson.D{{"$mul", bson.D{{"foo", int32(12)}, {"v", "string"}}}},
		},
		"ConflictPop": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$pop", bson.D{{"v", -1}}},
			},
		},
		"ConflictSet": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$set", bson.D{{"v", int32(43)}}},
			},
		},
		"ConflictInc": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$inc", bson.D{{"v", int32(43)}}},
			},
		},
		"ConflictMin": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$min", bson.D{{"v", int32(30)}}},
			},
		},
		"ConflictMax": {
			update: bson.D{
				{"$max", bson.D{{"v", int32(30)}}},
				{"$mul", bson.D{{"v", int32(42)}}},
			},
		},
		"ConflictSetOnInsert": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$setOnInsert", 43.13},
			},
		},
		"ConflictUnset": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$unset", bson.D{{"v", ""}}},
			},
		},
		"ConflictCurrentDate": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$currentDate", bson.D{{"v", bson.D{{"$type", "date"}}}}},
			},
		},
	}

	return testCases
}
