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

package backends_test // to avoid import cycle

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCollectionInsertAllQueryExplain(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	for name, b := range testBackends(t) {
		name, b := name, b
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbName := testutil.DatabaseName(t)
			collName, cappedCollName := testutil.CollectionName(t), testutil.CollectionName(t)+"capped"
			cleanupDatabase(t, ctx, b, dbName)

			db, err := b.Database(dbName)
			require.NoError(t, err)

			coll, err := db.Collection(collName)
			require.NoError(t, err)

			err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
				Name:       cappedCollName,
				CappedSize: 8192,
			})
			require.NoError(t, err)

			cappedColl, err := db.Collection(cappedCollName)
			require.NoError(t, err)

			insertDocs := []*types.Document{
				must.NotFail(types.NewDocument("_id", types.ObjectID{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})),
				must.NotFail(types.NewDocument("_id", types.ObjectID{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})),
				must.NotFail(types.NewDocument("_id", types.ObjectID{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})),
			}

			_, err = coll.InsertAll(ctx, &backends.InsertAllParams{Docs: insertDocs})
			require.NoError(t, err)

			_, err = cappedColl.InsertAll(ctx, &backends.InsertAllParams{Docs: insertDocs})
			require.NoError(t, err)

			t.Run("CappedCollection", func(t *testing.T) {
				t.Parallel()

				queryRes, err := cappedColl.Query(ctx, new(backends.QueryParams))
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				require.Len(t, docs, len(insertDocs))

				for i, doc := range docs {
					assert.NotEmpty(t, doc.RecordID())
					assert.Equal(t, insertDocs[i].Keys(), doc.Keys())
					assert.Equal(t, insertDocs[i].Values(), doc.Values())
				}

				explainRes, err := cappedColl.Explain(ctx, new(backends.ExplainParams))
				require.NoError(t, err)
				assert.True(t, explainRes.SortPushdown)
			})

			t.Run("CappedCollectionOnlyRecordIDs", func(t *testing.T) {
				t.Parallel()

				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{OnlyRecordIDs: true})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				require.Len(t, docs, 1)
				assert.NotEmpty(t, docs[0].RecordID())
				assert.Empty(t, docs[0].Keys())

				explainRes, err := cappedColl.Explain(ctx, new(backends.ExplainParams))
				require.NoError(t, err)
				assert.True(t, explainRes.SortPushdown)
			})

			t.Run("CappedCollectionSortAsc", func(t *testing.T) {
				t.Skip("https://github.com/FerretDB/FerretDB/issues/3181")

				t.Parallel()

				sort := backends.SortField{Key: "_id"}
				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Sort: &sort})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)

				require.Len(t, docs, len(insertDocs))
				assert.NotEmpty(t, docs[0].RecordID())
				assert.Equal(t, insertDocs[2].Keys(), docs[0].Keys())
				assert.Equal(t, insertDocs[2].Values(), docs[0].Values())

				assert.NotEmpty(t, docs[1].RecordID())
				assert.Equal(t, insertDocs[0].Keys(), docs[1].Keys())
				assert.Equal(t, insertDocs[0].Values(), docs[1].Values())

				assert.NotEmpty(t, docs[1].RecordID())
				assert.Equal(t, insertDocs[1].Keys(), docs[2].Keys())
				assert.Equal(t, insertDocs[1].Values(), docs[2].Values())

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{Sort: &sort})
				require.NoError(t, err)
				assert.True(t, explainRes.SortPushdown)
			})

			t.Run("CappedCollectionSortDesc", func(t *testing.T) {
				t.Skip("https://github.com/FerretDB/FerretDB/issues/3181")

				t.Parallel()

				sort := backends.SortField{Key: "_id", Descending: true}
				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Sort: &sort})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)

				require.Len(t, docs, len(insertDocs))
				assert.NotEmpty(t, docs[0].RecordID())
				assert.Equal(t, insertDocs[1].Keys(), docs[0].Keys())
				assert.Equal(t, insertDocs[1].Values(), docs[0].Values())

				assert.NotEmpty(t, docs[1].RecordID())
				assert.Equal(t, insertDocs[0].Keys(), docs[1].Keys())
				assert.Equal(t, insertDocs[0].Values(), docs[1].Values())

				assert.NotEmpty(t, docs[1].RecordID())
				assert.Equal(t, insertDocs[2].Keys(), docs[2].Keys())
				assert.Equal(t, insertDocs[2].Values(), docs[2].Values())

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{Sort: &sort})
				require.NoError(t, err)
				assert.True(t, explainRes.SortPushdown)
			})

			t.Run("CappedCollectionFilter", func(t *testing.T) {
				t.Parallel()

				filter := must.NotFail(types.NewDocument("_id", types.ObjectID{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Filter: filter})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				require.Len(t, docs, 1)
				assert.NotEmpty(t, docs[0].RecordID())
				assert.Empty(t, docs[0].Keys())

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{Filter: filter})
				require.NoError(t, err)
				assert.True(t, explainRes.QueryPushdown)
				assert.True(t, explainRes.SortPushdown)
			})

			t.Run("NonCappedCollectionOnlyRecordID", func(t *testing.T) {
				t.Parallel()

				queryRes, err := coll.Query(ctx, &backends.QueryParams{OnlyRecordIDs: true})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				require.Len(t, docs, len(insertDocs))

				for _, doc := range docs {
					assert.Empty(t, doc.RecordID())
				}

				explainRes, err := coll.Explain(ctx, new(backends.ExplainParams))
				require.NoError(t, err)
				assert.False(t, explainRes.SortPushdown)
			})
		})
	}
}

