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

package cursors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestCursorsTailableErrors(t *testing.T) {
	t.Parallel()

	t.Run("NonCapped", func(t *testing.T) {
		t.Parallel()

		ctx, collection := setup.Setup(t, shareddata.Scalars)

		for _, ct := range []options.CursorType{options.Tailable, options.TailableAwait} {
			cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetCursorType(ct))
			expected := mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: "error processing query: " +
					"ns=TestCursorsTailableErrors-NonCapped.TestCursorsTailableErrors-NonCappedTree: $and\nSort: {}\nProj: {}\n " +
					"tailable cursor requested on non capped collection",
			}
			integration.AssertEqualAltCommandError(t, expected, "tailable cursor requested on non capped collection", err)
			assert.Nil(t, cursor)
		}
	})

	t.Run("GetMoreDifferentCollection", func(t *testing.T) {
		t.Parallel()

		s := setup.SetupWithOpts(t, nil)

		db, ctx := s.Collection.Database(), s.Ctx

		opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
		err := db.CreateCollection(s.Ctx, t.Name(), opts)
		require.NoError(t, err)

		collection := db.Collection(t.Name())

		bsonArr, arr := integration.GenerateDocuments(0, 3)

		_, err = collection.InsertMany(ctx, bsonArr)
		require.NoError(t, err)

		var cursorID any

		findCmd := bson.D{
			{"find", collection.Name()},
			{"batchSize", 1},
			{"tailable", true},
		}

		var res bson.D
		err = collection.Database().RunCommand(ctx, findCmd).Decode(&res)
		require.NoError(t, err)

		var firstBatch *types.Array
		firstBatch, cursorID = getFirstBatch(t, res)

		expectedFirstBatch := integration.ConvertDocuments(t, arr[:1])
		require.Equal(t, len(expectedFirstBatch), firstBatch.Len())
		require.Equal(t, expectedFirstBatch[0], must.NotFail(firstBatch.Get(0)))

		getMoreCmd := bson.D{
			{"getMore", cursorID},
			{"collection", "different-collection"},
			{"batchSize", 1},
		}

		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)

		expected := mongo.CommandError{
			Code: 13,
			Name: "Unauthorized",
			Message: "Requested getMore on namespace 'TestCursorsTailableErrors-GetMoreDifferentCollection.different-collection', " +
				"but cursor belongs to a different namespace " +
				"TestCursorsTailableErrors-GetMoreDifferentCollection.TestCursorsTailableErrors/GetMoreDifferentCollection",
		}
		integration.AssertEqualCommandError(t, expected, err)

		// Check if cursor is not closed after the error
		err = collection.Database().RunCommand(ctx, bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 1},
		}).Decode(&res)

		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, cursorID, nextID)

		doc, _ := nextBatch.Get(0)
		require.NotNil(t, doc)
	})
}

func TestCursorsTailable(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, t.Name(), opts)
	require.NoError(t, err)

	collection := db.Collection(t.Name())

	bsonArr, arr := integration.GenerateDocuments(0, 3)

	_, err = collection.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	var cursorID any

	t.Run("FirstBatch", func(t *testing.T) {
		cmd := bson.D{
			{"find", collection.Name()},
			{"batchSize", 1},
			{"tailable", true},
		}

		var res bson.D
		err = collection.Database().RunCommand(ctx, cmd).Decode(&res)
		require.NoError(t, err)

		var firstBatch *types.Array
		firstBatch, cursorID = getFirstBatch(t, res)

		expectedFirstBatch := integration.ConvertDocuments(t, arr[:1])
		require.Equal(t, len(expectedFirstBatch), firstBatch.Len())
		require.Equal(t, expectedFirstBatch[0], must.NotFail(firstBatch.Get(0)))
	})

	getMoreCmd := bson.D{
		{"getMore", cursorID},
		{"collection", collection.Name()},
		{"batchSize", 1},
	}

	t.Run("GetMore", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			var res bson.D
			err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
			require.NoError(t, err)

			nextBatch, nextID := getNextBatch(t, res)
			expectedNextBatch := integration.ConvertDocuments(t, arr[i+1:i+2])

			assert.Equal(t, cursorID, nextID)

			require.Equal(t, len(expectedNextBatch), nextBatch.Len())
			require.Equal(t, expectedNextBatch[0], must.NotFail(nextBatch.Get(0)))
		}
	})

	t.Run("GetMoreEmpty", func(t *testing.T) {
		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, 0, nextBatch.Len())
		assert.Equal(t, cursorID, nextID)
	})

	t.Run("GetMoreNewDoc", func(t *testing.T) {
		newDoc := bson.D{{"_id", "new"}}
		_, err = collection.InsertOne(ctx, newDoc)
		require.NoError(t, err)

		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)

		assert.Equal(t, cursorID, nextID)

		require.Equal(t, 1, nextBatch.Len())
		require.Equal(t, integration.ConvertDocument(t, newDoc), must.NotFail(nextBatch.Get(0)))
	})

	t.Run("GetMoreEmptyAfterInsertion", func(t *testing.T) {
		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, 0, nextBatch.Len())
		assert.Equal(t, cursorID, nextID)
	})
}

