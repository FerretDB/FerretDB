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

func TestQueryProjectionCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"FindProjectionInclusions": {
			filter: bson.D{{"_id", "document-composite"}},
			// TODO: https://github.com/FerretDB/FerretDB/issues/537
			projection: bson.D{{"foo", int32(1)}, {"42", true}},
		},
		"FindProjectionExclusions": {
			filter: bson.D{{"_id", "document-composite"}},
			// TODO: https://github.com/FerretDB/FerretDB/issues/537
			projection: bson.D{{"foo", int32(0)}, {"array", false}},
		},
		"ProjectionSliceNonArrayField": {
			filter:     bson.D{{"_id", "document"}},
			projection: bson.D{{"_id", bson.D{{"$slice", 1}}}},
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryProjectionCompatElemMatch(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"ElemMatch": {
			filter: bson.D{{"_id", "document-composite"}},
			projection: bson.D{{
				"v",
				bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$eq", 42}}}}}},
			}},
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryProjectionCompatSlice(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"ElemMatch": {
			filter: bson.D{{"_id", "document-composite"}},
			projection: bson.D{{
				"v",
				bson.D{{"$elemMatch", bson.D{{"field", bson.D{{"$eq", 42}}}}}},
			}},
		},
	}

	testQueryCompat(t, testCases)
}
