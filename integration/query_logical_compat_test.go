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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func testQueryLogicalCompatAnd() map[string]queryCompatTestCase {
	testCases := map[string]queryCompatTestCase{
		"Zero": {
			filter: bson.D{{
				"$and", bson.A{},
			}},
			resultType: emptyResult,
		},
		"One": {
			filter: bson.D{{
				"$and", bson.A{
					bson.D{{"v", bson.D{{"$gt", int32(0)}}}},
				},
			}},
		},
		"Two": {
			filter: bson.D{{
				"$and", bson.A{
					bson.D{{"v", bson.D{{"$gt", int32(0)}}}},
					bson.D{{"v", bson.D{{"$lt", int64(42)}}}},
				},
			}},
		},
		"AndOr": {
			filter: bson.D{{
				"$and", bson.A{
					bson.D{{"v", bson.D{{"$gt", int32(0)}}}},
					bson.D{{"$or", bson.A{
						bson.D{{"v", bson.D{{"$lt", int64(42)}}}},
						bson.D{{"v", bson.D{{"$lte", 42.13}}}},
					}}},
				},
			}},
		},
		"AndAnd": {
			filter: bson.D{{
				"$and", bson.A{
					bson.D{{"$and", bson.A{
						bson.D{{"v", bson.D{{"$gt", int32(0)}}}},
						bson.D{{"v", bson.D{{"$lte", 42.13}}}},
					}}},
					bson.D{{"v", bson.D{{"$type", "int"}}}},
				},
			}},
		},
		"BadInput": {
			filter:     bson.D{{"$and", nil}},
			resultType: emptyResult,
		},
		"BadValue": {
			filter: bson.D{{
				"$and", bson.A{
					bson.D{{"v", bson.D{{"$gt", int32(0)}}}},
					true,
				},
			}},
			resultType: emptyResult,
		},
	}

	return testCases
}

func testQueryLogicalCompatOr() map[string]queryCompatTestCase {
	testCases := map[string]queryCompatTestCase{
		"Zero": {
			filter: bson.D{{
				"$or", bson.A{},
			}},
			resultType: emptyResult,
		},
		"One": {
			filter: bson.D{{
				"$or", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
				},
			}},
		},
		"Two": {
			filter: bson.D{{
				"$or", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
					bson.D{{"v", bson.D{{"$gt", int64(42)}}}},
				},
			}},
		},
		"OrAnd": {
			filter: bson.D{{
				"$or", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
					bson.D{{"$and", bson.A{
						bson.D{{"v", bson.D{{"$gt", int64(42)}}}},
						bson.D{{"v", bson.D{{"$lte", 42.13}}}},
					}}},
				},
			}},
		},
		"BadInput": {
			filter:     bson.D{{"$or", nil}},
			resultType: emptyResult,
		},
		"BadValue": {
			filter: bson.D{{
				"$or", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
					true,
				},
			}},
			resultType: emptyResult,
		},
	}

	return testCases
}

func testQueryLogicalCompatNor() map[string]queryCompatTestCase {
	testCases := map[string]queryCompatTestCase{
		"Zero": {
			filter: bson.D{{
				"$nor", bson.A{},
			}},
			resultType: emptyResult,
		},
		"One": {
			filter: bson.D{{
				"$nor", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
				},
			}},
		},
		"Two": {
			filter: bson.D{{
				"$nor", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
					bson.D{{"v", bson.D{{"$gt", int64(42)}}}},
				},
			}},
		},
		"BadInput": {
			filter:     bson.D{{"$nor", nil}},
			resultType: emptyResult,
		},
		"BadValue": {
			filter: bson.D{{
				"$nor", bson.A{
					bson.D{{"v", bson.D{{"$lt", int32(0)}}}},
					true,
				},
			}},
			resultType: emptyResult,
		},
	}

	return testCases
}

func testQueryLogicalCompatNot() map[string]queryCompatTestCase {
	testCases := map[string]queryCompatTestCase{
		"Not": {
			filter: bson.D{{
				"v", bson.D{{"$not", bson.D{{"$eq", int64(42)}}}},
			}},
		},
		"IDNull": {
			filter: bson.D{{
				"_id", bson.D{{"$not", nil}},
			}},
			resultType: emptyResult,
		},
		"NotEqNull": {
			filter: bson.D{{
				"v", bson.D{{"$not", bson.D{{"$eq", nil}}}},
			}},
		},
		"ValueRegex": {
			filter: bson.D{{
				"v", bson.D{{"$not", primitive.Regex{Pattern: "^fo"}}},
			}},
		},
		"NoSuchFieldRegex": {
			filter: bson.D{{
				"no-such-field", bson.D{{"$not", primitive.Regex{Pattern: "/someregex/"}}},
			}},
		},
		"NestedNot": {
			filter: bson.D{{
				"v", bson.D{{"$not", bson.D{{"$not", bson.D{{"$eq", int64(42)}}}}}},
			}},
		},
	}

	return testCases
}