func TestCursorsTailableTwoCursorsSameCollection(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, t.Name(), opts)
	require.NoError(t, err)

	collection := db.Collection(t.Name())

	bsonArr, arr := integration.GenerateDocuments(0, 50)

	_, err = collection.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	var cursorID1, cursorID2 any

	cmd := bson.D{
		{"find", collection.Name()},
		{"batchSize", 1},
		{"tailable", true},
	}

	var res bson.D

	err = collection.Database().RunCommand(ctx, cmd).Decode(&res)
	require.NoError(t, err)

	var firstBatch1 *types.Array
	firstBatch1, cursorID1 = getFirstBatch(t, res)

	err = collection.Database().RunCommand(ctx, cmd).Decode(&res)
	require.NoError(t, err)

	var firstBatch2 *types.Array
	firstBatch2, cursorID2 = getFirstBatch(t, res)

	expectedFirstBatch := integration.ConvertDocuments(t, arr[:1])

	require.Equal(t, len(expectedFirstBatch), firstBatch1.Len())
	require.Equal(t, expectedFirstBatch[0], must.NotFail(firstBatch1.Get(0)))

	require.Equal(t, len(expectedFirstBatch), firstBatch2.Len())
	require.Equal(t, expectedFirstBatch[0], must.NotFail(firstBatch2.Get(0)))

	getMoreCmd1 := bson.D{
		{"getMore", cursorID1},
		{"collection", collection.Name()},
		{"batchSize", 1},
	}

	getMoreCmd2 := bson.D{
		{"getMore", cursorID2},
		{"collection", collection.Name()},
		{"batchSize", 1},
	}

	for i := 0; i < 49; i++ {
		err = collection.Database().RunCommand(ctx, getMoreCmd1).Decode(&res)
		require.NoError(t, err)

		nextBatch1, nextID1 := getNextBatch(t, res)

		err = collection.Database().RunCommand(ctx, getMoreCmd2).Decode(&res)
		require.NoError(t, err)

		nextBatch2, nextID2 := getNextBatch(t, res)

		expectedNextBatch := integration.ConvertDocuments(t, arr[i+1:i+2])

		assert.Equal(t, cursorID1, nextID1)
		require.Equal(t, len(expectedNextBatch), nextBatch1.Len())
		require.Equal(t, expectedNextBatch[0], must.NotFail(nextBatch1.Get(0)))

		assert.Equal(t, cursorID2, nextID2)
		require.Equal(t, len(expectedNextBatch), nextBatch2.Len())
		require.Equal(t, expectedNextBatch[0], must.NotFail(nextBatch2.Get(0)))
	}

	err = collection.Database().RunCommand(ctx, getMoreCmd1).Decode(&res)
	require.NoError(t, err)

	nextBatch1, nextID1 := getNextBatch(t, res)

	err = collection.Database().RunCommand(ctx, getMoreCmd2).Decode(&res)
	require.NoError(t, err)

	nextBatch2, nextID2 := getNextBatch(t, res)

	require.Equal(t, 0, nextBatch1.Len())
	assert.Equal(t, cursorID1, nextID1)

	require.Equal(t, 0, nextBatch2.Len())
	assert.Equal(t, cursorID2, nextID2)
}

func TestCursorsTailableFirstBatchMaxTimeMS(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, t.Name(), opts)
	require.NoError(t, err)

	collection := db.Collection(t.Name())

	bsonArr, arr := integration.GenerateDocuments(0, 3)

	_, err = collection.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	var cursorID any

	t.Run("FirstBatch", func(t *testing.T) {
		cmd := bson.D{
			{"find", collection.Name()},
			{"batchSize", 1},
			{"tailable", true},
			{"maxTimeMS", 200},
		}

		var res bson.D
		err = collection.Database().RunCommand(ctx, cmd).Decode(&res)
		require.NoError(t, err)

		var firstBatch *types.Array
		firstBatch, cursorID = getFirstBatch(t, res)

		expectedFirstBatch := integration.ConvertDocuments(t, arr[:1])
		require.Equal(t, len(expectedFirstBatch), firstBatch.Len())
		require.Equal(t, expectedFirstBatch[0], must.NotFail(firstBatch.Get(0)))
	})

	getMoreCmd := bson.D{
		{"getMore", cursorID},
		{"collection", collection.Name()},
		{"batchSize", 1},
	}

	t.Run("GetMore", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			time.Sleep(100 * time.Millisecond)
			var res bson.D
			err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
			require.NoError(t, err)

			nextBatch, nextID := getNextBatch(t, res)
			expectedNextBatch := integration.ConvertDocuments(t, arr[i+1:i+2])

			assert.Equal(t, cursorID, nextID)

			require.Equal(t, len(expectedNextBatch), nextBatch.Len())
			require.Equal(t, expectedNextBatch[0], must.NotFail(nextBatch.Get(0)))
		}
	})

	t.Run("GetMoreEmpty", func(t *testing.T) {
		time.Sleep(100 * time.Millisecond)
		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, 0, nextBatch.Len())
		assert.Equal(t, cursorID, nextID)
	})
}
