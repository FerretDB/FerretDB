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
	"fmt"
	"math"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"

	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func TestUpdateFieldCompatCurrentDate(t *testing.T) {
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
			keys:   []string{"v"},
		},
		"BoolTwoTrue": {
			update: bson.D{{"$currentDate", bson.D{{"v", true}, {"nonexistent", true}}}},
			keys:   []string{"v", "nonexistent"},
		},
		"BoolFalse": {
			update: bson.D{{"$currentDate", bson.D{{"v", false}}}},
			keys:   []string{"v"},
		},
		"Int32": {
			update:     bson.D{{"$currentDate", bson.D{{"v", int32(1)}}}},
			resultType: emptyResult,
		},
		"Timestamp": {
			update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "timestamp"}}}}}},
			keys:   []string{"v"},
		},
		"TimestampCapitalised": {
			update:     bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "Timestamp"}}}}}},
			resultType: emptyResult,
		},
		"Date": {
			update: bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", "date"}}}}}},
			keys:   []string{"v"},
		},
		"WrongType": {
			update:     bson.D{{"$currentDate", bson.D{{"v", bson.D{{"$type", bson.D{{"abcd", int32(1)}}}}}}}},
			resultType: emptyResult,
		},
		"NoField": {
			update: bson.D{{"$currentDate", bson.D{{"nonexistent", bson.D{{"$type", "date"}}}}}},
			keys:   []string{"nonexistent"},
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
			update:           bson.D{{"$inc", bson.D{{"v", int32(42)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
				}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"Int32Negative": {
			update:           bson.D{{"$inc", bson.D{{"v", int32(-42)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
				}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"Int64Max": {
			update:           bson.D{{"$inc", bson.D{{"v", math.MaxInt64}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{"bool-false", "bool-true", "double-5", "double-max"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"Int64Min": {
			update:           bson.D{{"$inc", bson.D{{"v", math.MinInt64}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{"bool-false", "bool-true", "double-5", "double-max"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"EmptyUpdatePath": {
			update: bson.D{{"$inc", bson.D{{}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/673",
		},
		"WrongIncTypeArray": {
			update:     bson.D{{"$inc", bson.A{}}},
			resultType: emptyResult,
		},
		"DuplicateKeys": {
			update:     bson.D{{"$inc", bson.D{{"v", int32(42)}, {"v", int32(43)}}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatIncComplex(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"IntNegativeIncrement": {
			update:           bson.D{{"$inc", bson.D{{"v", int32(-1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
					"double-prec-min-minus", "double-prec-min-minus-two",
				}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"DoubleIncrement": {
			update:           bson.D{{"$inc", bson.D{{"v", float64(42.13)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
				}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"LongNegativeIncrement": {
			update:           bson.D{{"$inc", bson.D{{"v", int64(-1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
					"double-prec-min-minus", "double-prec-min-minus-two",
				}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"IncTwoFields": {
			update:           bson.D{{"$inc", bson.D{{"foo", int32(12)}, {"v", int32(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp"}},
			},
		},
		"DoubleBigDoubleIncrement": {
			update:           bson.D{{"$inc", bson.D{{"v", 42.13}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
				}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"DoubleIntIncrement": {
			update:           bson.D{{"$inc", bson.D{{"v", int32(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
					"double-prec-max-plus", "double-prec-max-plus-two",
				}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"IntOverflow": {
			update:           bson.D{{"$inc", bson.D{{"v", math.MaxInt64}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{"bool-false", "bool-true", "double-5", "double-max"}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"DoubleIncrementIntField": {
			update:           bson.D{{"$inc", bson.D{{"v", float64(1.13)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
				}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"DoubleLongIncrement": {
			update:           bson.D{{"$inc", bson.D{{"v", int64(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
					"double-prec-max-plus", "double-prec-max-plus-two",
				}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"DoubleNegativeIncrement": {
			update:           bson.D{{"$inc", bson.D{{"v", float64(-42.13)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
				}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"DoubleDoubleBigIncrement": {
			update:           bson.D{{"$inc", bson.D{{"v", float64(1 << 61)}}}}, // TODO https://github.com/FerretDB/FerretDB/issues/3626
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{"bool-false", "bool-true", "double-5", "double-max"}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"DoubleIncOnNullValue": {
			update:           bson.D{{"$inc", bson.D{{"v", float64(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"bool-false", "bool-true", "double-1", "double-4", "double-5",
					"double-big", "double-max", "double-max-overflow", "double-min-overflow", "int64-double-big",
				}},
				{provider: shareddata.Bools, ids: []string{"bool-false", "bool-true"}},
				{provider: shareddata.OverflowVergeDoubles, ids: []string{"double-7", "double-max"}},
				{provider: shareddata.Doubles, ids: []string{
					"double-1", "double-4", "double-big", "double-big-minus", "double-big-plus",
					"double-max-overflow", "double-min-overflow", "double-neg-big", "double-neg-big-minus", "double-neg-big-plus",
					"double-prec-max-plus", "double-prec-max-plus-two",
				}},
				{provider: shareddata.Decimal128s, ids: []string{"decimal128-max-exp", "decimal128-max-exp-sig"}},
			},
		},
		"FieldNotExist": {
			update: bson.D{{"$inc", bson.D{{"foo", int32(1)}}}},
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
		"DotNotationFieldExist": {
			update:           bson.D{{"$inc", bson.D{{"v.foo", int32(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents",
					"array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayInt64s},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null"}},
				{provider: shareddata.ArrayDocuments},
			},
		},
		"DotNotationArrayValue": {
			update:           bson.D{{"$inc", bson.D{{"v.0", int32(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.ArrayDoubles, ids: []string{"array-double-big", "array-double-big-plus", "array-double-prec-max-plus"}},
			},
		},
		"DotNotationFieldNotExist": {
			update: bson.D{{"$inc", bson.D{{"not.existent.path", int32(1)}}}},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$inc", bson.D{{"v..", int32(42)}}}},
			resultType: emptyResult,
		},
		"DotNotationNegativeIndex": {
			update:           bson.D{{"$inc", bson.D{{"v.-1", int32(42)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-null", "array-numbers-asc",
					"array-numbers-desc", "array-three", "array-three-reverse", "array-two", "array-strings-desc",
					"array-documents", "array-empty",
				}},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
				{provider: shareddata.ArrayDocuments, ids: []string{
					"array-documents-nested", "array-documents-nested-duplicate",
					"array-three-documents", "array-two-documents",
				}},
				{provider: shareddata.ArrayInt32s, ids: []string{
					"array-int32-one", "array-int32-two", "array-int32-three",
					"array-int32-six", "array-int32-empty",
				}},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayStrings, ids: []string{
					"array-string-desc", "array-string-duplicate",
					"array-string-numbers", "array-string-with-nil", "array-string-empty",
				}},
				{
					provider: shareddata.ArrayDoubles,
					ids: []string{
						"array-double-big", "array-double-big-plus", "array-double-desc",
						"array-double-duplicate", "array-double-prec-max", "array-double-prec-max-plus",
						"document-double-nil", "array-double-empty",
					},
				},
			},
		},
		"DotNotatIndexOutOfArray": {
			update: bson.D{{"$inc", bson.D{{"v.100", int32(42)}}}},
		},
		"DotNotatArrayFieldNotExist": {
			update:           bson.D{{"$inc", bson.D{{"v.array.foo", int32(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-null", "array-numbers-asc", "array-numbers-desc",
					"array-three", "array-three-reverse", "array-two", "array-strings-desc", "array-documents," +
						"array-empty", "document-composite", "document-composite-reverse", "array-documents", "array-empty",
				}},
				{provider: shareddata.ArrayStrings, ids: []string{
					"array-string-desc", "array-string-duplicate", "array-string-numbers",
					"array-string-with-nil", "array-string-empty",
				}},
				{provider: shareddata.ArrayInt32s, ids: []string{"array-int32-one", "array-int32-two", "array-int32-three", "array-int32-six", "array-int32-empty"}},
				{provider: shareddata.Mixed, ids: []string{"array-null", "array-empty"}},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDoubles, ids: []string{
					"array-double-big", "array-double-big-plus", "array-double-desc",
					"array-double-duplicate", "array-double-prec-max", "array-double-prec-max-plus",
					"document-double-nil", "array-double-empty",
				}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
				{provider: shareddata.ArrayDocuments, ids: []string{
					"array-documents-nested", "array-documents-nested-duplicate",
					"array-three-documents", "array-two-documents",
				}},
			},
		},
		"DotNotatArrFieldExist": {
			update:           bson.D{{"$inc", bson.D{{"v.array.0", int32(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/421",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-null", "array-numbers-asc", "array-numbers-desc",
					"array-three", "array-three-reverse", "array-two", "array-strings-desc", "array-documents",
					"array-empty",
				}},
				{provider: shareddata.ArrayStrings, ids: []string{
					"array-string-desc", "array-string-duplicate", "array-string-numbers",
					"array-string-with-nil", "array-string-empty",
				}},
				{provider: shareddata.ArrayInt32s, ids: []string{"array-int32-one", "array-int32-two", "array-int32-three", "array-int32-six", "array-int32-empty"}},
				{provider: shareddata.Mixed, ids: []string{"array-null", "array-empty"}},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDoubles, ids: []string{
					"array-double-big", "array-double-big-plus", "array-double-desc",
					"array-double-duplicate", "array-double-prec-max", "array-double-prec-max-plus",
					"document-double-nil", "array-double-empty",
				}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
				{provider: shareddata.ArrayDocuments, ids: []string{
					"array-documents-nested", "array-documents-nested-duplicate",
					"array-three-documents", "array-two-documents",
				}},
			},
		},
		"DotNotatArrFieldValue": {
			update: bson.D{{"$inc", bson.D{{"v.0.foo", int32(1)}}}},
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatIncMulti(t *testing.T) {
	t.Parallel()

	testCases := map[string]testUpdateManyCompatTestCase{
		"InvalidInc": {
			filter:     bson.D{{"v", bson.D{{"$eq", "non-existent"}}}},
			update:     bson.D{{"$inc", bson.D{{"v", 1}}}},
			updateOpts: options.Update().SetUpsert(true),
			providers:  []shareddata.Provider{shareddata.Scalars},
			skip:       "https://github.com/FerretDB/FerretDB/issues/3044",
		},
	}

	testUpdateManyCompat(t, testCases)
}

func TestUpdateFieldCompatMax(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Int32Lower": {
			update: bson.D{{"$max", bson.D{{"v", int32(30)}}}},
		},
		"Int32Higher": {
			update: bson.D{{"$max", bson.D{{"v", int32(60)}}}},
		},
		"Int32Negative": {
			update: bson.D{{"$max", bson.D{{"v", int32(-22)}}}},
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
			update: bson.D{{"$max", bson.D{{"v", -54.32}}}},
		},
		"NotExisting": {
			update: bson.D{{"$max", bson.D{{"v", int32(60)}}}},
		},

		"MultipleQueries": {
			update: bson.D{{"$max", bson.D{{"a", int32(30)}, {"v", int32(39)}}}},
		},
		"MultipleQueriesSorted": {
			update: bson.D{{"$max", bson.D{{"v", int32(39)}, {"a", int32(30)}}}},
		},
		"DuplicateKeys": {
			update:     bson.D{{"$max", bson.D{{"v", int32(39)}, {"v", int32(30)}}}},
			resultType: emptyResult,
		},

		// Strings are not converted to numbers
		"StringIntegerHigher": {
			update: bson.D{{"$max", bson.D{{"v", "60"}}}},
		},
		"StringIntegerLower": {
			update: bson.D{{"$max", bson.D{{"v", "30"}}}},
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
			update: bson.D{{"$max", bson.D{{"v", false}}}},
		},
		"EmptyOperand": {
			update:     bson.D{{"$max", bson.D{}}},
			resultType: emptyResult,
		},
		"DateTime": {
			update: bson.D{{"$max", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 12, 18, 42, 123000000, time.UTC))}}}},
		},
		"DateTimeLower": {
			update: bson.D{{"$max", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 3, 18, 42, 123000000, time.UTC))}}}},
		},
		"ArrayEmpty": {
			update: bson.D{{"$max", bson.D{{"v", bson.A{}}}}},
		},
		"ArrayOne": {
			update: bson.D{{"$max", bson.D{{"v", bson.A{int32(42)}}}}},
		},
		"Array": {
			update: bson.D{{"$max", bson.D{{"v", bson.A{int32(42), "foo", nil}}}}},
		},
		"ArrayReverse": {
			update: bson.D{{"$max", bson.D{{"v", bson.A{nil, "foo", int32(42)}}}}},
		},
		"ArrayNull": {
			update: bson.D{{"$max", bson.D{{"v", bson.A{nil}}}}},
		},
		"ArraySlice": {
			update: bson.D{{"$max", bson.D{{"v", bson.A{int32(42), "foo"}}}}},
		},
		"ArrayShuffledValues": {
			update: bson.D{{"$max", bson.D{{"v", bson.A{"foo", nil, int32(42)}}}}},
		},
		"ArrayDocuments": {
			update: bson.D{{"$max", bson.D{{"v", bson.A{bson.D{{"foo", int32(42)}}, bson.D{{"foo", nil}}}}}}},
		},
		"DotNotation": {
			update:           bson.D{{"$max", bson.D{{"v.foo", int32(42)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$max", bson.D{{"v..", int32(42)}}}},
			resultType: emptyResult,
		},
		"DotNotationNegativeIndex": {
			update:           bson.D{{"$max", bson.D{{"v.-1", int32(42)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationIndexOutsideArray": {
			update:           bson.D{{"$max", bson.D{{"v.100", int32(42)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.Mixed, ids: []string{"null"}},
			},
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatMin(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Int32Lower": {
			update: bson.D{{"$min", bson.D{{"v", int32(30)}}}},
		},
		"Int32Higher": {
			update: bson.D{{"$min", bson.D{{"v", int32(60)}}}},
		},
		"Int32Negative": {
			update: bson.D{{"$min", bson.D{{"v", int32(-22)}}}},
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
			update: bson.D{{"$min", bson.D{{"v", -54.32}}}},
		},
		"NotExisting": {
			update: bson.D{{"$min", bson.D{{"v", int32(60)}}}},
		},
		"MultipleQueries": {
			update: bson.D{{"$min", bson.D{{"a", int32(30)}, {"v", int32(39)}}}},
		},
		"MultipleQueriesSorted": {
			update: bson.D{{"$min", bson.D{{"v", int32(39)}, {"a", int32(30)}}}},
		},
		"DuplicateKeys": {
			update:     bson.D{{"$min", bson.D{{"v", int32(39)}, {"v", int32(30)}}}},
			resultType: emptyResult,
		},
		"StringIntegerHigher": {
			update: bson.D{{"$min", bson.D{{"v", "60"}}}},
		},
		"StringIntegerLower": {
			update: bson.D{{"$min", bson.D{{"v", "30"}}}},
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
			update: bson.D{{"$min", bson.D{{"v", true}}}},
		},
		"BoolFalse": {
			update: bson.D{{"$min", bson.D{{"v", false}}}},
		},
		"EmptyOperand": {
			update:     bson.D{{"$min", bson.D{}}},
			resultType: emptyResult,
		},
		"DateTime": {
			update: bson.D{{"$min", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 12, 18, 42, 123000000, time.UTC))}}}},
		},
		"DateTimeLower": {
			update: bson.D{{"$min", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 3, 18, 42, 123000000, time.UTC))}}}},
		},
		"ArrayEmpty": {
			update: bson.D{{"$min", bson.D{{"v", bson.A{}}}}},
		},
		"ArrayOne": {
			update: bson.D{{"$min", bson.D{{"v", bson.A{int32(42)}}}}},
		},
		"Array": {
			update: bson.D{{"$min", bson.D{{"v", bson.A{int32(42), "foo", nil}}}}},
		},
		"ArrayReverse": {
			update: bson.D{{"$min", bson.D{{"v", bson.A{nil, "foo", int32(42)}}}}},
		},
		"ArrayNull": {
			update: bson.D{{"$min", bson.D{{"v", bson.A{nil}}}}},
		},
		"ArraySlice": {
			update: bson.D{{"$min", bson.D{{"v", bson.A{int32(42), "foo"}}}}},
		},
		"ArrayShuffledValues": {
			update: bson.D{{"$min", bson.D{{"v", bson.A{"foo", nil, int32(42)}}}}},
		},
		"ArrayDocuments": {
			update: bson.D{{"$min", bson.D{{"v", bson.A{bson.D{{"foo", int32(42)}}, bson.D{{"foo", nil}}}}}}},
		},
		"DotNotation": {
			update:           bson.D{{"$min", bson.D{{"v.foo", int32(42)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$min", bson.D{{"v..", int32(42)}}}},
			resultType: emptyResult,
		},
		"DotNotationNegativeIndex": {
			update:           bson.D{{"$min", bson.D{{"v.-1", int32(42)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationIndexOutOfArray": {
			update:           bson.D{{"$min", bson.D{{"v.100", int32(42)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.Mixed, ids: []string{"null"}},
			},
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
			update:           bson.D{{"$rename", bson.D{{"v", "v"}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
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
		"DotNotationDocumentMove": {
			update:           bson.D{{"$rename", bson.D{{"v.foo", "boo"}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationDocumentDuplicate": {
			update:           bson.D{{"$rename", bson.D{{"v.foo", "v.array"}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/448", // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/429
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/448", // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/429
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationDocNonExistent": {
			update:     bson.D{{"$rename", bson.D{{"not.existent.path", ""}}}},
			resultType: emptyResult,
		},
		"DotNotationArrayField": {
			update:     bson.D{{"$rename", bson.D{{"v.array.0", ""}}}},
			resultType: emptyResult,
		},
		"DotNotationArrayNonExisting": {
			update:     bson.D{{"$rename", bson.D{{"foo.0.baz", int32(1)}}}},
			resultType: emptyResult,
		},
		"DotNotationMultipleFields": {
			update:           bson.D{{"$rename", bson.D{{"v.foo", "v.bar"}, {"v.42", "v.43"}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/448", // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/429
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/448", // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/429
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$rename", bson.D{{"v..", "v.bar"}}}},
			resultType: emptyResult,
		},
		"DotNotationNegativeIndex": {
			update:           bson.D{{"$rename", bson.D{{"v.-1.bar", "v.-1.baz"}}}},
			resultType:       emptyResult,
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/448", // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/429
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/448", // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/429
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationIndexOutOfArray": {
			update:           bson.D{{"$rename", bson.D{{"v.100.bar", "v.100.baz"}}}},
			resultType:       emptyResult,
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/448", // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/449
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/448", // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/449
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{"array-empty"}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings, ids: []string{"array-string-empty"}},
				{provider: shareddata.ArrayDoubles, ids: []string{"array-double-empty"}},
				{provider: shareddata.ArrayInt32s, ids: []string{"array-int32-empty"}},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "null"}},
			},
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
		"EmptyArray": {
			update:     bson.D{{"$unset", bson.A{}}},
			resultType: emptyResult,
		},
		"DotNotation": {
			update:           bson.D{{"$unset", bson.D{{"v.foo", ""}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/442", // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/445
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/442", // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/445
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationNonExistentPath": {
			update:     bson.D{{"$unset", bson.D{{"not.existent.path", ""}}}},
			resultType: emptyResult,
		},
		"DotArrayField": {
			update: bson.D{{"$unset", bson.D{{"v.array.0", ""}}}},
			skip:   "https://github.com/FerretDB/FerretDB/issues/1242",
		},
		"DotNotationArrNonExistentPath": {
			update:     bson.D{{"$unset", bson.D{{"non.0.existent", int32(1)}}}},
			resultType: emptyResult,
		},
		"DotNotationMissingField": {
			update:           bson.D{{"$unset", bson.D{{"v..", ""}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
		},
		"DotNotationNegativeIndex": {
			update:           bson.D{{"$unset", bson.D{{"v.-1.bar", ""}}}},
			resultType:       emptyResult,
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/442",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/442",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationIndexOutOfArray": {
			update:           bson.D{{"$unset", bson.D{{"v.100.bar", ""}}}},
			resultType:       emptyResult,
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/445",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/445",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{"array-empty"}},
				{provider: shareddata.ArrayStrings, ids: []string{"array-string-empty"}},
				{provider: shareddata.ArrayDoubles, ids: []string{"array-double-empty"}},
				{provider: shareddata.ArrayInt32s, ids: []string{"array-int32-empty"}},
				{provider: shareddata.Mixed, ids: []string{"array-empty"}},
			},
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatSet(t *testing.T) {
	t.Parallel()

	int32sProvider := []shareddata.Provider{shareddata.Int32s}
	int64sProvider := []shareddata.Provider{shareddata.Int64s}
	doublesProvider := []shareddata.Provider{shareddata.Doubles}

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
			update: bson.D{{"$set", bson.D{{"foo", int32(1)}}}},
		},
		"Double": {
			update: bson.D{{"$set", bson.D{{"v", float64(1)}}}},
		},
		"Null": {
			update: bson.D{{"$set", bson.D{{"v", nil}}}},
		},
		"Int32": {
			update: bson.D{{"$set", bson.D{{"v", int32(1)}}}},
		},
		"SetTwoFields": {
			update: bson.D{{"$set", bson.D{{"foo", int32(12)}, {"v", nil}}}},
		},
		"Int32Type": {
			update: bson.D{{"$set", bson.D{{"v", int32(42)}}}},
		},
		"Int32TypeOnly": {
			update:    bson.D{{"$set", bson.D{{"v", int32(42)}}}},
			providers: int32sProvider,
		},
		"Int64Type": {
			update: bson.D{{"$set", bson.D{{"v", int64(42)}}}},
		},
		"Int64TypeOnly": {
			update:    bson.D{{"$set", bson.D{{"v", int64(42)}}}},
			providers: int64sProvider,
		},
		"DoubleType": {
			update: bson.D{{"$set", bson.D{{"v", 42.0}}}},
		},
		"DoubleTypeOnly": {
			update:    bson.D{{"$set", bson.D{{"v", 42.0}}}},
			providers: doublesProvider,
		},
		"DocSameNumberType": {
			update: bson.D{{"$set", bson.D{{"v", bson.D{{"foo", int32(42)}}}}}},
		},
		"DocDifferentNumberType": {
			update:           bson.D{{"$set", bson.D{{"v", bson.D{{"foo", int64(42)}}}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/501",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{
					provider: shareddata.DocumentsDocuments,
					ids:      []string{fmt.Sprint(primitive.ObjectID{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01})},
				},
				{provider: shareddata.Composites, ids: []string{"document"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"document"}},
			},
		},

		"DocumentField": {
			update: bson.D{{"$set", bson.D{{"foo", int32(42)}, {"bar", "baz"}}}},
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
			update: bson.D{{"$set", bson.D{{"v", must.NotFail(primitive.ObjectIDFromHex("000102030405060708091011"))}}}},
		},
		"ObjectIDEmpty": {
			update: bson.D{{"$set", bson.D{{"v", primitive.NilObjectID}}}},
		},
		"Bool": {
			update: bson.D{{"$set", bson.D{{"v", true}}}},
		},
		"Datetime": {
			update: bson.D{{"$set", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000000, time.UTC))}}}},
		},
		"DatetimeNanoSecDiff": {
			update: bson.D{{"$set", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 10, 18, 42, 123000001, time.UTC))}}}},
		},
		"DatetimeEpoch": {
			update: bson.D{{"$set", bson.D{{"v", primitive.NewDateTimeFromTime(time.Unix(0, 0))}}}},
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
			update:           bson.D{{"$set", bson.D{{"v.foo", int32(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/479",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-empty", "array-documents", "array-null", "array-composite", "array-numbers-asc",
					"array-numbers-desc", "array-strings-desc", "array-three", "array-three-reverse", "array-two",
				}},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null"}},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
			},
		},
		"DocumentFieldNotExist": {
			update: bson.D{{"$set", bson.D{{"foo.bar", int32(1)}}}},
		},
		"ArrayFieldExist": {
			update:           bson.D{{"$set", bson.D{{"v.array.0", int32(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/479",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-empty", "array-documents", "array-null", "array-composite", "array-numbers-asc",
					"array-numbers-desc", "array-strings-desc", "array-three", "array-three-reverse", "array-two",
				}},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null"}},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
			},
		},
		"ArrayFieldNotExist": {
			update: bson.D{{"$set", bson.D{{"foo.0.baz", int32(1)}}}},
		},
		"DocArrFieldNotExists_0": {
			update: bson.D{{"$set", bson.D{{"v.0.foo", int32(1)}}}},
		},
		"DocArrFieldNotExists_1": {
			update: bson.D{{"$set", bson.D{{"v.1.foo", int32(1)}}}},
		},
		"DocArrFieldNotExists_2": {
			update: bson.D{{"$set", bson.D{{"v.2", int32(1)}}}},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$set", bson.D{{"v..", int32(1)}}}},
			resultType: emptyResult,
		},
		"DotNotationNegativeIndex": {
			update:           bson.D{{"$set", bson.D{{"v.-1.bar", int32(1)}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/479",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null"}},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-documents", "array-null", "array-empty", "array-composite", "array-numbers-asc",
					"array-numbers-desc", "array-strings-desc", "array-three", "array-three-reverse", "array-two",
				}},
			},
		},
		"DotNotationIndexOutOfArray": {
			update: bson.D{{"$set", bson.D{{"v.100.bar", int32(1)}}}},
		},
		"ID": {
			update:     bson.D{{"$set", bson.D{{"_id", "non-existent"}}}},
			resultType: emptyResult,
		},
		"SetID": {
			update: bson.D{{"$set", bson.D{{"_id", "int32"}, {"v", int32(2)}}}},
		},
		"ConflictKey": {
			update: bson.D{
				{"$set", bson.D{{"v", "val"}}},
				{"$min", bson.D{{"v.foo", "val"}}},
			},
			resultType: emptyResult,
		},
		"ConflictKeyPrefix": {
			update: bson.D{
				{"$set", bson.D{{"v.foo", "val"}}},
				{"$min", bson.D{{"v", "val"}}},
			},
			resultType: emptyResult,
		},
		"ExistingID": {
			filter:     bson.D{{"_id", "int32"}},
			update:     bson.D{{"$set", bson.D{{"_id", "int32-1"}, {"v", int32(2)}}}},
			resultType: emptyResult,
		},
		"SameID": {
			filter: bson.D{{"_id", "int32"}},
			update: bson.D{{"$set", bson.D{{"_id", "int32"}, {"v", int32(2)}}}},
		},
		"DifferentID": {
			filter:     bson.D{{"_id", "int32"}},
			update:     bson.D{{"$set", bson.D{{"_id", "another-id"}, {"v", int32(2)}}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatSetMulti(t *testing.T) {
	t.Parallel()

	testCases := map[string]testUpdateManyCompatTestCase{
		"QueryOperatorExists": {
			filter:     bson.D{{"v", bson.D{{"$lt", 3}}}},
			update:     bson.D{{"$set", bson.D{{"new", "val"}}}},
			updateOpts: options.Update().SetUpsert(true),
			// only use providers contain filter match, no match results in
			// upsert with generated ID which is tested in integration test
			providers: []shareddata.Provider{shareddata.Scalars, shareddata.Int32s, shareddata.Doubles},
		},
		"QueryOperatorUpsertFalse": {
			filter:     bson.D{{"v", int32(4080)}},
			update:     bson.D{{"$set", bson.D{{"new", "val"}}}},
			updateOpts: options.Update().SetUpsert(false),
		},
		"QueryOperatorModified": {
			filter:     bson.D{{"v", bson.D{{"$eq", 4080}}}},
			update:     bson.D{{"$set", bson.D{{"new", "val"}}}},
			updateOpts: options.Update().SetUpsert(false),
		},
		"QueryOperatorEmptySet": {
			filter:     bson.D{{"v", bson.D{{"$eq", 4080}}}},
			update:     bson.D{{"$set", bson.D{}}},
			updateOpts: options.Update().SetUpsert(false),
			resultType: emptyResult,
		},
	}

	testUpdateManyCompat(t, testCases)
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
		},
		"EmptyArray": {
			update: bson.D{{"$set", bson.D{{"v", bson.A{}}}}},
		},
		"ArrayStringsDesc": {
			update: bson.D{{"$set", bson.D{{"v", bson.A{"c", "b", "a"}}}}},
		},
		"ArrayChangedNumberType": {
			update:           bson.D{{"$set", bson.D{{"v", bson.A{int64(42), int64(43), 45.5}}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/501",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{"array-numbers-asc"}},
			},
		},
		"ArrayUnchangedNumberType": {
			update: bson.D{{"$set", bson.D{{"v", bson.A{int32(42), int64(43), 45.5}}}}},
		},
		"DocSameNumberType": {
			update: bson.D{{"$set", bson.D{{"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int32(42), "foo", nil}}}}}}},
		},
		"DocDifferentNumberType": {
			update:           bson.D{{"$set", bson.D{{"v", bson.D{{"foo", int32(42)}, {"42", "foo"}, {"array", bson.A{int64(42), "foo", nil}}}}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/501",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{"document-composite"}},
			},
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
		"DuplicateKeys": {
			update:           bson.D{{"$setOnInsert", bson.D{{"v", 1}, {"v", 2}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1041",
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
		"DotNotationMissingField": {
			update:           bson.D{{"$setOnInsert", bson.D{{"v..", int32(1)}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1041",
		},
		"DotNotationNegativeIdx": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v.-1.bar", int32(1)}}}},
			resultType: emptyResult,
		},
		"DotNotatIndexOutOfArr": {
			update:     bson.D{{"$setOnInsert", bson.D{{"v.100.bar", int32(1)}}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatSetOnInsertComplex(t *testing.T) {
	t.Parallel()

	testCases := map[string]testUpdateManyCompatTestCase{
		"IDExists": {
			filter:     bson.D{{"_id", "int32"}},
			update:     bson.D{{"$setOnInsert", bson.D{{"new", "val"}}}},
			updateOpts: options.Update().SetUpsert(true),
			providers:  []shareddata.Provider{shareddata.Int32s},
			resultType: emptyResult,
		},
		"IDNotExists": {
			filter:     bson.D{{"_id", "non-existent"}},
			update:     bson.D{{"$setOnInsert", bson.D{{"new", "val"}}}},
			updateOpts: options.Update().SetUpsert(true),
		},
		"UpsertFalse": {
			filter:     bson.D{{"_id", "non-existent"}},
			update:     bson.D{{"$setOnInsert", bson.D{{"new", "val"}}}},
			updateOpts: options.Update().SetUpsert(false),
			resultType: emptyResult,
		},
		"SetWithSetOnInsert": {
			filter: bson.D{{"_id", "non-existent"}},
			update: bson.D{
				{"$set", bson.D{{"new", "val"}}},
				{"$setOnInsert", bson.D{{"v", int32(42)}}},
			},
			updateOpts: options.Update().SetUpsert(true),
		},
		"ApplySetSkipSOI": {
			filter: bson.D{{"_id", "int32"}},
			update: bson.D{
				{"$set", bson.D{{"new", "val"}}},
				{"$setOnInsert", bson.D{{"v", int32(43)}}},
			},
			updateOpts: options.Update().SetUpsert(true),
			providers:  []shareddata.Provider{shareddata.Int32s},
		},
	}

	testUpdateManyCompat(t, testCases)
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
		"UpsertQueryOperatorEq": {
			filter:     bson.D{{"_id", bson.D{{"$eq", "non-existent"}}}},
			update:     bson.D{{"$set", bson.D{{"new", "val"}}}},
			updateOpts: options.Update().SetUpsert(true),
		},
		"UpsertQueryOperatorMixed": {
			filter: bson.D{
				{"_id", bson.D{{"$eq", "non-existent"}}},
				{"v", bson.D{{"$lt", 43}}},
				{"non_existent", int32(0)},
			},
			update:     bson.D{{"$set", bson.D{{"new", "val"}}}},
			updateOpts: options.Update().SetUpsert(true),
		},
		"UpsertQueryObject": {
			filter:     bson.D{{"_id", "non-existent"}, {"v", bson.D{{"k1", "v1"}}}},
			update:     bson.D{{"$set", bson.D{{"new", "val"}}}},
			updateOpts: options.Update().SetUpsert(true),
		},
		"UpsertQueryObjectNested": {
			filter:     bson.D{{"_id", "non-existent"}, {"v", bson.D{{"k1", "v1"}, {"k2", bson.D{{"k21", "v21"}}}}}},
			update:     bson.D{{"$set", bson.D{{"new", "val"}}}},
			updateOpts: options.Update().SetUpsert(true),
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatMul(t *testing.T) {
	t.Parallel()

	providers := shareddata.AllProviders().
		// OverflowVergeDoubles and Scalars contain numbers that produces +INF on compat,
		// validation error on target upon $mul operation.
		Remove(shareddata.OverflowVergeDoubles, shareddata.Scalars)

	testCases := map[string]updateCompatTestCase{
		"Int32": {
			update:           bson.D{{"$mul", bson.D{{"v", int32(42)}}}},
			providers:        providers,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/434",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Int64s, ids: []string{"int64-min"}},
			},
		},
		"Int32Negative": {
			update:           bson.D{{"$mul", bson.D{{"v", int32(-42)}}}},
			providers:        providers,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/434",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Int64s, ids: []string{"int64-min"}},
			},
		},
		"Int32Min": {
			update:           bson.D{{"$mul", bson.D{{"v", math.MinInt32}}}},
			providers:        providers,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/434",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Int64s, ids: []string{"int64-big", "int64-min", "int64-neg-big", "int64-prec-max-plus", "int64-prec-min-minus"}},
			},
		},
		"Int32Max": {
			update:    bson.D{{"$mul", bson.D{{"v", math.MaxInt32}}}},
			providers: providers,
		},
		"Int64": {
			update:           bson.D{{"$mul", bson.D{{"v", int64(42)}}}},
			providers:        providers,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/434",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Int64s, ids: []string{"int64-min"}},
			},
		},
		"Int64Negative": {
			update:           bson.D{{"$mul", bson.D{{"v", int32(-42)}}}},
			providers:        providers,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/434",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Int64s, ids: []string{"int64-min"}},
			},
		},
		"Int64Min": {
			update:           bson.D{{"$mul", bson.D{{"v", math.MinInt64}}}},
			providers:        providers,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/434",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Int32s, ids: []string{"int32", "int32-1", "int32-2", "int32-3", "int32-min"}},
				{provider: shareddata.Int64s, ids: []string{
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-min", "int64-neg-big", "int64-prec-max-minus", "int64-prec-max-plus",
					"int64-prec-min-minus", "int64-prec-min-plus",
				}},
			},
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
			update:           bson.D{{"$mul", bson.D{{"v", int32(-42)}}}},
			providers:        providers,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/434",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Int64s, ids: []string{"int64-min"}},
			},
		},
		"DoubleSmallestNonZero": {
			update:    bson.D{{"$mul", bson.D{{"v", math.SmallestNonzeroFloat64}}}},
			providers: providers,
		},
		"DoubleBig": {
			update:    bson.D{{"$mul", bson.D{{"v", float64(1 << 61)}}}}, // TODO https://github.com/FerretDB/FerretDB/issues/3626
			providers: providers,
		},
		"Empty": {
			update:     bson.D{{"$mul", bson.D{}}},
			resultType: emptyResult,
		},
		"Null": {
			update:     bson.D{{"$mul", bson.D{{"v", nil}}}},
			resultType: emptyResult,
		},
		"String": {
			update:     bson.D{{"$mul", bson.D{{"v", "string"}}}},
			resultType: emptyResult,
		},
		"MissingField": {
			update:     bson.D{{"$mul", "invalid"}},
			resultType: emptyResult,
		},
		"StringFieldNotExist": {
			update:     bson.D{{"$mul", bson.D{{"foo.bar", "bad value"}}}},
			resultType: emptyResult,
		},
		"FieldNotExist": {
			update: bson.D{{"$mul", bson.D{{"foo", int32(45)}}}},
		},
		"TwoFields": {
			update: bson.D{{"$mul", bson.D{{"foo", int32(12)}, {"v", int32(1)}}}},
		},
		"DuplicateKeys": {
			update:     bson.D{{"$mul", bson.D{{"v", int32(42)}, {"v", int32(43)}}}},
			resultType: emptyResult,
		},
		"InvalidLastField": {
			update:     bson.D{{"$mul", bson.D{{"foo", int32(12)}, {"v", "string"}}}},
			resultType: emptyResult,
		},
		"MultipleOperator": {
			update: bson.D{
				{"$set", bson.D{{"foo", int32(43)}}},
				{"$mul", bson.D{{"v", int32(42)}}},
			},
			providers:        providers,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/434",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Int64s, ids: []string{"int64-min"}},
			},
		},
		"ConflictPop": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$pop", bson.D{{"v", -1}}},
			},
			providers:  providers,
			resultType: emptyResult,
		},
		"ConflictSet": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$set", bson.D{{"v", int32(43)}}},
			},
			providers:  providers,
			resultType: emptyResult,
		},
		"ConflictInc": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$inc", bson.D{{"v", int32(43)}}},
			},
			providers:  providers,
			resultType: emptyResult,
		},
		"ConflictMin": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$min", bson.D{{"v", int32(30)}}},
			},
			providers:  providers,
			resultType: emptyResult,
		},
		"ConflictMax": {
			update: bson.D{
				{"$max", bson.D{{"v", int32(30)}}},
				{"$mul", bson.D{{"v", int32(42)}}},
			},
			providers:  providers,
			resultType: emptyResult,
		},
		"ConflictSetOnInsert": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$setOnInsert", 43.13},
			},
			providers:  providers,
			resultType: emptyResult,
		},
		"ConflictUnset": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$unset", bson.D{{"v", ""}}},
			},
			providers:  providers,
			resultType: emptyResult,
		},
		"ConflictCurrentDate": {
			update: bson.D{
				{"$mul", bson.D{{"v", int32(42)}}},
				{"$currentDate", bson.D{{"v", bson.D{{"$type", "date"}}}}},
			},
			providers:  providers,
			resultType: emptyResult,
		},
		"DotNotation": {
			update:           bson.D{{"$mul", bson.D{{"v.foo", int32(45)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationNotExistentPath": {
			update: bson.D{{"$mul", bson.D{{"not.existent.path", int32(45)}}}},
		},
		"DotNotationArrayFieldExist": {
			update:           bson.D{{"$mul", bson.D{{"v.array.0", int32(45)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationArrayFieldNotExist": {
			update:           bson.D{{"$mul", bson.D{{"v.array.0.foo", int32(45)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two", "document-composite", "document-composite-reverse",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$mul", bson.D{{"v..", int32(45)}}}},
			resultType: emptyResult,
		},
		"DotNotatIndexOverArrayLen": {
			update:           bson.D{{"$mul", bson.D{{"v.100.bar", int32(45)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.Mixed, ids: []string{"null"}},
			},
		},
		"DotNotationFieldNumericName": {
			update:           bson.D{{"$mul", bson.D{{"v.array.42", int32(42)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationNegativeIndex": {
			update:           bson.D{{"$mul", bson.D{{"v.array.-1", int32(42)}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch", "datetime-year-max", "datetime-year-min", "double",
					"double-1", "double-2", "double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3", "int32-max", "int32-min", "int32-zero",
					"int64", "int64-1", "int64-2", "int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid",
					"objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two", "document-composite", "document-composite-reverse",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateFieldCompatBit(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"And": {
			update: bson.D{{"$bit", bson.D{{"v", bson.D{{"and", 1}}}}}},
		},
		"Or": {
			update: bson.D{{"$bit", bson.D{{"v", bson.D{{"or", 1}}}}}},
		},
		"Xor": {
			update: bson.D{{"$bit", bson.D{{"v", bson.D{{"xor", 1}}}}}},
		},
		"Int32": {
			update: bson.D{
				{"$bit", bson.D{
					{"v", bson.D{{"and", int32(1)}}},
				}},
			},
		},
		"Int32Negative": {
			update: bson.D{{"$bit", bson.D{{"v", bson.D{{"and", int32(-1)}}}}}},
		},
		"Int32Min": {
			update: bson.D{{"$bit", bson.D{{"v", bson.D{{"or", math.MinInt32}}}}}},
		},
		"Int32Max": {
			update: bson.D{{"$bit", bson.D{{"v", bson.D{{"xor", math.MaxInt32}}}}}},
		},
		"Int64": {
			update: bson.D{{"$bit", bson.D{
				{"v", bson.D{{"or", int64(11)}}},
			}}},
		},
		"Int64Min": {
			update: bson.D{{"$bit", bson.D{{"v", bson.D{{"xor", math.MinInt64}}}}}},
		},
		"Int64Max": {
			update: bson.D{{"$bit", bson.D{{"v", bson.D{{"and", math.MaxInt64}}}}}},
		},
		"Int64MaxUnderflow": {
			update: bson.D{{"$bit", bson.D{{"v", bson.D{{"or", -math.MaxInt64}}}}}},
		},
		"Int64MaxOverflow": {
			update: bson.D{{"$bit", bson.D{{"v", bson.D{{"or", math.MaxInt64}}}}}},
		},
		"Double": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"and", float64(1)}}}}}},
			resultType: emptyResult,
		},
		"String": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"and", "string"}}}}}},
			resultType: emptyResult,
		},
		"Binary": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"and", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}}}},
			resultType: emptyResult,
		},
		"ObjectID": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"or", primitive.ObjectID{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11}}}}}}},
			resultType: emptyResult,
		},
		"Bool": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"or", true}}}}}},
			resultType: emptyResult,
		},
		"DateTime": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"or", primitive.NewDateTimeFromTime(time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC))}}}}}},
			resultType: emptyResult,
		},
		"Nil": {
			update:     bson.D{{"$bit", bson.D{{"and", nil}}}},
			resultType: emptyResult,
		},
		"Regex": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"xor", primitive.Regex{Pattern: "foo", Options: "i"}}}}}}},
			resultType: emptyResult,
		},
		"Timestamp": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"xor", primitive.Timestamp{T: 42, I: 13}}}}}}},
			resultType: emptyResult,
		},
		"Object": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"xor", bson.D{{"foo", int32(42)}}}}}}}},
			resultType: emptyResult,
		},
		"Array": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"xor", bson.A{int32(42)}}}}}}},
			resultType: emptyResult,
		},
		"NonExistent": {
			update: bson.D{{"$bit", bson.D{{"non-existent", bson.D{{"xor", int32(1)}}}}}},
		},
		"DotNotation": {
			update:           bson.D{{"$bit", bson.D{{"v.foo", bson.D{{"xor", int32(1)}}}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch",
					"datetime-year-max", "datetime-year-min", "double", "double-1", "double-2", "double-3",
					"double-4", "double-5", "double-big", "double-max", "double-max-overflow",
					"double-min-overflow", "double-smallest", "double-whole", "double-zero", "int32", "int32-1",
					"int32-2", "int32-3", "int32-max", "int32-min", "int32-zero", "int64", "int64-1", "int64-2",
					"int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null",
					"objectid", "objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty",
					"string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc",
					"array-strings-desc", "array-three", "array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationArray": {
			update:           bson.D{{"$bit", bson.D{{"v.0", bson.D{{"xor", int32(1)}}}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch",
					"datetime-year-max", "datetime-year-min", "double", "double-1", "double-2", "double-3",
					"double-4", "double-5", "double-big", "double-max", "double-max-overflow",
					"double-min-overflow", "double-smallest", "double-whole", "double-zero", "int32", "int32-1",
					"int32-2", "int32-3", "int32-max", "int32-min", "int32-zero", "int64", "int64-1", "int64-2",
					"int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null",
					"objectid", "objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty",
					"string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.Mixed, ids: []string{"null"}},
			},
		},
		"DotNotationMissingField": {
			update:     bson.D{{"$bit", bson.D{{"v..", int32(1)}}}},
			resultType: emptyResult,
		},
		"DotNotationNegativeIndex": {
			update:           bson.D{{"$bit", bson.D{{"v.-1", bson.D{{"or", int32(10)}}}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch",
					"datetime-year-max", "datetime-year-min", "double", "double-1", "double-2", "double-3",
					"double-4", "double-5", "double-big", "double-max", "double-max-overflow",
					"double-min-overflow", "double-smallest", "double-whole", "double-zero", "int32", "int32-1",
					"int32-2", "int32-3", "int32-max", "int32-min", "int32-zero", "int64", "int64-1", "int64-2",
					"int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null",
					"objectid", "objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty",
					"string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc",
					"array-strings-desc", "array-three", "array-three-reverse", "array-two",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotationArrayFieldNotExist": {
			update:           bson.D{{"$bit", bson.D{{"v.array.0.foo", bson.D{{"xor", int32(11)}}}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch",
					"datetime-year-max", "datetime-year-min", "double", "double-1", "double-2", "double-3",
					"double-4", "double-5", "double-big", "double-max", "double-max-overflow",
					"double-min-overflow", "double-smallest", "double-whole", "double-zero", "int32", "int32-1",
					"int32-2", "int32-3", "int32-max", "int32-min", "int32-zero", "int64", "int64-1", "int64-2",
					"int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null",
					"objectid", "objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty",
					"string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"array", "array-composite", "array-documents", "array-empty", "array-null", "array-numbers-asc",
					"array-strings-desc", "array-three", "array-three-reverse", "array-two", "document-composite",
					"document-composite-reverse",
				}},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayDocuments},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null", "null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
			},
		},
		"DotNotAtIndexOverArrayLen": {
			update:           bson.D{{"$bit", bson.D{{"v.100.foo", bson.D{{"and", int32(11)}}}}}},
			skip:             "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/429",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime", "datetime-epoch",
					"datetime-year-max", "datetime-year-min", "double", "double-1", "double-2", "double-3",
					"double-4", "double-5", "double-big", "double-max", "double-max-overflow",
					"double-min-overflow", "double-smallest", "double-whole", "double-zero", "int32", "int32-1",
					"int32-2", "int32-3", "int32-max", "int32-min", "int32-zero", "int64", "int64-1", "int64-2",
					"int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null",
					"objectid", "objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty",
					"string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.Doubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Bools},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Nulls},
				{provider: shareddata.Regexes},
				{provider: shareddata.Int32s},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Int64s},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.DocumentsDoubles, ids: []string{"document-double-null"}},
				{provider: shareddata.DocumentsStrings, ids: []string{"document-string-nil"}},
				{provider: shareddata.Mixed, ids: []string{"null"}},
			},
		},
		"EmptyBitwiseOperation": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{}}}}},
			resultType: emptyResult,
		},
		"InvalidBitwiseOperation": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"not", int32(10)}}}}}},
			resultType: emptyResult,
		},
		"InvalidBitwiseOperand": {
			update:     bson.D{{"$bit", bson.D{{"v", bson.D{{"and", bson.A{}}}}}}},
			resultType: emptyResult,
		},
		"EmptyUpdateOperand": {
			update:     bson.D{{"$bit", bson.D{}}},
			resultType: emptyResult,
		},
		"DuplicateKeys": {
			update: bson.D{
				{"$bit", bson.D{
					{"v", bson.D{{"and", int32(1)}}},
					{"v", bson.D{{"or", int32(1)}}},
				}},
			},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}
