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

func TestQueryLogicalCompatAnd(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"And": {
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
					nil,
				},
			}},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/962",
		},
	}

	testQueryCompat(t, testCases)
}
