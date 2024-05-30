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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestOplogBasic(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t)
	local := coll.Database().Client().Database("local")
	ns := fmt.Sprintf("%s.%s", coll.Database().Name(), coll.Name())
	opts := options.FindOne().SetSort(bson.D{{"$natural", -1}})

	if err := local.CreateCollection(ctx, "oplog.rs", options.CreateCollection().SetCapped(true).SetSizeInBytes(536870912)); err != nil {
		require.Contains(t, err.Error(), "local.oplog.rs already exists")
	}

	expectedKeys := []string{"op", "ns", "o", "ts", "v"}

	expectedV := int64(1)
	if setup.IsMongoDB(t) {
		expectedV = int64(2)
	}

	t.Run("Insert", func(tt *testing.T) {
		tt.Run("SingleDocument", func(tt *testing.T) {
			t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/3556")

			_, err := coll.InsertOne(ctx, bson.D{{"_id", int64(1)}, {"foo", "bar"}})
			require.NoError(t, err)

			var lastOplogEntry bson.D
			err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&lastOplogEntry)
			require.NoError(t, err)

			actual := integration.ConvertDocument(t, lastOplogEntry)
			unsetUnusedOplogFields(actual)
			actualKeys := actual.Keys()
			assert.ElementsMatch(t, expectedKeys, actualKeys)

			expected, err := types.NewDocument(
				"op", "i",
				"ns", ns,
				"o", must.NotFail(types.NewDocument("_id", int64(1), "foo", "bar")),
				"ts", must.NotFail(actual.Get("ts")).(types.Timestamp),
				"v", expectedV,
			)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)
		})

		tt.Run("MultipleDocuments", func(tt *testing.T) {
			t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/3556")

			_, err := coll.InsertMany(ctx, []any{
				bson.D{{"_id", int64(2)}, {"foo2", "bar2"}},
				bson.D{{"_id", int64(3)}, {"foo3", "bar3"}},
			})
			require.NoError(t, err)

			var lastOplogEntry bson.D
			err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&lastOplogEntry)
			require.NoError(t, err)

			actual := integration.ConvertDocument(t, lastOplogEntry)
			unsetUnusedOplogFields(actual)
			actualKeys := actual.Keys()
			assert.ElementsMatch(t, expectedKeys, actualKeys)

			expected, err := types.NewDocument(
				"op", "i",
				"ns", ns,
				"o", must.NotFail(types.NewDocument("_id", int64(3), "foo3", "bar3")),
				"ts", must.NotFail(actual.Get("ts")).(types.Timestamp),
				"v", expectedV,
			)
			require.NoError(t, err)
			assert.Equal(t, expected, actual) // The last oplog entry for multiple inserts only contains the last insert (unlike delete).
		})
	})

	t.Run("Delete", func(tt *testing.T) {
		tt.Run("SingleDocument", func(tt *testing.T) {
			t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/3556")

			_, err := coll.DeleteOne(ctx, bson.D{{"_id", int64(1)}})
			require.NoError(t, err)

			var lastOplogEntry bson.D
			err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&lastOplogEntry)
			require.NoError(t, err)

			actual := integration.ConvertDocument(t, lastOplogEntry)
			unsetUnusedOplogFields(actual)
			actualKeys := actual.Keys()
			assert.ElementsMatch(t, expectedKeys, actualKeys)

			expected, err := types.NewDocument(
				"op", "d",
				"ns", ns,
				"o", must.NotFail(types.NewDocument("_id", int64(1))),
				"ts", must.NotFail(actual.Get("ts")).(types.Timestamp),
				"v", expectedV,
			)
			require.NoError(t, err)
			assert.Equal(t, expected, actual)

			_, err = coll.DeleteOne(ctx, bson.D{{"_id", "non-existent"}})
			require.NoError(t, err)

			var newOplogEntry bson.D
			err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"ns", ns}}, opts).Decode(&newOplogEntry)
			require.NoError(t, err)
			assert.Equal(t, lastOplogEntry, newOplogEntry) // If an entry to delete is not found, expect oplog not to be written.
		})

		tt.Run("MultipleDocuments", func(tt *testing.T) {
			t := setup.FailsForFerretDB(tt, "https://github.com/FerretDB/FerretDB/issues/3556")

			_, err := coll.DeleteMany(ctx, bson.D{{"_id", bson.D{{"$gte", int64(2)}}}})
			require.NoError(t, err)

			var lastOplogEntry bson.D
			err = local.Collection("oplog.rs").FindOne(ctx, bson.D{{"o.applyOps.ns", ns}}, opts).Decode(&lastOplogEntry)
			require.NoError(t, err)

			actual := integration.ConvertDocument(t, lastOplogEntry)
			unsetUnusedOplogFields(actual)
			actualKeys := actual.Keys()
			assert.ElementsMatch(t, expectedKeys, actualKeys)

			applyOps := must.NotFail(must.NotFail(actual.Get("o")).(*types.Document).Get("applyOps")).(*types.Array)
			ui := must.NotFail(must.NotFail(applyOps.Get(0)).(*types.Document).Get("ui")).(types.Binary)
			expected, err := types.NewDocument(
				"op", "c",
				"ns", "admin.$cmd",
				"o", must.NotFail(types.NewDocument(
					"applyOps", must.NotFail(types.NewArray(
						must.NotFail(types.NewDocument(
							"op", "d",
							"ns", ns,
							"ui", ui,
							"o", must.NotFail(types.NewDocument("_id", int64(2))),
						)),
						must.NotFail(types.NewDocument(
							"op", "d",
							"ns", ns,
							"ui", ui,
							"o", must.NotFail(types.NewDocument("_id", int64(3))),
						)),
					)),
				)),
				"ts", must.NotFail(actual.Get("ts")).(types.Timestamp),
				"v", expectedV,
			)
			require.NoError(t, err)
			assert.Equal(t, expected, actual) // The last applyOps oplog entry for multiple deletes contains all the deletes (unlike insert).
		})
	})
}

// unsetUnusedOplogFields removes the fields that are not used in the oplog response.
func unsetUnusedOplogFields(d *types.Document) {
	d.Remove("lsid")
	d.Remove("txnNumber")
	d.Remove("ui")
	d.Remove("o2")
	d.Remove("stmtId")
	d.Remove("t")
	d.Remove("wall")
	d.Remove("prevOpTime")
	d.Remove("_id")
}
