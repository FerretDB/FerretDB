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

package tigris

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

// TODO This is a temporary test to check how ObjectID works.
func TestSmokeObjectID(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	// Insert, update, delete a document with a "semantic" (not 12-bytes) ObjectID.
	/*ins, err := collection.InsertOne(ctx, bson.D{{"_id", "semantic"}, {"string_value", "foo"}})
	require.NoError(t, err)

	up, err := collection.UpdateOne(ctx, bson.D{{"_id", "semantic"}}, bson.D{{"$set", bson.D{{"string_value", "bar"}}}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), up.MatchedCount)
	assert.Equal(t, int64(1), up.ModifiedCount)

	del, err := collection.DeleteOne(ctx, bson.D{{"_id", ins.InsertedID}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), del.DeletedCount)*/

	// Insert, update, delete a document with a "proper" ObjectID.
	ins, err := collection.InsertOne(ctx, bson.D{{"string_value", "foo_2"}})
	require.NoError(t, err)

	up, err := collection.UpdateOne(ctx, bson.D{{"_id", ins.InsertedID}}, bson.D{{"$set", bson.D{{"string_value", "bar_2"}}}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), up.MatchedCount)
	assert.Equal(t, int64(1), up.ModifiedCount)

	del, err := collection.DeleteOne(ctx, bson.D{{"_id", ins.InsertedID}})
	require.NoError(t, err)
	assert.Equal(t, int64(1), del.DeletedCount)
}
