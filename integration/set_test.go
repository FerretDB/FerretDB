// Copyright 2021 FerretDB Set.
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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type testCase struct {
	id     string
	update bson.D
	result any
	err    *mongo.WriteError
	stat   *mongo.UpdateResult
	alt    string
}

func TestSetOperator(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]testCase{
		"BadSetString": {
			id:     "string",
			update: bson.D{{"$set", "string"}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type string instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: \"string\"}",
			},
		},
		"BadSetDouble": {
			id:     "string",
			update: bson.D{{"$set", float64(42.12345)}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type double instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: 42.12345}",
			},
			alt: "Modifiers operate on fields but we found type double instead. " +
				"For example: {$mod: {<field>: ...}} not {$set: 42.12}",
		},
		"BadSetArray": {
			id:     "string",
			update: bson.D{{"$set", bson.A{}}},
			err: &mongo.WriteError{
				Code: 9,
				Message: "Modifiers operate on fields but we found type array instead. " +
					"For example: {$mod: {<field>: ...}} not {$set: []}",
			},
		},
		"SetEmptyDoc": {
			id:     "string",
			update: bson.D{{"$set", bson.D{}}},
			result: "foo",
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
		"SetValueString": {
			id:     "string",
			update: bson.D{{"$set", bson.D{{"value", "ok value"}}}},
			result: "ok value",
			stat: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup(t)
			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", "string"}, {"value", "foo"}},
			})
			require.NoError(t, err)

			res, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)
			if tc.err != nil {
				t.Log(err)
				AssertEqualWriteError(t, tc.err, tc.alt, err)
				return
			}
			require.Equal(t, tc.stat, res)

			var actual bson.D
			err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
			require.NoError(t, err)
			expectedRes := bson.D{{"_id", "string"}, {"value", tc.result}}
			if !AssertEqualDocuments(t, expectedRes, actual) {
				t.FailNow()
			}
		})
	}
}
