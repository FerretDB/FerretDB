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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestMatch(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		match    bson.D
		expected []bson.D
	}{
		"Simple": {
			match: bson.D{{"a", 1}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
				bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
			},
		},
		"MultipleFields": {
			match: bson.D{{"a", 1}, {"b", 2}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
			},
		},
		"Nested": {
			match: bson.D{{"c", bson.D{{"name", "Felipe"}, {"age", 12}}}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)}, {"a", int32(2)}, {"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}}},
			},
		},
		"And": {
			match: bson.D{{"$and", bson.A{
				bson.D{{"a", 1}},
				bson.D{{"b", 2}},
			}}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
			},
		},
		"Or": {
			match: bson.D{{"$or", bson.A{
				bson.D{{"a", 1}},
				bson.D{{"b", 8}},
			}}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
				bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
			},
		},
		"GreaterThan": {
			match: bson.D{{"a", bson.D{{"$gt", 1}}}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)},
					{"a", int32(2)},
					{"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}},
				},
			},
		},
		"NotEqual": {
			match: bson.D{{"_id", bson.D{{"$ne", 3}}}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
				bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
			},
		},
		"In": {
			match: bson.D{{"_id", bson.D{{"$in", bson.A{1, 2}}}}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
				bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
			},
		},
		"NotIn": {
			match: bson.D{{"_id", bson.D{{"$nin", bson.A{1, 2}}}}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)}, {"a", int32(2)}, {"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}}},
			},
		},
		"Exists": {
			match: bson.D{{"c", bson.D{{"$exists", true}}}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)}, {"a", int32(2)}, {"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}}},
			},
		},
		"ExistsFalse": {
			match: bson.D{{"c", bson.D{{"$exists", false}}}},
			expected: []bson.D{
				bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
				bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
			},
		},
		"NestedExists": {
			match: bson.D{{"c", bson.D{{"name", bson.D{{"$exists", true}}}}}},
			expected: []bson.D{
				bson.D{
					{"_id", int32(3)}, {"a", int32(2)}, {"b", int32(3)},
					{"c", bson.D{{"name", "Felipe"}, {"age", int32(12)}}}},
			},
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx, collection := setup(t)

			_, err := collection.InsertMany(ctx, []any{
				bson.D{{"_id", 1}, {"a", 1}, {"b", 2}},
				bson.D{{"_id", 2}, {"a", 1}, {"b", 8}},
				bson.D{{"_id", 3}, {"a", 2}, {"b", 3}, {"c", bson.D{{"name", "Felipe"}, {"age", 12}}}},
			})
			require.NoError(t, err)

			match := bson.D{{"$match", tc.match}}
			cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match})
			require.NoError(t, err)

			var results []bson.D
			if err := cursor.All(ctx, &results); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expected, results)
		})
	}

}
