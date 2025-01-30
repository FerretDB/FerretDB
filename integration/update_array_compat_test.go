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
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func TestUpdateArrayCompatPop(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"DuplicateKeys": {
			update:     bson.D{{"$pop", bson.D{{"v", 1}, {"v", 1}}}},
			resultType: emptyResult,
		},
		"Pop": {
			update:           bson.D{{"$pop", bson.D{{"v", 1}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/314",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Binaries},
				{provider: shareddata.Bools},
				{provider: shareddata.Composites, ids: []string{
					"document", "document-composite",
					"document-composite-numerical-field-name", "document-composite-reverse", "document-empty", "document-null",
				}},
				{provider: shareddata.Decimal128s},
				{provider: shareddata.Doubles},
				{provider: shareddata.Int32s},
				{provider: shareddata.Int64s},
				{provider: shareddata.Nulls},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime",
					"datetime-epoch", "datetime-year-max", "datetime-year-min", "double", "double-1", "double-2",
					"double-3", "double-4", "double-5", "double-big", "double-max", "double-max-overflow",
					"double-min-overflow", "double-smallest", "double-whole", "double-zero",
					"decimal128", "decimal128-int", "decimal128-int-zero", "decimal128-zero", "decimal128-double", "decimal128-whole", "int32", "int32-1",
					"int32-2", "int32-3", "int32-max", "int32-min", "int32-zero", "int64", "int64-1", "int64-2",
					"int64-3", "int64-big", "int64-double-big", "int64-max", "int64-min", "int64-zero", "null",
					"objectid", "objectid-empty", "regex", "regex-empty", "string", "string-double", "string-empty",
					"string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Mixed, ids: []string{"null"}},
				{provider: shareddata.Regexes},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"document"}},
				{provider: shareddata.DocumentsDocuments},
				{provider: shareddata.DocumentsDeeplyNested},
				{provider: shareddata.DocumentsDoubles},
				{provider: shareddata.DocumentsStrings},
				{provider: shareddata.PostgresEdgeCases},
			},
		},
		"PopFirst": {
			update:           bson.D{{"$pop", bson.D{{"v", -1}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/314",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Binaries},
				{provider: shareddata.Bools},
				{provider: shareddata.Composites, ids: []string{
					"document", "document-composite",
					"document-composite-numerical-field-name", "document-composite-reverse", "document-empty",
					"document-null",
				}},
				{provider: shareddata.Decimal128s},
				{provider: shareddata.Doubles},
				{provider: shareddata.Int32s},
				{provider: shareddata.Int64s},
				{provider: shareddata.Nulls},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.Scalars, ids: []string{
					"binary", "binary-empty", "bool-false", "bool-true", "datetime",
					"datetime-epoch", "datetime-year-max", "datetime-year-min", "double", "double-1", "double-2", "double-3",
					"decimal128", "decimal128-int", "decimal128-int-zero", "decimal128-zero", "decimal128-double", "decimal128-whole",
					"double-4", "double-5", "double-big", "double-max", "double-max-overflow", "double-min-overflow",
					"double-smallest", "double-whole", "double-zero", "int32", "int32-1", "int32-2", "int32-3",
					"int32-max", "int32-min", "int32-zero", "int64", "int64-1", "int64-2", "int64-3", "int64-big",
					"int64-double-big", "int64-max", "int64-min", "int64-zero", "null", "objectid", "objectid-empty",
					"regex", "regex-empty", "string", "string-double", "string-empty", "string-whole", "timestamp", "timestamp-i",
				}},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.Strings},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Timestamps},
				{provider: shareddata.Mixed, ids: []string{"null"}},
				{provider: shareddata.Regexes},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"document"}},
				{provider: shareddata.DocumentsDocuments},
				{provider: shareddata.DocumentsDeeplyNested},
				{provider: shareddata.DocumentsDoubles},
				{provider: shareddata.DocumentsStrings},
				{provider: shareddata.PostgresEdgeCases},
			},
		},
		"NonExistentField": {
			update:     bson.D{{"$pop", bson.D{{"non-existent-field", 1}}}},
			resultType: emptyResult,
		},
		"DotNotation": {
			filter: bson.D{{"_id", "array-documents-nested"}},
			update: bson.D{{"$pop", bson.D{{"v.0.foo", 1}}}},
		},
		"DotNotationPopFirst": {
			filter: bson.D{{"_id", "array-documents-nested"}},
			update: bson.D{{"$pop", bson.D{{"v.0.foo", -1}}}},
		},
		"DotNotationNonArray": {
			filter:           bson.D{{"_id", "array-documents-nested"}},
			update:           bson.D{{"$pop", bson.D{{"v.0.foo.0.bar", 1}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/314",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.ArrayDocuments},
			},
		},
		"DotNotationNonExistentPath": {
			update:     bson.D{{"$pop", bson.D{{"non.existent.path", 1}}}},
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
		"DotNotationObjectInArray": {
			update:           bson.D{{"$pop", bson.D{{"v.array.foo.array", 1}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/413",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-empty", "array-null", "array-composite",
					"array-documents", "array-numbers-asc", "array-numbers-desc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two", "document-composite", "document-composite-reverse",
				}},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields"}},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayDocuments},
			},
		},
		"DotNotationObject": {
			update:           bson.D{{"$pop", bson.D{{"v.foo", 1}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/413",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Composites, ids: []string{
					"array", "array-empty", "array-null", "array-composite",
					"array-documents", "array-numbers-asc", "array-numbers-desc", "array-strings-desc", "array-three",
					"array-three-reverse", "array-two", "document", "document-composite", "document-composite-reverse",
					"document-composite-numerical-field-name", "document-null",
				}},
				{provider: shareddata.Mixed, ids: []string{"array-empty", "array-null"}},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"array-documents", "array-documents-two-fields", "document"}},
				{
					provider: shareddata.DocumentsDocuments,
					ids:      []string{fmt.Sprint(primitive.ObjectID{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01})},
				},
				{provider: shareddata.ArrayStrings},
				{provider: shareddata.ArrayDoubles},
				{provider: shareddata.ArrayRegexes},
				{provider: shareddata.ArrayInt32s},
				{provider: shareddata.ArrayDocuments},
			},
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateArrayCompatPush(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"DuplicateKeys": {
			update:     bson.D{{"$push", bson.D{{"v", "foo"}, {"v", "bar"}}}},
			resultType: emptyResult, // conflict because of duplicate keys "v" set in $push
		},
		"String": {
			update: bson.D{{"$push", bson.D{{"v", "foo"}}}},
		},
		"Int32": {
			update: bson.D{{"$push", bson.D{{"v", int32(42)}}}},
		},
		"NonExistentField": {
			update: bson.D{{"$push", bson.D{{"non-existent-field", int32(42)}}}},
		},
		"DotNotation": {
			filter: bson.D{{"_id", "array-documents-nested"}},
			update: bson.D{{"$push", bson.D{{"v.0.foo", bson.D{{"bar", "zoo"}}}}}},
		},
		"DotNotationNonArray": {
			filter:     bson.D{{"_id", "array-documents-nested"}},
			update:     bson.D{{"$push", bson.D{{"v.0.foo.0.bar", "boo"}}}},
			resultType: emptyResult, // attempt to push to non-array
		},
		"DotNotationNonExistentPath": {
			update: bson.D{{"$push", bson.D{{"non.existent.path", int32(42)}}}},
		},
		"TwoElements": {
			update: bson.D{{"$push", bson.D{{"non.existent.path", int32(42)}, {"v", int32(42)}}}},
		},
	}

	testUpdateCompat(t, testCases)
}

// TestUpdateArrayCompatAddToSet tests the $addToSet update operator.
// Test case "String" will cover the case where the value is already in set when ran against "array-two" document.
func TestUpdateArrayCompatAddToSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"DuplicateKeys": {
			update:     bson.D{{"$addToSet", bson.D{{"v", int32(1)}, {"v", int32(1)}}}},
			resultType: emptyResult,
		},
		"String": {
			update: bson.D{{"$addToSet", bson.D{{"v", "foo"}}}},
		},
		"Document": {
			update: bson.D{{"$addToSet", bson.D{{"v", bson.D{{"foo", "bar"}}}}}},
		},
		"Int32": {
			update: bson.D{{"$addToSet", bson.D{{"v", int32(42)}}}},
		},
		"Int64": {
			update: bson.D{{"$addToSet", bson.D{{"v", int64(42)}}}},
		},
		"Float64": {
			update: bson.D{{"$addToSet", bson.D{{"v", float64(42)}}}},
		},
		"NonExistentField": {
			update: bson.D{{"$addToSet", bson.D{{"non-existent-field", int32(42)}}}},
		},
		"DotNotation": {
			filter: bson.D{{"_id", "array-documents-nested"}},
			update: bson.D{{"$addToSet", bson.D{{"v.0.foo", bson.D{{"bar", "zoo"}}}}}},
		},
		"DotNotationNonArray": {
			filter:     bson.D{{"_id", "array-documents-nested"}},
			update:     bson.D{{"$addToSet", bson.D{{"v.0.foo.0.bar", int32(1)}}}},
			resultType: emptyResult,
		},
		"DotNotationNonExistentPath": {
			update: bson.D{{"$addToSet", bson.D{{"non.existent.path", int32(1)}}}},
		},
		"EmptyValue": {
			update:     bson.D{{"$addToSet", bson.D{}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

// TestUpdateArrayCompatPullAll tests the $pullAll update operator.
func TestUpdateArrayCompatPullAll(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"DuplicateKeys": {
			update:     bson.D{{"$pullAll", bson.D{{"v", bson.A{int32(1)}}, {"v", bson.A{int32(1)}}}}},
			resultType: emptyResult,
		},
		"StringValue": {
			update:     bson.D{{"$pullAll", bson.D{{"v", "foo"}}}},
			resultType: emptyResult,
		},
		"String": {
			update: bson.D{{"$pullAll", bson.D{{"v", bson.A{"foo"}}}}},
		},
		"Document": {
			update: bson.D{{"$pullAll", bson.D{{"v", bson.A{bson.D{{"field", int32(42)}}}}}}},
		},
		"Int32": {
			update: bson.D{{"$pullAll", bson.D{{"v", bson.A{int32(42)}}}}},
		},
		"Int32-Six-Elements": {
			update: bson.D{{"$pullAll", bson.D{{"v", bson.A{int32(42), int32(43)}}}}},
		},
		"Int64": {
			update: bson.D{{"$pullAll", bson.D{{"v", bson.A{int64(42)}}}}},
		},
		"Float64": {
			update: bson.D{{"$pullAll", bson.D{{"v", bson.A{float64(42)}}}}},
		},
		"NonExistentField": {
			update:     bson.D{{"$pullAll", bson.D{{"non-existent-field", bson.A{int32(42)}}}}},
			resultType: emptyResult,
		},
		"NotSuitableField": {
			filter:     bson.D{{"_id", "int32"}},
			update:     bson.D{{"$pullAll", bson.D{{"v.foo", bson.A{int32(42)}}}}},
			resultType: emptyResult,
		},
		"DotNotation": {
			filter: bson.D{{"_id", "array-documents-nested"}},
			update: bson.D{{"$pullAll", bson.D{{"v.0.foo", bson.A{bson.D{{"bar", "hello"}}}}}}},
		},
		"DotNotationNonArray": {
			filter:     bson.D{{"_id", "array-documents-nested"}},
			update:     bson.D{{"$pullAll", bson.D{{"v.0.foo.0.bar", bson.A{int32(42)}}}}},
			resultType: emptyResult,
		},
		"DotNotationNonExistentPath": {
			update:     bson.D{{"$pullAll", bson.D{{"non.existent.path", bson.A{int32(42)}}}}},
			resultType: emptyResult,
		},
		"EmptyValue": {
			update:     bson.D{{"$pullAll", bson.D{}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateArrayCompatAddToSetEach(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Document": {
			update: bson.D{{"$addToSet", bson.D{{"v", bson.D{
				{"$each", bson.A{bson.D{{"field", int32(42)}}}},
			}}}}},
		},
		"String": {
			update: bson.D{{"$addToSet", bson.D{{"v", bson.D{{"$each", bson.A{"foo"}}}}}}},
		},
		"Int32": {
			update: bson.D{{"$addToSet", bson.D{{"v", bson.D{
				{"$each", bson.A{int32(1), int32(42), int32(2)}},
			}}}}},
		},
		"NotArray": {
			update:           bson.D{{"$addToSet", bson.D{{"v", bson.D{{"$each", int32(1)}}}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/478",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.Int32s},
				{provider: shareddata.Int64s},
				{provider: shareddata.Scalars, ids: []string{
					"decimal128", "decimal128-int", "decimal128-int-zero", "decimal128-zero", "decimal128-double", "decimal128-whole",
					"double", "double-whole", "double-zero", "double-max",
					"double-smallest", "double-big", "double-1", "double-2", "double-3", "double-4", "double-5",
					"double-max-overflow", "double-min-overflow", "string", "string-double", "string-whole",
					"string-empty", "binary", "binary-empty", "objectid", "objectid-empty", "bool-false", "bool-true",
					"datetime", "datetime-epoch", "datetime-year-min", "datetime-year-max", "null", "regex", "regex-empty",
					"int32", "int32-zero", "int32-max", "int32-min", "int32-1", "int32-2", "int32-3", "timestamp",
					"timestamp-i", "int64", "int64-zero", "int64-max", "int64-min", "int64-big", "int64-double-big",
					"int64-1", "int64-2", "int64-3",
				}},
				{provider: shareddata.Decimal128s},
				{provider: shareddata.Doubles},
				{provider: shareddata.SmallDoubles},
				{provider: shareddata.Bools},
				{provider: shareddata.Nulls},
				{provider: shareddata.Timestamps},
				{provider: shareddata.DateTimes},
				{provider: shareddata.Strings},
				{provider: shareddata.Binaries},
				{provider: shareddata.ObjectIDs},
				{provider: shareddata.ObjectIDKeys},
				{provider: shareddata.Composites, ids: []string{
					"document", "document-composite", "document-composite-reverse",
					"document-composite-numerical-field-name", "document-empty", "document-null",
				}},
				{provider: shareddata.Mixed, ids: []string{"null"}},
				{provider: shareddata.Regexes},
				{provider: shareddata.PostgresEdgeCases},
				{provider: shareddata.OverflowVergeDoubles},
				{provider: shareddata.DocumentsStrings},
				{provider: shareddata.DocumentsDoubles},
				{provider: shareddata.DocumentsDeeplyNested},
				{provider: shareddata.DocumentsDocuments},
				{provider: shareddata.ArrayAndDocuments, ids: []string{"document"}},
			},
		},
		"EmptyArray": {
			filter:     bson.D{{"_id", "array-documents-nested"}},
			update:     bson.D{{"$addToSet", bson.D{{"v", bson.D{{"$each", bson.A{}}}}}}},
			resultType: emptyResult,
		},
		"ArrayMixedValuesExists": {
			update: bson.D{{"$addToSet", bson.D{{"v", bson.D{{"$each", bson.A{int32(42), "foo"}}}}}}},
		},
		"NonExistentField": {
			update: bson.D{{"$addToSet", bson.D{{"non-existent-field", bson.D{{"$each", bson.A{int32(42)}}}}}}},
		},
		"DotNotation": {
			update: bson.D{{"$addToSet", bson.D{{"v.0.foo", bson.D{{"$each", bson.A{int32(42)}}}}}}},
		},
		"DotNotationNonArray": {
			filter:     bson.D{{"_id", "array-documents-nested"}},
			update:     bson.D{{"$addToSet", bson.D{{"v.0.foo.0.bar", bson.D{{"$each", bson.A{int32(42)}}}}}}},
			resultType: emptyResult,
		},
		"DotNotatPathNotExist": {
			update: bson.D{{"$addToSet", bson.D{{"non.existent.path", bson.D{{"$each", bson.A{int32(42)}}}}}}},
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateArrayCompatPushEach(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Document": {
			update: bson.D{{"$push", bson.D{{"v", bson.D{
				{"$each", bson.A{bson.D{{"field", int32(42)}}}},
			}}}}},
		},
		"String": {
			update: bson.D{{"$push", bson.D{{"v", bson.D{{"$each", bson.A{"foo"}}}}}}},
		},
		"Int32": {
			update: bson.D{{"$push", bson.D{{"v", bson.D{
				{"$each", bson.A{int32(1), int32(42), int32(2)}},
			}}}}},
		},
		"NotArray": {
			update:     bson.D{{"$push", bson.D{{"v", bson.D{{"$each", int32(1)}}}}}},
			resultType: emptyResult,
		},
		"EmptyArray": {
			filter:           bson.D{{"_id", "array-documents-nested"}},
			update:           bson.D{{"$push", bson.D{{"v", bson.D{{"$each", bson.A{}}}}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/373",
			failsIDs: []struct {
				provider shareddata.Provider
				ids      []string
			}{
				{provider: shareddata.ArrayDocuments, ids: []string{
					"array-documents-nested", "array-two-documents",
					"array-three-documents", "array-documents-nested-duplicate",
				}},
			},
		},
		"MixedValuesExists": {
			update: bson.D{{"$push", bson.D{{"v", bson.D{{"$each", bson.A{int32(42), "foo"}}}}}}},
		},
		"NonExistentField": {
			update: bson.D{{"$push", bson.D{{"non-existent-field", bson.D{{"$each", bson.A{int32(42)}}}}}}},
		},
		"DotNotation": {
			update: bson.D{{"$push", bson.D{{"v.0.foo", bson.D{{"$each", bson.A{int32(42)}}}}}}},
		},
		"DotNotationNonArray": {
			filter:     bson.D{{"_id", "array-documents-nested"}},
			update:     bson.D{{"$push", bson.D{{"v.0.foo.0.bar", bson.D{{"$each", bson.A{int32(42)}}}}}}},
			resultType: emptyResult,
		},
		"DotNotationPathNotExist": {
			update: bson.D{{"$push", bson.D{{"non.existent.path", bson.D{{"$each", bson.A{int32(42)}}}}}}},
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateArrayCompatPull(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Int32": {
			update: bson.D{{"$pull", bson.D{{"v", int32(42)}}}},
		},
		"String": {
			update: bson.D{{"$pull", bson.D{{"v", "foo"}}}},
		},
		"StringDuplicates": {
			update: bson.D{{"$pull", bson.D{{"v", "b"}}}},
		},
		"FieldNotExist": {
			update:     bson.D{{"$pull", bson.D{{"non-existent-field", int32(42)}}}},
			resultType: emptyResult,
		},
		"Array": {
			update:     bson.D{{"$pull", bson.D{{"v", bson.A{int32(42)}}}}},
			resultType: emptyResult,
		},
		"Null": {
			update: bson.D{{"$pull", bson.D{{"v", nil}}}},
		},
		"DotNotation": {
			update: bson.D{{"$pull", bson.D{{"v.0.foo", bson.D{{"bar", "hello"}}}}}},
		},
		"DotNotationPathNotExist": {
			update:     bson.D{{"$pull", bson.D{{"non.existent.path", int32(42)}}}},
			resultType: emptyResult,
		},
		"DotNotationNotArray": {
			update:     bson.D{{"$pull", bson.D{{"v.0.foo.0.bar", int32(42)}}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}
