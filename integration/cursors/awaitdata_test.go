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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
)

func TestCursorsTailableAwaitData(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/2283")

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000000000)
	err := db.CreateCollection(s.Ctx, t.Name(), opts)
	require.NoError(t, err)

	collection := db.Collection(t.Name())

	_, err = collection.InsertOne(ctx, bson.D{{"v", "foo"}})
	require.NoError(t, err)

	cmd := bson.D{
		{"find", collection.Name()},
		{"batchSize", 1},
		{"tailable", true},
		{"awaitData", true},
		{"maxTimeMS", 1},
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
		{"maxTimeMS", 1000 * 60 * 10},
	}

	go func() {
		time.Sleep(2 * time.Second)
		_, err := collection.InsertOne(ctx, bson.D{{"v", "bar"}})
		require.NoError(t, err)
	}()

	err = collection.Database().RunCommand(ctx, getMoreCmd).Decode(&res)
	require.NoError(t, err)

	nextBatch, nextID := getNextBatch(t, res)
	require.Equal(t, cursorID, nextID)
	require.Equal(t, 1, nextBatch.Len())
}

func TestCursorsAwaitDataErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)

	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1000)
	err := db.CreateCollection(s.Ctx, t.Name(), opts)
	require.NoError(t, err)

	collection := db.Collection(t.Name())

	_, err = collection.InsertOne(ctx, bson.D{{"v", "foo"}})
	require.NoError(t, err)

	t.Run("NonTailable", func(tt *testing.T) {
		t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/2283")

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
