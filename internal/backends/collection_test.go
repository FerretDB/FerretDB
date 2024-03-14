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

			invertedDocs := make([]*types.Document, len(insertDocs))
			copy(invertedDocs, insertDocs)

			slices.Reverse(invertedDocs)

			_, err = coll.InsertAll(ctx, &backends.InsertAllParams{Docs: insertDocs})
			require.NoError(t, err)

			_, err = cappedColl.InsertAll(ctx, &backends.InsertAllParams{Docs: insertDocs})
			require.NoError(t, err)

			t.Run("CappedCollection", func(t *testing.T) {
				t.Parallel()

				sort := must.NotFail(types.NewDocument("$natural", int64(1)))

				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{
					Sort: sort,
				})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				require.Len(t, docs, len(insertDocs))
				testutil.AssertEqualSlices(t, insertDocs, docs)
				assertEqualRecordID(t, insertDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{
					Sort: sort,
				})
				require.NoError(t, err)
				assert.True(t, explainRes.SortPushdown)
			})

			t.Run("CappedCollectionDesc", func(t *testing.T) {
				t.Parallel()

				sort := must.NotFail(types.NewDocument("$natural", int64(-1)))

				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{
					Sort: sort,
				})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				require.Len(t, docs, len(invertedDocs))
				testutil.AssertEqualSlices(t, invertedDocs, docs)

				assertEqualRecordID(t, invertedDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{
					Sort: sort,
				})
				require.NoError(t, err)
				assert.True(t, explainRes.SortPushdown)
			})

			t.Run("CappedCollectionOnlyRecordIDs", func(t *testing.T) {
				t.Parallel()

				sort := must.NotFail(types.NewDocument("$natural", int64(1)))

				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{
					Sort:          sort,
					OnlyRecordIDs: true,
				})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				testutil.AssertEqualSlices(t, []*types.Document{{}, {}, {}}, docs)
				assertEqualRecordID(t, insertDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{
					Sort: sort,
				})
				require.NoError(t, err)
				assert.True(t, explainRes.SortPushdown)
			})

			t.Run("CappedCollectionSortAsc", func(t *testing.T) {
				t.Skip("https://github.com/FerretDB/FerretDB/issues/3181")

				t.Parallel()

				sort := must.NotFail(types.NewDocument("_id", int64(1)))

				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Sort: sort})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				expectedDocs := []*types.Document{insertDocs[2], insertDocs[0], insertDocs[1]}
				testutil.AssertEqualSlices(t, expectedDocs, docs)
				assertEqualRecordID(t, expectedDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{Sort: sort})

				require.NoError(t, err)
				assert.True(t, explainRes.SortPushdown)
			})

			t.Run("CappedCollectionSortDesc", func(t *testing.T) {
				t.Skip("https://github.com/FerretDB/FerretDB/issues/3181")

				t.Parallel()

				sort := must.NotFail(types.NewDocument("_id", int64(-1)))

				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Sort: sort})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				expectedDocs := []*types.Document{insertDocs[1], insertDocs[0], insertDocs[2]}
				testutil.AssertEqualSlices(t, expectedDocs, docs)
				assertEqualRecordID(t, expectedDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{Sort: sort})
				require.NoError(t, err)
				assert.True(t, explainRes.SortPushdown)
			})

			t.Run("CappedCollectionFilter", func(t *testing.T) {
				t.Parallel()

				filter := must.NotFail(types.NewDocument("_id", types.ObjectID{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}))
				sort := must.NotFail(types.NewDocument("$natural", int64(1)))

				queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{
					Filter: filter,
					Sort:   sort,
				})
				require.NoError(t, err)

				docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
				require.NoError(t, err)
				testutil.AssertEqualSlices(t, []*types.Document{filter}, docs)
				expectedDocs := []*types.Document{insertDocs[2]}
				testutil.AssertEqualSlices(t, expectedDocs, docs)
				assertEqualRecordID(t, expectedDocs, docs)

				explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{
					Filter: filter,
					Sort:   sort,
				})
				require.NoError(t, err)
				assert.True(t, explainRes.FilterPushdown)
				assert.True(t, explainRes.SortPushdown)
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
				assert.False(t, explainRes.SortPushdown)
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
			doc2 := must.NotFail(types.NewDocument("_id", int32(2)))
			doc3 := must.NotFail(types.NewDocument("_id", int32(4)))

			_, err = coll.InsertAll(ctx, &backends.InsertAllParams{Docs: []*types.Document{doc1, doc2, doc3}})
			require.NoError(t, err)

			res, err := coll.Query(ctx, nil)
			require.NoError(t, err)

			docs, err := iterator.ConsumeValues[struct{}, *types.Document](res.Iter)
			require.NoError(t, err)
			assertEqualRecordID(t, []*types.Document{doc1, doc2, doc3}, docs)

			params := &backends.DeleteAllParams{
				RecordIDs: []int64{doc1.RecordID(), doc3.RecordID()},
			}
			del, err := coll.DeleteAll(ctx, params)
			require.NoError(t, err)
			require.Equal(t, int32(2), del.Deleted)

			res, err = coll.Query(ctx, nil)
			require.NoError(t, err)

			docs, err = iterator.ConsumeValues[struct{}, *types.Document](res.Iter)
			require.NoError(t, err)
			assertEqualRecordID(t, []*types.Document{doc2}, docs)
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
				require.NotNil(t, collRes)

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

