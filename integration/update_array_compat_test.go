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
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestUpdateArrayCompatPop(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"DuplicateKeys": {
			update:     bson.D{{"$pop", bson.D{{"v", 1}, {"v", 1}}}},
			resultType: emptyResult,
		},
		"Pop": {
			update: bson.D{{"$pop", bson.D{{"v", 1}}}},
		},
		"PopFirst": {
			update: bson.D{{"$pop", bson.D{{"v", -1}}}},
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
			filter:     bson.D{{"_id", "array-documents-nested"}},
			update:     bson.D{{"$pop", bson.D{{"v.0.foo.0.bar", 1}}}},
			resultType: emptyResult,
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
			update:     bson.D{{"$pop", bson.D{{"v.array.foo.array", 1}}}},
			resultType: emptyResult,
		},
		"DotNotationObject": {
			update:     bson.D{{"$pop", bson.D{{"v.foo", 1}}}},
			resultType: emptyResult,
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
			update:        bson.D{{"$push", bson.D{{"v", int32(42)}}}},
			skipForTigris: "Some tests would fail because Tigris might convert int32 to float/int64 based on the schema",
		},
		"NonExistentField": {
			update:        bson.D{{"$push", bson.D{{"non-existent-field", int32(42)}}}},
			skipForTigris: "Tigris does not support adding new fields to documents",
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
			update:        bson.D{{"$push", bson.D{{"non.existent.path", int32(42)}}}},
			skipForTigris: "Tigris does not support adding new fields to documents",
		},
		"TwoElements": {
			update:        bson.D{{"$push", bson.D{{"non.existent.path", int32(42)}, {"v", int32(42)}}}},
			skipForTigris: "Tigris does not support adding new fields to documents",
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
			update:        bson.D{{"$addToSet", bson.D{{"v", bson.D{{"foo", "bar"}}}}}},
			skipForTigris: "Tigris does not support adding new array elements with different types",
		},
		"Int32": {
			update:        bson.D{{"$addToSet", bson.D{{"v", int32(42)}}}},
			skipForTigris: "Some tests would fail because Tigris might convert int32 to float/int64 based on the schema",
		},
		"Int64": {
			update:        bson.D{{"$addToSet", bson.D{{"v", int64(42)}}}},
			skipForTigris: "Some tests would fail because Tigris might convert int64 to float/int64 based on the schema",
		},
		"Float64": {
			update:        bson.D{{"$addToSet", bson.D{{"v", float64(42)}}}},
			skipForTigris: "Some tests would fail because of schema mismatch.",
		},
		"NonExistentField": {
			update:        bson.D{{"$addToSet", bson.D{{"non-existent-field", int32(42)}}}},
			skipForTigris: "Tigris does not support adding new fields to documents",
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
			update:        bson.D{{"$addToSet", bson.D{{"non.existent.path", int32(1)}}}},
			skipForTigris: "Tigris does not support adding new fields to documents",
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
			update:        bson.D{{"$pullAll", bson.D{{"v", bson.A{bson.D{{"field", int32(42)}}}}}}},
			skipForTigris: "We don't have such documents for Tigris.",
		},
		"Int32": {
			update: bson.D{{"$pullAll", bson.D{{"v", bson.A{int32(42)}}}}},
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