func TestCollectionUpdateAll(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	for name, b := range testBackends(t) {
		name, b := name, b
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			t.Run("DatabaseDoesNotExist", func(t *testing.T) {
				t.Parallel()

				dbName, collName := testutil.DatabaseName(t), testutil.CollectionName(t)
				cleanupDatabase(t, ctx, b, dbName)

				db, err := b.Database(dbName)
				require.NoError(t, err)

				coll, err := db.Collection(collName)
				require.NoError(t, err)

				updateRes, err := coll.UpdateAll(ctx, &backends.UpdateAllParams{
					Docs: []*types.Document{
						must.NotFail(types.NewDocument("_id", int32(42))),
					},
				})
				assert.NoError(t, err)
				require.NotNil(t, updateRes)
				assert.Zero(t, updateRes.Updated)

				dbRes, err := b.ListDatabases(ctx, nil)
				require.NoError(t, err)
				require.NotNil(t, dbRes)

				present := slices.ContainsFunc(dbRes.Databases, func(di backends.DatabaseInfo) bool {
					return di.Name == dbName
				})
				assert.False(t, present)

				collRes, err := db.ListCollections(ctx, nil)
				require.NoError(t, err)
				require.NotNil(t, dbRes)

				present = slices.ContainsFunc(collRes.Collections, func(ci backends.CollectionInfo) bool {
					return ci.Name == collName
				})
				assert.False(t, present)
			})

			t.Run("CollectionDoesNotExist", func(t *testing.T) {
				t.Parallel()

				dbName, collName := testutil.DatabaseName(t), testutil.CollectionName(t)
				otherCollName := collName + "_other"
				cleanupDatabase(t, ctx, b, dbName)

				db, err := b.Database(dbName)
				require.NoError(t, err)

				// to create database
				err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
					Name: otherCollName,
				})
				require.NoError(t, err)

				coll, err := db.Collection(collName)
				require.NoError(t, err)

				updateRes, err := coll.UpdateAll(ctx, &backends.UpdateAllParams{
					Docs: []*types.Document{
						must.NotFail(types.NewDocument("_id", int32(42))),
					},
				})
				assert.NoError(t, err)
				require.NotNil(t, updateRes)
				assert.Zero(t, updateRes.Updated)

				dbRes, err := b.ListDatabases(ctx, nil)
				require.NoError(t, err)
				require.NotNil(t, dbRes)

				present := slices.ContainsFunc(dbRes.Databases, func(di backends.DatabaseInfo) bool {
					return di.Name == dbName
				})
				assert.True(t, present)

				collRes, err := db.ListCollections(ctx, nil)
				require.NoError(t, err)
				require.NotNil(t, dbRes)

				present = slices.ContainsFunc(collRes.Collections, func(ci backends.CollectionInfo) bool {
					return ci.Name == collName
				})
				assert.False(t, present)

				present = slices.ContainsFunc(collRes.Collections, func(ci backends.CollectionInfo) bool {
					return ci.Name == otherCollName
				})
				assert.True(t, present)
			})
		})
	}
}

