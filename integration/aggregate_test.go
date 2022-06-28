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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestSimpleMatch(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	_, err := collection.InsertMany(ctx, []any{
		bson.D{{"_id", 1}, {"a", 1}, {"b", 2}},
		bson.D{{"_id", 2}, {"a", 1}, {"b", 8}},
		bson.D{{"_id", 3}, {"a", 2}, {"b", 3}},
	})
	require.NoError(t, err)

	match := bson.D{{"$match", bson.D{{"a", 1}}}}
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{match})
	require.NoError(t, err)

	var results []bson.D
	if err := cursor.All(ctx, &results); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(results), 2)

	expected := []bson.D{
		bson.D{{"_id", int32(1)}, {"a", int32(1)}, {"b", int32(2)}},
		bson.D{{"_id", int32(2)}, {"a", int32(1)}, {"b", int32(8)}},
	}
	assert.Equal(t, expected, results)
}
