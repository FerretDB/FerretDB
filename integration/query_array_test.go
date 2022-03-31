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
)

func TestQueryArraySize(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", "array-empty"}, {"value", bson.A{}}},
		bson.D{{"_id", "array-one"}, {"value", bson.A{"1"}}},
		bson.D{{"_id", "array-two"}, {"value", bson.A{"1", "2"}}},
		bson.D{{"_id", "array-three"}, {"value", bson.A{"1", "2", "3"}}},
		bson.D{{"_id", "string"}, {"value", "12"}},
		bson.D{{"_id", "document"}, {"value", bson.D{{"value", bson.A{"1", "2"}}}}},
	})
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		q        bson.D
		expected []bson.D
	}{
		"int32": {
			q:        bson.D{{"value", bson.D{{"$size", int32(2)}}}},
			expected: []bson.D{{{"_id", "array-two"}, {"value", bson.A{"1", "2"}}}},
		},
		"int64": {
			q:        bson.D{{"value", bson.D{{"$size", int64(2)}}}},
			expected: []bson.D{{{"_id", "array-two"}, {"value", bson.A{"1", "2"}}}},
		},
		"float64": {
			q:        bson.D{{"value", bson.D{{"$size", 2.0}}}},
			expected: []bson.D{{{"_id", "array-two"}, {"value", bson.A{"1", "2"}}}},
		},
		"NotFound": {
			q:        bson.D{{"value", bson.D{{"$size", 4}}}},
			expected: []bson.D{{{"_id", "array-two"}, {"value", bson.A{"1", "2"}}}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual []bson.D
			cursor, err := collection.Find(ctx, tc.q)
			require.NoError(t, err)
			err = cursor.All(ctx, &actual)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
