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

func TestUpdateArrayCompatPush(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"Int32": {
			update: bson.D{{"$push", bson.D{{"v", int32(42)}}}},
		},
		"StringMany": {
			update: bson.D{{"$push", bson.D{{"v", "foo"}, {"v", "bar"}, {"v", "baz"}}}},
		},
		//		"DotNotation": {

		//		},
		"NonExistentField": {
			update:     bson.D{{"$push", bson.D{{"non-existent-field", int32(42)}}}},
			resultType: emptyResult,
		},
		"NonExistentPath": {
			update:     bson.D{{"$push", bson.D{{"non.existent.path", int32(42)}}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}
