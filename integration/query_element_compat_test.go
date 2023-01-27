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

	"go.mongodb.org/mongo-driver/bson"
)

func testQueryElementCompatExists() map[string]queryCompatTestCase {
	testCases := map[string]queryCompatTestCase{
		"Exists": {
			filter: bson.D{{"_id", bson.D{{"$exists", true}}}},
		},
		"ExistsSecondField": {
			filter: bson.D{{"v", bson.D{{"$exists", true}}}},
		},
		"NonExistentField": {
			filter:     bson.D{{"non-existent", bson.D{{"$exists", true}}}},
			resultType: emptyResult,
		},
		"ExistsFalse": {
			filter: bson.D{{"field", bson.D{{"$exists", false}}}},
		},
		"NonBool": {
			filter: bson.D{{"_id", bson.D{{"$exists", -123}}}},
		},
	}

	return testCases
}

func testQueryElementCompatElementType() map[string]queryCompatTestCase {
	testCases := map[string]queryCompatTestCase{
		"Document": {
			filter: bson.D{{"v", bson.D{{"$type", "object"}}}},
		},
		"Array": {
			filter: bson.D{{"v", bson.D{{"$type", "array"}}}},
		},
		"Double": {
			filter: bson.D{{"v", bson.D{{"$type", "double"}}}},
		},
		"String": {
			filter: bson.D{{"v", bson.D{{"$type", "string"}}}},
		},
		"Binary": {
			filter: bson.D{{"v", bson.D{{"$type", "binData"}}}},
		},
		"ObjectID": {
			filter: bson.D{{"v", bson.D{{"$type", "objectId"}}}},
		},
		"Bool": {
			filter: bson.D{{"v", bson.D{{"$type", "bool"}}}},
		},
		"Datetime": {
			filter: bson.D{{"v", bson.D{{"$type", "date"}}}},
		},
		"Null": {
			filter: bson.D{{"v", bson.D{{"$type", "null"}}}},
		},
		"Regex": {
			filter: bson.D{{"v", bson.D{{"$type", "regex"}}}},
		},
		"Integer": {
			filter: bson.D{{"v", bson.D{{"$type", "int"}}}},
		},
		"Timestamp": {
			filter: bson.D{{"v", bson.D{{"$type", "timestamp"}}}},
		},
		"Long": {
			filter: bson.D{{"v", bson.D{{"$type", "long"}}}},
		},
		"Number": {
			filter: bson.D{{"v", bson.D{{"$type", "number"}}}},
		},
		"BadTypeCode": {
			filter:     bson.D{{"v", bson.D{{"$type", 42}}}},
			resultType: emptyResult,
		},
		"BadTypeName": {
			filter:     bson.D{{"v", bson.D{{"$type", "float"}}}},
			resultType: emptyResult,
		},
		"IntegerNumericalInput": {
			filter: bson.D{{"v", bson.D{{"$type", 16}}}},
		},
		"FloatTypeCode": {
			filter: bson.D{{"v", bson.D{{"$type", 16.0}}}},
		},
		"TypeArrayAliases": {
			filter: bson.D{{"v", bson.D{{"$type", []any{"bool", "binData"}}}}},
		},
		"TypeArrayCodes": {
			filter: bson.D{{"v", bson.D{{"$type", []any{5, 8}}}}},
		},
		"TypeArrayAliasAndCodeMixed": {
			filter: bson.D{{"v", bson.D{{"$type", []any{5, "binData"}}}}},
		},
		"TypeArrayBadValue": {
			filter:     bson.D{{"v", bson.D{{"$type", []any{"binData", -123}}}}},
			resultType: emptyResult,
		},
		"TypeArrayBadValuePlusInf": {
			filter:     bson.D{{"v", bson.D{{"$type", []any{"binData", math.Inf(+1)}}}}},
			resultType: emptyResult,
		},
		"TypeArrayBadValueMinusInf": {
			filter:     bson.D{{"v", bson.D{{"$type", []any{"binData", math.Inf(-1)}}}}},
			resultType: emptyResult,
		},
		"TypeArrayBadValueNegativeFloat": {
			filter:     bson.D{{"v", bson.D{{"$type", []any{"binData", -1.123}}}}},
			resultType: emptyResult,
		},
		"TypeArrayFloat": {
			filter: bson.D{{"v", bson.D{{"$type", []any{5, 8.0}}}}},
		},
	}

	return testCases
}
