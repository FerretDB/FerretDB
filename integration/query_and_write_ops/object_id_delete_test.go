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

package query_and_write_ops

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestStringAsID(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	// Insert, update, delete a document with a string (not 12-bytes array) id.
	ins, err := collection.InsertOne(ctx, bson.D{{"_id", "string_id"}, {"string_value", "foo"}})
	require.NoError(t, err)

	up, err := collection.UpdateOne(ctx, bson.D{{"_id", "string_id"}}, bson.D{{"$set", bson.D{{"string_value", "bar"}}}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), up.MatchedCount)
	assert.Equal(t, int64(1), up.ModifiedCount)

	var doc bson.D
	err = collection.FindOne(ctx, bson.D{{"_id", "string_id"}}).Decode(&doc)
	require.NoError(t, err)
	integration.AssertEqualDocuments(t, bson.D{{"_id", "string_id"}, {"string_value", "bar"}}, doc)

	del, err := collection.DeleteOne(ctx, bson.D{{"_id", ins.InsertedID}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), del.DeletedCount)

	err = collection.FindOne(ctx, bson.D{{"_id", "string_id"}}).Decode(&doc)
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func TestSmokeObjectIDBinary(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	id, err := primitive.ObjectIDFromHex("62e7d8a3d23915343c4a5f3a")
	require.NoError(t, err)

	// Insert, update, delete a document with a "proper" ObjectID.
	ins, err := collection.InsertOne(ctx, bson.D{{"_id", id}, {"string_value", "foo_2"}})
	require.NoError(t, err)
	insID := ins.InsertedID.(primitive.ObjectID)
	require.Equal(t, id, insID)

	up, err := collection.UpdateOne(ctx, bson.D{{"_id", id}}, bson.D{{"$set", bson.D{{"string_value", "bar_2"}}}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), up.MatchedCount)
	assert.Equal(t, int64(1), up.ModifiedCount)

	var doc bson.D
	err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&doc)
	require.NoError(t, err)
	integration.AssertEqualDocuments(t, bson.D{{"_id", id}, {"string_value", "bar_2"}}, doc)

	del, err := collection.DeleteOne(ctx, bson.D{{"_id", id}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), del.DeletedCount)

	err = collection.FindOne(ctx, bson.D{}).Decode(&doc)
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func TestDeleteMany(t *testing.T) {
	// In this test we insert and delete many (two) documents by filter.

	t.Parallel()
	ctx, collection := setup.Setup(t)

	_, err := collection.InsertOne(ctx, bson.D{{"string_value", "foo"}})
	require.NoError(t, err)

	_, err = collection.InsertOne(ctx, bson.D{{"string_value", "foo"}})
	require.NoError(t, err)

	del, err := collection.DeleteMany(ctx, bson.D{{"string_value", "foo"}})
	require.NoError(t, err)
	assert.Equal(t, int64(2), del.DeletedCount)
}
