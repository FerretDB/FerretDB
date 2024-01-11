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
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

func TestCursorsTailableAwaitDataGetMoreMaxTimeMS(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, testutil.CollectionName(t), opts)
	require.NoError(t, err)

	collection := db.Collection(testutil.CollectionName(t))

	_, err = collection.InsertOne(ctx, bson.D{{"v", "foo"}})
	require.NoError(t, err)

	cmd := bson.D{
		{"find", collection.Name()},
		{"batchSize", 1},
		{"tailable", true},
		{"awaitData", true},
	}

	var res bson.D
	err = collection.Database().RunCommand(ctx, cmd).Decode(&res)
	require.NoError(t, err)

	var firstBatch *types.Array
	firstBatch, cursorID := getFirstBatch(t, res)

	require.Equal(t, 1, firstBatch.Len())

	getMoreCmd := bson.D{
		{"getMore", cursorID},
		{"collection", collection.Name()},
		{"batchSize", 1},
		{"maxTimeMS", (10 * time.Minute).Milliseconds()},
	}

	insertChan := make(chan error)

	go func() {
		time.Sleep(100 * time.Millisecond)
		_, insertErr := collection.InsertOne(ctx, bson.D{{"v", "bar"}})
		insertChan <- insertErr
	}()

	err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
	require.NoError(t, err)

	require.NoError(t, <-insertChan)

	nextBatch, nextID := getNextBatch(t, res)
	require.Equal(t, cursorID, nextID)
	require.Equal(t, 1, nextBatch.Len())
}

func TestCursorsTailableAwaitDataNonFullBatch(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1000)
	err := db.CreateCollection(s.Ctx, testutil.CollectionName(t), opts)
	require.NoError(t, err)

	collection := db.Collection(testutil.CollectionName(t))

	bsonArr, _ := integration.GenerateDocuments(0, 2)

	_, err = collection.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	cmd := bson.D{
		{"find", collection.Name()},
		{"batchSize", 1},
		{"tailable", true},
		{"awaitData", true},
	}

	var res bson.D
	err = collection.Database().RunCommand(ctx, cmd).Decode(&res)
	require.NoError(t, err)

	var firstBatch *types.Array
	firstBatch, cursorID := getFirstBatch(t, res)

	require.Equal(t, 1, firstBatch.Len())

	getMoreCmd := bson.D{
		{"getMore", cursorID},
		{"collection", collection.Name()},
		{"batchSize", 2},
		{"maxTimeMS", (30 * time.Second).Milliseconds()},
	}

	err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
	require.NoError(t, err)

	nextBatch, nextID := getNextBatch(t, res)
	require.Equal(t, cursorID, nextID)
	require.Equal(t, 1, nextBatch.Len())

	insertChan := make(chan error)

	go func() {
		time.Sleep(1 * time.Second)
		_, insertErr := collection.InsertOne(ctx, bson.D{{"v", "bar"}})
		insertChan <- insertErr
	}()

	err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
	require.NoError(t, err)

	require.NoError(t, <-insertChan)

	nextBatch, nextID = getNextBatch(t, res)
	require.Equal(t, cursorID, nextID)
	require.Equal(t, 1, nextBatch.Len())
}

func TestCursorsAwaitDataErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1000)
	err := db.CreateCollection(s.Ctx, testutil.CollectionName(t), opts)
	require.NoError(t, err)

	collection := db.Collection(testutil.CollectionName(t))

	_, err = collection.InsertOne(ctx, bson.D{{"v", "foo"}})
	require.NoError(t, err)

	t.Run("NonTailable", func(t *testing.T) {
		err = collection.Database().RunCommand(ctx, bson.D{
			{"find", collection.Name()},
			{"batchSize", 1},
			{"awaitData", true},
		}).Err()

		expectedErr := mongo.CommandError{
			Code:    9,
			Name:    "FailedToParse",
			Message: "Cannot set 'awaitData' without also setting 'tailable'",
		}
		integration.AssertEqualCommandError(t, expectedErr, err)
	})
}
func TestCursorsTailableAwaitDataTODO(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, testutil.CollectionName(t), opts)
	require.NoError(t, err)

	collection := db.Collection(t.Name())

	bsonArr, _ := integration.GenerateDocuments(0, 3)

	_, err = collection.InsertMany(ctx, bsonArr)
	require.NoError(t, err)

	findOpts := options.Find().SetCursorType(options.TailableAwait).SetMaxAwaitTime(20 * time.Millisecond).SetBatchSize(1)

	cur, err := collection.Find(ctx, bson.D{}, findOpts)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		require.True(t, cur.Next(ctx))
	}

	require.False(t, cur.Next(ctx))

}

