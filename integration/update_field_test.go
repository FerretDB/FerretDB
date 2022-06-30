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
	"time"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// This file is for tests of:
// - $currentDate
// - $inc
// - $min
// - $max
// - $mul
// - $rename
// - $set
// - $setOnInsert
// - $unset

func TestUpdateTimestamp(t *testing.T) {
	t.Parallel()

	// store the current timestamp with $currentDate operator;
	t.Run("currentDateReadBack", func(t *testing.T) {
		maxDifference := time.Duration(2 * time.Second)
		nowTimestamp := primitive.Timestamp{T: uint32(time.Now().Unix()), I: uint32(0)}
		id := "string-empty"

		stat := &mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 1,
			UpsertedCount: 0,
		}
		path := types.NewPathFromString("value")
		result := bson.D{{"_id", id}, {"value", nowTimestamp}}

		ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

		update := bson.D{{"$currentDate", bson.D{{"value", bson.D{{"$type", "timestamp"}}}}}}
		res, err := collection.UpdateOne(ctx, bson.D{{"_id", id}}, update)
		require.NoError(t, err)
		require.Equal(t, stat, res)

		// read it, check that it is close to the current time;
		var actualBSON bson.D
		err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&actualBSON)
		require.NoError(t, err)

		expected := ConvertDocument(t, result)
		actualDocument := ConvertDocument(t, actualBSON)

		testutil.CompareAndSetByPathTime(t, expected, actualDocument, maxDifference, path)

		// write a new timestamp value with the same time;
		updateBSON := bson.D{{"$set", bson.D{{"value", nowTimestamp}}}}
		expectedBSON := bson.D{{"_id", id}, {"value", nowTimestamp}}
		res, err = collection.UpdateOne(ctx, bson.D{{"_id", id}}, updateBSON)
		require.NoError(t, err)
		require.Equal(t, stat, res)

		// read it back, and check that it is still close to the current time.
		err = collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&actualBSON)
		require.NoError(t, err)

		AssertEqualDocuments(t, expectedBSON, actualBSON)
		actualY := ConvertDocument(t, actualBSON)
		testutil.CompareAndSetByPathTime(t, actualY, actualDocument, maxDifference, path)
	})
}
