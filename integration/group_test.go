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

func TestGroupDistinct(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", 1}, {"item", "abc"}},
		bson.D{{"_id", 2}, {"item", "jkl"}},
		bson.D{{"_id", 3}, {"item", "xyz"}},
		bson.D{{"_id", 4}, {"item", "xyz"}},
		bson.D{{"_id", 5}, {"item", "abc"}},
		bson.D{{"_id", 6}, {"item", "def"}},
		bson.D{{"_id", 7}, {"item", "def"}},
		bson.D{{"_id", 8}, {"item", "abc"}},
	})
	require.NoError(t, err)

	group := bson.D{{"$group", bson.D{{"_id", "$item"}}}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{group})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []bson.D{
		bson.D{{"_id", "abc"}},
		bson.D{{"_id", "def"}},
		bson.D{{"_id", "jkl"}},
		bson.D{{"_id", "xyz"}},
	}, results)
}
