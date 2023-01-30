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

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestUpdateArrayCompatPop(t *testing.T) {
	t.Parallel()

	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1834")

	testCases := map[string]updateCompatTestCase{
		"Pop": {
			update: bson.D{{"$pop", bson.D{{"v", 1}}}},
		},
		"PopFirst": {
			update: bson.D{{"$pop", bson.D{{"v", -1}}}},
		},
		"DotNotation": {
			filter: bson.D{{"_id", "array-documents-nested"}},
			update: bson.D{{"$pop", bson.D{{"v.0.foo", 1}}}},
		},
		"DotNotationPopFirst": {
			filter: bson.D{{"_id", "array-documents-nested"}},
			update: bson.D{{"$pop", bson.D{{"v.0.foo", -1}}}},
		},
	}

	testUpdateCompat(t, testCases)
}

func testUpdateArrayCompatPop() map[string]updateCompatEmptyResultTestCase {
	//setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1834")

	testCases := map[string]updateCompatEmptyResultTestCase{
		"DuplicateKeys": {
			update: bson.D{{"$pop", bson.D{{"v", 1}, {"v", 1}}}},
		},
		"NonExistentField": {
			update: bson.D{{"$pop", bson.D{{"non-existent-field", 1}}}},
		},
		"DotNotationNonArray": {
			filter: bson.D{{"_id", "array-documents-nested"}},
			update: bson.D{{"$pop", bson.D{{"v.0.foo.0.bar", 1}}}},
		},
		"DotNotationNonExistentPath": {
			update: bson.D{{"$pop", bson.D{{"non.existent.path", 1}}}},
		},
		"PopEmptyValue": {
			update: bson.D{{"$pop", bson.D{}}},
		},
		"PopNotValidValueString": {
			update: bson.D{{"$pop", bson.D{{"v", "foo"}}}},
		},
		"PopNotValidValueInt": {
			update: bson.D{{"$pop", bson.D{{"v", int32(42)}}}},
		},
	}

	return testCases
}

func TestUpdateArrayCompatPush(t *testing.T) {
	t.Parallel()

	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1834")

	testCases := map[string]updateCompatTestCase{
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

func testUpdateArrayCompatPush() map[string]updateCompatEmptyResultTestCase {
	// setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/1834")

	testCases := map[string]updateCompatEmptyResultTestCase{
		"DuplicateKeys": {
			update: bson.D{{"$push", bson.D{{"v", "foo"}, {"v", "bar"}}}},
		},
		"DotNotationNonArray": {
			filter: bson.D{{"_id", "array-documents-nested"}},
			update: bson.D{{"$push", bson.D{{"v.0.foo.0.bar", "boo"}}}},
		},
	}

	return testCases
}
