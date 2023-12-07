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

// Package cursors contains tests for cursors, tailable cursors, `getMore` command, etc.
//
// It does not contains tests for simple `find`/`aggregate` cases.
package cursors

import (
	"net/url"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCursors(t *testing.T) {
	t.Parallel()

	// use a single connection pool
	s := setup.SetupWithOpts(t, &setup.SetupOpts{
		ExtraOptions: url.Values{
			"minPoolSize": []string{"1"},
			"maxPoolSize": []string{"1"},
		},
	})

	opts := &options.FindOptions{
		BatchSize: pointer.ToInt32(1),
	}

	ctx := s.Ctx
	databaseName := s.Collection.Database().Name()
	collectionName := s.Collection.Name()

	u, err := url.Parse(s.MongoDBURI)
	require.NoError(t, err)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(u.String()))
	require.NoError(t, err)

	collection := client.Database(databaseName).Collection(collectionName)

	arr, _ := integration.GenerateDocuments(1, 5)
	_, err = collection.InsertMany(ctx, arr)
	require.NoError(t, err)

	t.Run("RemoveLastDocument", func(t *testing.T) {
		cur, err := collection.Find(ctx, bson.D{})
		require.NoError(t, err)

		_, err = collection.DeleteOne(ctx, bson.D{{"_id", 4}})
		require.NoError(t, err)

		cur.Next(ctx)
		cur.Next(ctx)
		cur.Next(ctx)
		assert.True(t, cur.TryNext(ctx))
	})

	t.Run("QueryPlanKilledByDrop", func(t *testing.T) {
		cur, err := collection.Find(ctx, bson.D{}, opts)
		require.NoError(t, err)
		cur.Next(ctx)

		err = collection.Database().Drop(ctx)
		require.NoError(t, err)

		res := bson.D{}
		err = cur.All(ctx, &res)

		assert.ErrorContains(t, err, "QueryPlanKilled")
	})

	t.Run("IdleCursorReusedAfterDisconnect", func(t *testing.T) {
		t.Skip("needs work")
		// test that idleCursor can be reused when a client disconnects
		sess, err := client.StartSession()
		require.NoError(t, err)

		cur, err := sess.Client().Database(databaseName).Collection(collectionName).Find(ctx, bson.D{}, opts)
		require.NoError(t, err)

		assert.True(t, cur.TryNext(ctx))

		sessID := sess.ID() // reuse this somehow..
		t.Log(sessID)

		command := bson.D{
			{"aggregate", collectionName},
			{"pipeline", bson.A{
				bson.D{
					{"$currentOp", bson.D{{"idleCursors", true}}},
				},
			}},
			{"cursor", bson.D{}},
		}

		currentOp := sess.Client().Database("admin").RunCommand(ctx, command)
		t.Log(currentOp)

		client.Disconnect(ctx)

		t.Log(cur.TryNext(ctx))
	})
}
