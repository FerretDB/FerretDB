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

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestMostCommandsAreCaseSensitive(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)
	db := collection.Database()

	res := db.RunCommand(ctx, bson.D{{"listcollections", 1}})
	err := res.Err()
	require.Error(t, err)
	assert.Equal(t, mongo.CommandError{Code: 59, Name: "CommandNotFound", Message: `no such command: 'listcollections'`}, err)

	res = db.RunCommand(ctx, bson.D{{"listCollections", 1}})
	assert.NoError(t, res.Err())

	// special case
	res = db.RunCommand(ctx, bson.D{{"ismaster", 1}})
	assert.NoError(t, res.Err())
	res = db.RunCommand(ctx, bson.D{{"isMaster", 1}})
	assert.NoError(t, res.Err())
}

func TestFindNothing(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t)

	cursor, err := collection.Find(ctx, bson.D{})
	require.NoError(t, err)
	var docs []bson.D
	err = cursor.All(ctx, &docs)
	require.NoError(t, err)
	assert.Equal(t, []bson.D(nil), docs)

	var doc bson.D
	err = collection.FindOne(ctx, bson.D{}).Decode(&doc)
	require.Equal(t, mongo.ErrNoDocuments, err)
	assert.Equal(t, bson.D(nil), doc)
}

func TestInsertFindScalars(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars)

	for _, expected := range shareddata.Scalars.Docs() {
		id := expected.Map()["_id"]
		var actual bson.D
		err := collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&actual)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	}
}
