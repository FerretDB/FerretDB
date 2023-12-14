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
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

// TestTailableCursorsBetweenSessions runs multiple cursors in parallel, that should spawn in different sessions,
// and result in error.
func TestTailableCursorsBetweenSessions(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/153")

	s := setup.SetupWithOpts(tt, nil)
	db, ctx := s.Collection.Database(), s.Ctx

	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(10000)
	err := db.CreateCollection(s.Ctx, tt.Name(), opts)
	require.NoError(tt, err)

	collection := db.Collection(tt.Name())

	bsonArr, _ := integration.GenerateDocuments(0, 50)

	_, err = collection.InsertMany(ctx, bsonArr)
	require.NoError(tt, err)

	var passed atomic.Bool

	teststress.StressN(t, 10, func(ready chan<- struct{}, start <-chan struct{}) {
		findCmd := bson.D{
			{"find", collection.Name()},
			{"batchSize", 1},
			{"tailable", true},
		}

		var res bson.D
		err := db.RunCommand(ctx, findCmd).Decode(&res)
		require.NoError(t, err)

		_, cursorID := getFirstBatch(t, res)
		require.NotEmpty(t, cursorID)

		t.Cleanup(func() {
			killCursorCmd := bson.D{
				{"killCursors", collection.Name()},
				{"cursors", bson.A{cursorID}},
			}

			err = db.RunCommand(ctx, killCursorCmd).Decode(&res)
			require.NoError(t, err)
		})

		getMoreCmd := bson.D{
			{"getMore", cursorID},
			{"collection", collection.Name()},
			{"batchSize", 1},
		}

		ready <- struct{}{}
		<-start

		for i := 1; i < 50; i++ {
			err := db.RunCommand(ctx, getMoreCmd).Err()
			if err != nil {
				var ce mongo.CommandError

				require.True(t, errors.As(err, &ce))
				require.Equal(t, int32(50738), ce.Code)
				require.Equal(t, "Location50738", ce.Name)

				passed.Store(true)
				return
			}
		}

		err = db.RunCommand(ctx, getMoreCmd).Err()
		require.NoError(t, err)
	})

	require.True(t, passed.Load(), "Expected cross-session cursor error in at least one goroutine")
}