func TestCollectionStats(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	for name, b := range testBackends(t) {
		name, b := name, b
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			t.Run("DatabaseDoesNotExist", func(t *testing.T) {
				t.Parallel()

				dbName, collName := testutil.DatabaseName(t), testutil.CollectionName(t)
				cleanupDatabase(t, ctx, b, dbName)

				db, err := b.Database(dbName)
				require.NoError(t, err)

				coll, err := db.Collection(collName)
				require.NoError(t, err)

				_, err = coll.Stats(ctx, nil)
				assertErrorCode(t, err, backends.ErrorCodeCollectionDoesNotExist)
			})

			t.Run("CollectionDoesNotExist", func(t *testing.T) {
				t.Parallel()

				dbName, collName := testutil.DatabaseName(t), testutil.CollectionName(t)
				otherCollName := collName + "_other"
				cleanupDatabase(t, ctx, b, dbName)

				db, err := b.Database(dbName)
				require.NoError(t, err)

				// to create database
				err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
					Name: otherCollName,
				})
				require.NoError(t, err)

				coll, err := db.Collection(collName)
				require.NoError(t, err)

				_, err = coll.Stats(ctx, nil)
				assertErrorCode(t, err, backends.ErrorCodeCollectionDoesNotExist)
			})

			t.Run("Stats", func(t *testing.T) {
				dbName := testutil.DatabaseName(t)
				cleanupDatabase(t, ctx, b, dbName)

				db, err := b.Database(dbName)
				require.NoError(t, err)

				var c backends.Collection
				cNames := []string{"collectionOne", "collectionTwo"}
				for _, cName := range cNames {
					err = db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: cName})
					require.NoError(t, err)

					c, err = db.Collection(cName)
					require.NoError(t, err)

					_, err = c.InsertAll(ctx, &backends.InsertAllParams{
						Docs: []*types.Document{must.NotFail(types.NewDocument("_id", types.NewObjectID()))},
					})
					require.NoError(t, err)
				}

				dbStatsRes, err := db.Stats(ctx, &backends.DatabaseStatsParams{
					Refresh: true,
				})
				require.NoError(t, err)
				res, err := c.Stats(ctx, &backends.CollectionStatsParams{
					Refresh: true,
				})
				require.NoError(t, err)
				require.NotZero(t, res.SizeTotal)
				require.Less(t, res.SizeTotal, dbStatsRes.SizeTotal)
				require.NotZero(t, res.SizeCollection)
				require.Less(t, res.SizeCollection, dbStatsRes.SizeCollections)
				require.Equal(t, res.CountDocuments, int64(1))
				require.NotZero(t, res.SizeIndexes)
			})
		})
	}
}

func TestCollectionCompact(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	for name, b := range testBackends(t) {
		name, b := name, b
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			t.Run("DatabaseDoesNotExist", func(t *testing.T) {
				t.Parallel()

				dbName, collName := testutil.DatabaseName(t), testutil.CollectionName(t)
				cleanupDatabase(t, ctx, b, dbName)

				db, err := b.Database(dbName)
				require.NoError(t, err)

				coll, err := db.Collection(collName)
				require.NoError(t, err)

				_, err = coll.Compact(ctx, nil)
				assertErrorCode(t, err, backends.ErrorCodeDatabaseDoesNotExist)
			})

			t.Run("CollectionDoesNotExist", func(t *testing.T) {
				t.Parallel()

				dbName, collName := testutil.DatabaseName(t), testutil.CollectionName(t)
				otherCollName := collName + "_other"
				cleanupDatabase(t, ctx, b, dbName)

				db, err := b.Database(dbName)
				require.NoError(t, err)

				// to create database
				err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
					Name: otherCollName,
				})
				require.NoError(t, err)

				coll, err := db.Collection(collName)
				require.NoError(t, err)

				_, err = coll.Compact(ctx, nil)
				assertErrorCode(t, err, backends.ErrorCodeCollectionDoesNotExist)
			})
		})
	}
}