func TestListCollections(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	for name, b := range testBackends(t) {
		name, b := name, b
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// setup 1 DB with 3 collections respectively
			dbName := "testDB"
			collectionNames := []string{"testCollection2", "testCollection1", "testCollection3"}

			testDB, err := b.Database(dbName)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := b.DropDatabase(ctx, &backends.DropDatabaseParams{Name: dbName})
				require.NoError(t, err)
			})

			for _, collectionName := range collectionNames {
				err = testDB.CreateCollection(ctx, &backends.CreateCollectionParams{Name: collectionName})
				require.NoError(t, err)
			}

			// retrieve database details
			dbRes, err := b.ListDatabases(ctx, &backends.ListDatabasesParams{Name: dbName})
			require.NoError(t, err)
			require.Equal(t, 1, len(dbRes.Databases), "expected len 1 , since only 1 db with name testDB")
			require.Equal(t, dbName, dbRes.Databases[0].Name, "expected name testDB")

			db, err := b.Database(dbRes.Databases[0].Name)
			require.NoError(t, err)

			// test ListCollections with 4 different params
			t.Run("ListCollectionWithGivenName", func(t *testing.T) {
				t.Parallel()
				collRes, err := db.ListCollections(ctx, &backends.ListCollectionsParams{Name: collectionNames[2]})
				require.NoError(t, err)
				require.Equal(t, 1, len(collRes.Collections), "expected len 1 , with name testCollection3")
				require.Equal(t, collectionNames[2], collRes.Collections[0].Name, "expected name testCollection3")
			})

			t.Run("ListCollectionWithDummyName", func(t *testing.T) {
				t.Parallel()
				collRes, err := db.ListCollections(ctx, &backends.ListCollectionsParams{Name: "dummy"})
				require.NoError(t, err)
				require.Equal(t, 0, len(collRes.Collections), "expected len 0 since no collection with name dummy")
			})

			t.Run("ListCollectionWithNilParams", func(t *testing.T) {
				t.Parallel()
				collRes, err := db.ListCollections(ctx, nil)
				require.NoError(t, err)
				require.Equal(t, 3, len(collRes.Collections), "expected full list len 3")
				require.Equal(t, collectionNames[1], collRes.Collections[0].Name, "expected name testCollection1")
				require.Equal(t, collectionNames[0], collRes.Collections[1].Name, "expected name testCollection2")
				require.Equal(t, collectionNames[2], collRes.Collections[2].Name, "expected name testCollection3")
			})

			t.Run("ListCollectionWithEmptyParams", func(t *testing.T) {
				t.Parallel()
				var param backends.ListCollectionsParams
				collRes, err := db.ListCollections(ctx, &param)
				require.NoError(t, err)
				require.Equal(t, 3, len(collRes.Collections), "expected full list len 3")
				require.Equal(t, collectionNames[1], collRes.Collections[0].Name, "expected name testCollection1")
				require.Equal(t, collectionNames[0], collRes.Collections[1].Name, "expected name testCollection2")
				require.Equal(t, collectionNames[2], collRes.Collections[2].Name, "expected name testCollection3")
			})
		})
	}
}