func TestCursorsTailableAwaitData(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, testutil.CollectionName(t), opts)
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
			{"awaitData", true},
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

	t.Run("GetMore", func(t *testing.T) {
		getMoreCmd := bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 1},
		}

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
		getMoreCmd := bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 1},
		}

		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, 0, nextBatch.Len())
		assert.Equal(t, cursorID, nextID)
	})

	t.Run("GetMoreNewDoc", func(t *testing.T) {
		getMoreCmd := bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 1},
			{"maxTimeMS", 2000},
		}

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
		getMoreCmd := bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 1},
		}

		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, 0, nextBatch.Len())
		assert.Equal(t, cursorID, nextID)
	})
}

func TestCursorsTailableAwaitDataTwoCursorsSameCollection(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, testutil.CollectionName(t), opts)
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
		{"awaitData", true},
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
		{"maxTimeMS", (2 * time.Second).Milliseconds()},
	}

	getMoreCmd2 := bson.D{
		{"getMore", cursorID2},
		{"collection", collection.Name()},
		{"batchSize", 1},
		{"maxTimeMS", (2 * time.Second).Milliseconds()},
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

func TestCursorsTailableAwaitDataStress(t *testing.T) {
	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	bsonArr, arr := integration.GenerateDocuments(0, 5)

	var count atomic.Int32

	teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
		testID := count.Add(1)
		collName := fmt.Sprintf("%s_%d", testutil.CollectionName(t), testID)

		ready <- struct{}{}
		<-start

		opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
		err := db.CreateCollection(s.Ctx, collName, opts)
		require.NoError(t, err)

		collection := db.Collection(collName)

		_, err = collection.InsertMany(ctx, bsonArr)
		require.NoError(t, err)

		var cursorID1, cursorID2 any

		cmd := bson.D{
			{"find", collection.Name()},
			{"batchSize", 1},
			{"tailable", true},
			{"awaitData", true},
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
			{"maxTimeMS", (10 * time.Millisecond).Milliseconds()},
		}
		getMoreCmd2 := bson.D{
			{"getMore", cursorID2},
			{"collection", collection.Name()},
			{"batchSize", 1},
			{"maxTimeMS", (10 * time.Millisecond).Milliseconds()},
		}

		for i := 0; i < 5; i++ {
			_ = collection.Database().RunCommand(ctx, getMoreCmd1)
			_ = collection.Database().RunCommand(ctx, getMoreCmd2)
		}
	})
}

func TestCursorsAwaitDataFirstBatchMaxTimeMS(t *testing.T) {
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
			{"awaitData", true},
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
		time.Sleep(150 * time.Millisecond)
		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, 0, nextBatch.Len())
		assert.Equal(t, cursorID, nextID)
	})
}

func TestCursorsAwaitDataGetMoreAfterInsertion(t *testing.T) {
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
			{"tailable", true},
			{"awaitData", true},
			{"batchSize", 1},
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

	t.Run("GetMoreEmpty", func(tt *testing.T) {
		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, 0, nextBatch.Len())
		assert.Equal(t, cursorID, nextID)
	})

	t.Run("GetMoreNewDoc", func(tt *testing.T) {
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

	t.Run("GetMoreEmptyAfterInsertion", func(tt *testing.T) {
		var res bson.D
		err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
		require.NoError(t, err)

		nextBatch, nextID := getNextBatch(t, res)
		require.Equal(t, 0, nextBatch.Len())
		assert.Equal(t, cursorID, nextID)
	})
}
