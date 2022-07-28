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

	"go.mongodb.org/mongo-driver/bson"
)

func TestQueryArrayCompatAll(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"String": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{"foo"}}}}},
		},
		"StringRepeated": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{"foo", "foo", "foo"}}}}},
		},
		"StringEmpty": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{""}}}}},
		},
		"Whole": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{int32(42)}}}}},
		},
		"Zero": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{math.Copysign(0, +1)}}}}},
		},
		"Double": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{42.13}}}}},
		},
		"DoubleMax": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{math.MaxFloat64}}}}},
		},
		"DoubleMin": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{math.SmallestNonzeroFloat64}}}}},
		},
		"Nil": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{nil}}}}},
		},
		"MultiAll": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{"foo", 42}}}}},
		},
		"MultiAllWithNil": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{"foo", nil}}}}},
		},
		"ArrayEmbeddedEqual": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{bson.A{int32(42), "foo"}}}}}},
		},
		"ArrayEmbeddedReverseOrder": {
			filter:     bson.D{{"v", bson.D{{"$all", bson.A{bson.A{"foo", int32(42)}}}}}},
			resultType: emptyResult,
		},
		"Empty": {
			filter:     bson.D{{"v", bson.D{{"$all", bson.A{}}}}},
			resultType: emptyResult,
		},
		"EmptyNested": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{bson.A{}}}}}},
		},
		"NotFound": {
			filter:     bson.D{{"v", bson.D{{"$all", bson.A{"hello"}}}}},
			resultType: emptyResult,
		},
		"NaN": {
			filter: bson.D{{"v", bson.D{{"$all", bson.A{math.NaN()}}}}},
		},
	}

	testQueryCompat(t, testCases)
}
