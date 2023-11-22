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
	"math"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// assertEqualRecordID asserts recordIDs of slices are equal and not zero.
func assertEqualRecordID(t *testing.T, expected, actual []*types.Document) {
	require.Len(t, actual, len(expected))

	for i, doc := range actual {
		assert.Equal(t, expected[i].RecordID(), doc.RecordID())
		assert.NotZero(t, doc.RecordID())
	}
}

func TestCollectionInsertAllQueryExplain(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	for name, b := range testBackends(t) {
		name, b := name, b
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbName := testutil.DatabaseName(t)
			collName, cappedCollName := testutil.CollectionName(t), testutil.CollectionName(t)+"capped"

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
				testutil.AssertEqualSlices(t, insertDocs, docs)
				assertEqualRecordID(t, insertDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, new(backends.ExplainParams))
				require.NoError(t, err)
				assert.True(t, explainRes.UnsafeSortPushdown)
			})

			t.Run("CappedCollectionOnlyRecordIDs", func(t *testing.T) {
				t.Parallel()

				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{OnlyRecordIDs: true})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				testutil.AssertEqualSlices(t, []*types.Document{{}, {}, {}}, docs)
				assertEqualRecordID(t, insertDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, new(backends.ExplainParams))
				require.NoError(t, err)
				assert.True(t, explainRes.UnsafeSortPushdown)
			})

			t.Run("CappedCollectionSortAsc", func(t *testing.T) {
				if name == "sqlite" {
					t.Skip("https://github.com/FerretDB/FerretDB/issues/3181")
				}

				t.Parallel()

				sort := backends.SortField{Key: "_id"}
				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Sort: &sort})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				expectedDocs := []*types.Document{insertDocs[2], insertDocs[0], insertDocs[1]}
				testutil.AssertEqualSlices(t, expectedDocs, docs)
				assertEqualRecordID(t, expectedDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{Sort: &sort})
				require.NoError(t, err)
				assert.True(t, explainRes.UnsafeSortPushdown)
			})

			t.Run("CappedCollectionSortDesc", func(t *testing.T) {
				if name == "sqlite" {
					t.Skip("https://github.com/FerretDB/FerretDB/issues/3181")
				}

				t.Parallel()

				sort := backends.SortField{Key: "_id", Descending: true}
				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Sort: &sort})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				expectedDocs := []*types.Document{insertDocs[1], insertDocs[0], insertDocs[2]}
				testutil.AssertEqualSlices(t, expectedDocs, docs)
				assertEqualRecordID(t, expectedDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{Sort: &sort})
				require.NoError(t, err)
				assert.True(t, explainRes.UnsafeSortPushdown)
			})

			t.Run("CappedCollectionFilter", func(t *testing.T) {
				t.Parallel()

				filter := must.NotFail(types.NewDocument("_id", types.ObjectID{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Filter: filter})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				testutil.AssertEqualSlices(t, []*types.Document{filter}, docs)
				expectedDocs := []*types.Document{insertDocs[2]}
				testutil.AssertEqualSlices(t, expectedDocs, docs)
				assertEqualRecordID(t, expectedDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{Filter: filter})
				require.NoError(t, err)
				assert.True(t, explainRes.QueryPushdown)
				assert.True(t, explainRes.UnsafeSortPushdown)
			})

			t.Run("NonCappedCollectionOnlyRecordID", func(t *testing.T) {
				t.Parallel()

				queryRes, err := coll.Query(ctx, &backends.QueryParams{OnlyRecordIDs: true})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				testutil.AssertEqualSlices(t, []*types.Document{{}, {}, {}}, docs)
				for _, doc := range docs {
					assert.Zero(t, doc.RecordID())
				}

				explainRes, err := coll.Explain(ctx, new(backends.ExplainParams))
				require.NoError(t, err)
				assert.False(t, explainRes.UnsafeSortPushdown)
			})
		})
	}
}

func TestCappedCollectionInsertAllDeleteAll(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	for name, b := range testBackends(t) {
		name, b := name, b
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dbName := testutil.DatabaseName(t)
			collName := testutil.CollectionName(t)

			db, err := b.Database(dbName)
			require.NoError(t, err)

			err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
				Name:       collName,
				CappedSize: 8192,
			})
			require.NoError(t, err)

			coll, err := db.Collection(collName)
			require.NoError(t, err)

			doc1 := must.NotFail(types.NewDocument("_id", int32(1)))
			doc1.SetRecordID(1)
			_, err = coll.InsertAll(ctx, &backends.InsertAllParams{Docs: []*types.Document{doc1}})
			require.NoError(t, err)

			docMax := must.NotFail(types.NewDocument("_id", int32(2)))
			docMax.SetRecordID(math.MaxInt64)
			_, err = coll.InsertAll(ctx, &backends.InsertAllParams{Docs: []*types.Document{docMax}})
			require.NoError(t, err)

			docMaxUint := must.NotFail(types.NewDocument("_id", int32(3)))
			docMaxUint.SetRecordID(math.MaxUint64)
			_, err = coll.InsertAll(ctx, &backends.InsertAllParams{Docs: []*types.Document{docMaxUint}})
			require.Error(t, err)

			docEpochalypse := must.NotFail(types.NewDocument("_id", int32(4)))
			date := time.Date(2038, time.January, 19, 3, 14, 6, 0, time.UTC)
			docEpochalypse.SetRecordID(types.NextTimestamp(date))
			_, err = coll.InsertAll(ctx, &backends.InsertAllParams{Docs: []*types.Document{docEpochalypse}})
			require.NoError(t, err)

			res, err := coll.Query(ctx, nil)
			require.NoError(t, err)

			docs, err := iterator.ConsumeValues[struct{}, *types.Document](res.Iter)
			require.NoError(t, err)
			require.Equal(t, 3, len(docs))

			assert.Equal(t, doc1.RecordID(), docs[0].RecordID())
			assert.Equal(t, docEpochalypse.RecordID(), docs[1].RecordID())
			assert.Equal(t, docMax.RecordID(), docs[2].RecordID())

			params := &backends.DeleteAllParams{
				RecordIDs: []types.Timestamp{docMax.RecordID(), docEpochalypse.RecordID()},
			}
			del, err := coll.DeleteAll(ctx, params)
			require.NoError(t, err)
			require.Equal(t, int32(2), del.Deleted)

			res, err = coll.Query(ctx, nil)
			require.NoError(t, err)

			docs, err = iterator.ConsumeValues[struct{}, *types.Document](res.Iter)
			require.NoError(t, err)
			assertEqualRecordID(t, []*types.Document{doc1}, docs)
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
