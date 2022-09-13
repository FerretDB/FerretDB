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
	"math"
	"testing"
)

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

func TestUpdateFieldCompatMax(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		//"Int32Lower": {
		//	update: bson.D{{"$max", bson.D{{"v", int32(30)}}}},
		//},
		"Int32Higher": {
			update: bson.D{{"$max", bson.D{{"v", int32(60)}}}},
		},
		//"Int32Negative": {
		//	update: bson.D{{"$max", bson.D{{"v", int32(-22)}}}},
		//},
		//"StringIntHigher": {
		//	update: bson.D{{"$max", bson.D{{"v", "60"}}}},
		//},
		//"StringIntLower": {
		//	update: bson.D{{"$max", bson.D{{"v", "30"}}}},
		//},
		//"Double": {
		//	update: bson.D{{"$max", bson.D{{"v", 62.34}}}},
		//},
		//"Times": {
		//	update: bson.D{{"$max", bson.D{{"v", bson.D{{"foo", "bar"}}}}}},
		//},
		//"Document": {
		//	update: bson.D{{"$max", bson.D{{"v", bson.D{{"foo", "bar"}}}}}},
		//},
		//"DateTime": {
		//	update: bson.D{{"$max", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 12, 18, 42, 123000000, time.UTC))}}}},
		//},
		//"DateTimeLower": {
		//	update: bson.D{{"$max", bson.D{{"v", primitive.NewDateTimeFromTime(time.Date(2021, 11, 1, 3, 18, 42, 123000000, time.UTC))}}}},
		//},
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
		"DotNotationDocument": {
			update: bson.D{{"$unset", bson.D{{"v.foo", ""}}}},
		},
		"DotNotationDocumentNonExisting": {
			update:     bson.D{{"$unset", bson.D{{"foo.bar", ""}}}},
			resultType: emptyResult,
		},
		"DotNotationArrayField": {
			update:        bson.D{{"$unset", bson.D{{"v.array.0", ""}}}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/908",
		},
		"DotNotationArrayNonExisting": {
			update:     bson.D{{"$unset", bson.D{{"foo.0.baz", int32(1)}}}},
			resultType: emptyResult,
		},
	}

	testUpdateCompat(t, testCases)
}
