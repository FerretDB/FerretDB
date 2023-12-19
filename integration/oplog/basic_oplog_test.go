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

package oplog

import (
	"fmt"
	"testing"
	"time"

	"github.com/FerretDB/FerretDB/internal/util/must"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/types"

	"github.com/FerretDB/FerretDB/integration"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestOplogBasic(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t)
	local := coll.Database().Client().Database("local")
	ns := fmt.Sprintf("%s.%s", coll.Database().Name(), coll.Name())
	opts := options.FindOne().SetSort(bson.D{{"$natural", -1}})

	// This test uses subtests to group test cases, but subtests can't be run in parallel as we need to ensure oplog order.
	t.Run("Insert", func(t *testing.T) {
		_, err := coll.InsertOne(ctx, bson.D{{"_id", int64(1)}, {"foo", "bar"}})
		require.NoError(t, err)

		var lastOplogEntry bson.D
		err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&lastOplogEntry)
		require.NoError(t, err)

		expectedKeys := []string{"lsid", "txnNumber", "op", "ns", "ui", "o", "o2", "stmtId", "ts", "t", "v", "wall", "prevOpTime"}

		actual := integration.ConvertDocument(t, lastOplogEntry)
		actualKeys := actual.Keys()

		assert.ElementsMatch(t, expectedKeys, actualKeys)

		// Exact values might vary, so we just check types.
		require.IsType(t, &types.Document{}, must.NotFail(actual.Get("lsid")))
		lsid := must.NotFail(actual.Get("lsid")).(*types.Document)
		assert.IsType(t, types.Binary{}, must.NotFail(lsid.Get("id")))
		assert.IsType(t, types.Binary{}, must.NotFail(lsid.Get("uid")))
		assert.IsType(t, int64(0), must.NotFail(actual.Get("txnNumber")))
		assert.IsType(t, int32(0), must.NotFail(actual.Get("stmtId")))
		assert.IsType(t, types.Timestamp(0), must.NotFail(actual.Get("ts")))
		assert.IsType(t, int64(0), must.NotFail(actual.Get("t")))
		assert.IsType(t, time.Time{}, must.NotFail(actual.Get("wall")))
		assert.IsType(t, &types.Document{}, must.NotFail(actual.Get("prevOpTime")))
		prevOpsTime := must.NotFail(actual.Get("prevOpTime")).(*types.Document)
		assert.IsType(t, types.Timestamp(0), must.NotFail(prevOpsTime.Get("ts")))
		assert.IsType(t, int64(0), must.NotFail(prevOpsTime.Get("t")))

		actual.Remove("lsid")
		actual.Remove("txnNumber")  // transaction number
		actual.Remove("ui")         // user ID
		actual.Remove("stmtId")     // statement ID within transaction
		actual.Remove("ts")         // timestamp
		actual.Remove("t")          // term
		actual.Remove("wall")       // wall clock time
		actual.Remove("prevOpTime") // previous operation time

		// Exact values are known, so we check them.
		expected, err := types.NewDocument(
			"op", "i", // operation - i, u, d, n, c
			"ns", ns,
			"o", must.NotFail(types.NewDocument("_id", int64(1), "foo", "bar")),
			"o2", must.NotFail(types.NewDocument("_id", int64(1))),
			"v", int64(2), // protocol version
		)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Update", func(t *testing.T) {
		_, err := coll.UpdateOne(ctx, bson.D{{"_id", int64(1)}}, bson.D{{"$set", bson.D{{"fiz", "baz"}}}})
		require.NoError(t, err)

		var lastOplogEntry bson.D
		err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&lastOplogEntry)
		require.NoError(t, err)

		expectedKeys := []string{"lsid", "txnNumber", "op", "ns", "ui", "o", "o2", "stmtId", "ts", "t", "v", "wall", "prevOpTime"}

		actual := integration.ConvertDocument(t, lastOplogEntry)
		actualKeys := actual.Keys()

		assert.ElementsMatch(t, expectedKeys, actualKeys)

		// Exact values might vary, so we just check types.
		require.IsType(t, &types.Document{}, must.NotFail(actual.Get("lsid")))
		lsid := must.NotFail(actual.Get("lsid")).(*types.Document)
		assert.IsType(t, types.Binary{}, must.NotFail(lsid.Get("id")))
		assert.IsType(t, types.Binary{}, must.NotFail(lsid.Get("uid")))
		assert.IsType(t, int64(0), must.NotFail(actual.Get("txnNumber")))
		assert.IsType(t, int32(0), must.NotFail(actual.Get("stmtId")))
		assert.IsType(t, types.Timestamp(0), must.NotFail(actual.Get("ts")))
		assert.IsType(t, int64(0), must.NotFail(actual.Get("t")))
		assert.IsType(t, time.Time{}, must.NotFail(actual.Get("wall")))
		assert.IsType(t, &types.Document{}, must.NotFail(actual.Get("prevOpTime")))
		prevOpsTime := must.NotFail(actual.Get("prevOpTime")).(*types.Document)
		assert.IsType(t, types.Timestamp(0), must.NotFail(prevOpsTime.Get("ts")))
		assert.IsType(t, int64(0), must.NotFail(prevOpsTime.Get("t")))

		actual.Remove("lsid")
		actual.Remove("txnNumber")  // transaction number
		actual.Remove("ui")         // user ID
		actual.Remove("stmtId")     // statement ID within transaction
		actual.Remove("ts")         // timestamp
		actual.Remove("t")          // term
		actual.Remove("wall")       // wall clock time
		actual.Remove("prevOpTime") // previous operation time

		// Exact values are known, so we check them.
		expected, err := types.NewDocument(
			"op", "u", // operation - i, u, d, n, c
			"ns", ns,
			"o", must.NotFail(types.NewDocument(
				"$v", int32(2),
				"diff", must.NotFail(types.NewDocument("i", must.NotFail(types.NewDocument("fiz", "baz")))),
			)),
			"o2", must.NotFail(types.NewDocument("_id", int64(1))),
			"v", int64(2), // protocol version
		)

		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		t.Run("UpdateField", func(t *testing.T) {
			_, err = coll.UpdateOne(ctx, bson.D{{"_id", int64(1)}}, bson.D{{"$set", bson.D{{"foo", "moo"}}}})

			require.NoError(t, err)

			err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&lastOplogEntry)
			require.NoError(t, err)

			expectedKeys = []string{"lsid", "txnNumber", "op", "ns", "ui", "o", "o2", "stmtId", "ts", "t", "v", "wall", "prevOpTime"}

			actual = integration.ConvertDocument(t, lastOplogEntry)
			actualKeys = actual.Keys()

			assert.ElementsMatch(t, expectedKeys, actualKeys)

			// Exact values might vary, so we just check types.
			require.IsType(t, &types.Document{}, must.NotFail(actual.Get("lsid")))
			lsid = must.NotFail(actual.Get("lsid")).(*types.Document)
			assert.IsType(t, types.Binary{}, must.NotFail(lsid.Get("id")))
			assert.IsType(t, types.Binary{}, must.NotFail(lsid.Get("uid")))
			assert.IsType(t, int64(0), must.NotFail(actual.Get("txnNumber")))
			assert.IsType(t, int32(0), must.NotFail(actual.Get("stmtId")))
			assert.IsType(t, types.Timestamp(0), must.NotFail(actual.Get("ts")))
			assert.IsType(t, int64(0), must.NotFail(actual.Get("t")))
			assert.IsType(t, time.Time{}, must.NotFail(actual.Get("wall")))
			assert.IsType(t, &types.Document{}, must.NotFail(actual.Get("prevOpTime")))
			prevOpsTime = must.NotFail(actual.Get("prevOpTime")).(*types.Document)
			assert.IsType(t, types.Timestamp(0), must.NotFail(prevOpsTime.Get("ts")))
			assert.IsType(t, int64(0), must.NotFail(prevOpsTime.Get("t")))

			actual.Remove("lsid")
			actual.Remove("txnNumber")  // transaction number
			actual.Remove("ui")         // user ID
			actual.Remove("stmtId")     // statement ID within transaction
			actual.Remove("ts")         // timestamp
			actual.Remove("t")          // term
			actual.Remove("wall")       // wall clock time
			actual.Remove("prevOpTime") // previous operation time

			// Exact values are known, so we check them.
			expected, err = types.NewDocument(
				"op", "u", // operation - i, u, d, n, c
				"ns", ns,
				"o", must.NotFail(types.NewDocument(
					"$v", int32(2),
					"diff", must.NotFail(types.NewDocument("u", must.NotFail(types.NewDocument("foo", "moo")))),
				)),
				"o2", must.NotFail(types.NewDocument("_id", int64(1))),
				"v", int64(2), // protocol version
			)

			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})

		t.Run("UnsetField", func(t *testing.T) {
			_, err = coll.UpdateOne(ctx, bson.D{{"_id", int64(1)}}, bson.D{{"$unset", bson.D{{"foo", ""}}}})

			require.NoError(t, err)

			err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&lastOplogEntry)
			require.NoError(t, err)

			expectedKeys = []string{"lsid", "txnNumber", "op", "ns", "ui", "o", "o2", "stmtId", "ts", "t", "v", "wall", "prevOpTime"}

			actual = integration.ConvertDocument(t, lastOplogEntry)
			actualKeys = actual.Keys()

			assert.ElementsMatch(t, expectedKeys, actualKeys)

			// Exact values might vary, so we just check types.
			require.IsType(t, &types.Document{}, must.NotFail(actual.Get("lsid")))
			lsid = must.NotFail(actual.Get("lsid")).(*types.Document)
			assert.IsType(t, types.Binary{}, must.NotFail(lsid.Get("id")))
			assert.IsType(t, types.Binary{}, must.NotFail(lsid.Get("uid")))
			assert.IsType(t, int64(0), must.NotFail(actual.Get("txnNumber")))
			assert.IsType(t, int32(0), must.NotFail(actual.Get("stmtId")))
			assert.IsType(t, types.Timestamp(0), must.NotFail(actual.Get("ts")))
			assert.IsType(t, int64(0), must.NotFail(actual.Get("t")))
			assert.IsType(t, time.Time{}, must.NotFail(actual.Get("wall")))
			assert.IsType(t, &types.Document{}, must.NotFail(actual.Get("prevOpTime")))
			prevOpsTime = must.NotFail(actual.Get("prevOpTime")).(*types.Document)
			assert.IsType(t, types.Timestamp(0), must.NotFail(prevOpsTime.Get("ts")))
			assert.IsType(t, int64(0), must.NotFail(prevOpsTime.Get("t")))

			actual.Remove("lsid")
			actual.Remove("txnNumber")  // transaction number
			actual.Remove("ui")         // user ID
			actual.Remove("stmtId")     // statement ID within transaction
			actual.Remove("ts")         // timestamp
			actual.Remove("t")          // term
			actual.Remove("wall")       // wall clock time
			actual.Remove("prevOpTime") // previous operation time

			// Exact values are known, so we check them.
			expected, err = types.NewDocument(
				"op", "u", // operation - i, u, d, n, c
				"ns", ns,
				"o", must.NotFail(types.NewDocument(
					"$v", int32(2),
					"diff", must.NotFail(types.NewDocument("d", must.NotFail(types.NewDocument("foo", false)))),
				)),
				"o2", must.NotFail(types.NewDocument("_id", int64(1))),
				"v", int64(2), // protocol version
			)

			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		_, err := coll.DeleteOne(ctx, bson.D{{"_id", int64(1)}})
		require.NoError(t, err)

		var lastOplogEntry bson.D
		err = local.Collection("oplog.rs").FindOne(ctx, bson.D{}, opts).Decode(&lastOplogEntry)
		require.NoError(t, err)

		actual := integration.ConvertDocument(t, lastOplogEntry)
		actualKeys := actual.Keys()

		expectedKeys := []string{"lsid", "txnNumber", "op", "ns", "ui", "o", "stmtId", "ts", "t", "v", "wall", "prevOpTime"}

		assert.ElementsMatch(t, expectedKeys, actualKeys)

		// Exact values might vary, so we just check types.
		require.IsType(t, &types.Document{}, must.NotFail(actual.Get("lsid")))
		lsid := must.NotFail(actual.Get("lsid")).(*types.Document)
		assert.IsType(t, types.Binary{}, must.NotFail(lsid.Get("id")))
		assert.IsType(t, types.Binary{}, must.NotFail(lsid.Get("uid")))
		assert.IsType(t, int64(0), must.NotFail(actual.Get("txnNumber")))
		assert.IsType(t, types.Timestamp(0), must.NotFail(actual.Get("ts")))
		assert.IsType(t, time.Time{}, must.NotFail(actual.Get("wall")))
		require.IsType(t, &types.Document{}, must.NotFail(actual.Get("prevOpTime")))
		prevOpsTime := must.NotFail(actual.Get("prevOpTime")).(*types.Document)
		assert.IsType(t, types.Timestamp(0), must.NotFail(prevOpsTime.Get("ts")))
		assert.IsType(t, int64(0), must.NotFail(prevOpsTime.Get("t")))

		actual.Remove("lsid")
		actual.Remove("txnNumber")  // transaction number
		actual.Remove("ui")         // user ID
		actual.Remove("ts")         // timestamp
		actual.Remove("wall")       // wall clock time
		actual.Remove("prevOpTime") // previous operation time

		// Exact values are known, so we check them.
		expected, err := types.NewDocument(
			"op", "d", // operation - i, u, d, n, c
			"ns", ns,
			"o", must.NotFail(types.NewDocument("_id", int64(1))),
			"stmtId", int32(0), // statement ID within transaction
			"t", int64(1), // term
			"v", int64(2), // protocol version
		)

		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		// Attempt to delete a non-existent entry, expect oplog not to be written.
		_, err = coll.DeleteOne(ctx, bson.D{{"_id", "non-existent"}})
		require.NoError(t, err)

		var newOplogEntry bson.D
		err = local.Collection("oplog.rs").FindOne(ctx, bson.D{}, opts).Decode(&newOplogEntry)
		require.NoError(t, err)
		assert.Equal(t, lastOplogEntry, newOplogEntry)
	})
}
