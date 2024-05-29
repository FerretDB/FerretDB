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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/util/testutil"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestCursorsKill(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.Strings)

	// does not show up in cursorsAlive or anywhere else
	cursor, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(1))
	require.NoError(t, err)
	require.True(t, cursor.Next(ctx))

	defer cursor.Close(ctx)

	t.Run("Empty", func(t *testing.T) {
		t.Parallel()

		var a bson.D
		err := collection.Database().RunCommand(ctx, bson.D{
			{"killCursors", collection.Name()},
			{"cursors", bson.A{}},
		}).Decode(&a)
		require.NoError(t, err)

		actual := integration.ConvertDocument(t, a)
		actual.Remove("$clusterTime")
		actual.Remove("operationTime")

		expected := integration.ConvertDocument(t, bson.D{
			{"cursorsKilled", bson.A{}},
			{"cursorsNotFound", bson.A{}},
			{"cursorsAlive", bson.A{}},
			{"cursorsUnknown", bson.A{}},
			{"ok", float64(1)},
		})
		testutil.AssertEqual(t, expected, actual)
	})

	t.Run("WrongType", func(t *testing.T) {
		t.Parallel()

		c, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(1))
		require.NoError(t, err)
		require.True(t, c.Next(ctx))
		defer c.Close(ctx)

		var a bson.D
		err = collection.Database().RunCommand(ctx, bson.D{
			{"killCursors", collection.Name()},
			{"cursors", bson.A{c.ID(), int32(100500)}},
		}).Decode(&a)

		expectedErr := mongo.CommandError{
			Code:    14,
			Name:    "TypeMismatch",
			Message: "BSON field 'killCursors.cursors.1' is the wrong type 'int', expected type 'long'",
		}
		integration.AssertEqualCommandError(t, expectedErr, err)

		assert.True(t, c.Next(ctx))
		assert.NoError(t, c.Err())
	})

	t.Run("Found", func(t *testing.T) {
		t.Parallel()

		c, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(1))
		require.NoError(t, err)
		require.True(t, c.Next(ctx))
		defer c.Close(ctx)

		var a bson.D
		err = collection.Database().RunCommand(ctx, bson.D{
			{"killCursors", collection.Name()},
			{"cursors", bson.A{c.ID()}},
		}).Decode(&a)
		require.NoError(t, err)

		actual := integration.ConvertDocument(t, a)
		actual.Remove("$clusterTime")
		actual.Remove("operationTime")

		expected := integration.ConvertDocument(t, bson.D{
			{"cursorsKilled", bson.A{c.ID()}},
			{"cursorsNotFound", bson.A{}},
			{"cursorsAlive", bson.A{}},
			{"cursorsUnknown", bson.A{}},
			{"ok", float64(1)},
		})
		testutil.AssertEqual(t, expected, actual)

		assert.False(t, c.Next(ctx))
		expectedErr := mongo.CommandError{
			Code: 43,
			Name: "CursorNotFound",
		}
		integration.AssertMatchesCommandError(t, expectedErr, c.Err())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		c, err := collection.Find(ctx, bson.D{}, options.Find().SetBatchSize(1))
		require.NoError(t, err)
		require.True(t, c.Next(ctx))
		defer c.Close(ctx)

		var a bson.D
		err = collection.Database().RunCommand(ctx, bson.D{
			{"killCursors", collection.Name()},
			{"cursors", bson.A{c.ID(), int64(100500)}},
		}).Decode(&a)
		require.NoError(t, err)

		actual := integration.ConvertDocument(t, a)
		actual.Remove("$clusterTime")
		actual.Remove("operationTime")

		expected := integration.ConvertDocument(t, bson.D{
			{"cursorsKilled", bson.A{c.ID()}},
			{"cursorsNotFound", bson.A{int64(100500)}},
			{"cursorsAlive", bson.A{}},
			{"cursorsUnknown", bson.A{}},
			{"ok", float64(1)},
		})
		testutil.AssertEqual(t, expected, actual)

		assert.False(t, c.Next(ctx))
		expectedErr := mongo.CommandError{
			Code: 43,
			Name: "CursorNotFound",
		}
		integration.AssertMatchesCommandError(t, expectedErr, c.Err())
	})
}
