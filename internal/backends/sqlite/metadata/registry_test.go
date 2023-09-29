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

package metadata

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

// testCollection creates, tests, and drops an unique collection in existing database.
func testCollection(t *testing.T, ctx context.Context, r *Registry, db *fsql.DB, dbName, collectionName string) {
	t.Helper()

	c := r.CollectionGet(ctx, dbName, collectionName)
	require.Nil(t, c)

	created, err := r.CollectionCreate(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.True(t, created)

	created, err = r.CollectionCreate(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.False(t, created)

	c = r.CollectionGet(ctx, dbName, collectionName)
	require.NotNil(t, c)
	require.Equal(t, collectionName, c.Name)

	list, err := r.CollectionList(ctx, dbName)
	require.NoError(t, err)
	require.Contains(t, list, c)

	q := fmt.Sprintf("INSERT INTO %q (%s) VALUES(?)", c.TableName, DefaultColumn)
	doc := `{"$s": {"p": {"_id": {"t": "int"}}, "$k": ["_id"]}, "_id": 42}`
	_, err = db.ExecContext(ctx, q, doc)
	require.NoError(t, err)

	dropped, err := r.CollectionDrop(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.True(t, dropped)

	dropped, err = r.CollectionDrop(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.False(t, dropped)

	c = r.CollectionGet(ctx, dbName, collectionName)
	require.Nil(t, c)
}

func TestCreateDrop(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	r, err := NewRegistry("file:./?mode=memory", testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName := testutil.DatabaseName(t)

	db, err := r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, db)

	state := sp.Get()
	require.Equal(t, "3.41.2", state.HandlerVersion)

	t.Cleanup(func() {
		r.DatabaseDrop(ctx, dbName)
	})

	collectionName := testutil.CollectionName(t)

	testCollection(t, ctx, r, db, dbName, collectionName)
}

func TestCreateDropStress(t *testing.T) {
	// Otherwise, the test might fail with "database schema has changed".
	// That error code is SQLITE_SCHEMA (17).
	// See https://www.sqlite.org/rescode.html#schema and https://www.sqlite.org/compile.html#max_schema_retry
	require.Less(t, teststress.NumGoroutines, 50)

	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	for testName, uri := range map[string]string{
		"file":             "file:./",
		"file-immediate":   "file:./?_txlock=immediate",
		"memory":           "file:./?mode=memory",
		"memory-immediate": "file:./?mode=memory&_txlock=immediate",
	} {
		t.Run(testName, func(t *testing.T) {
			r, err := NewRegistry(uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			dbName := "db"
			r.DatabaseDrop(ctx, dbName)

			db, err := r.DatabaseGetOrCreate(ctx, dbName)
			require.NoError(t, err)
			require.NotNil(t, db)

			t.Cleanup(func() {
				r.DatabaseDrop(ctx, dbName)
			})

			var i atomic.Int32

			teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
				collectionName := fmt.Sprintf("collection_%03d", i.Add(1))

				ready <- struct{}{}
				<-start

				testCollection(t, ctx, r, db, dbName, collectionName)
			})

			require.Equal(t, int32(teststress.NumGoroutines), i.Load())
		})
	}
}

func TestCreateSameStress(t *testing.T) {
	// Otherwise, the test might fail with "database schema has changed".
	// That error code is SQLITE_SCHEMA (17).
	// See https://www.sqlite.org/rescode.html#schema and https://www.sqlite.org/compile.html#max_schema_retry
	require.Less(t, teststress.NumGoroutines, 50)

	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	for testName, uri := range map[string]string{
		"file":             "file:./",
		"file-immediate":   "file:./?_txlock=immediate",
		"memory":           "file:./?mode=memory",
		"memory-immediate": "file:./?mode=memory&_txlock=immediate",
	} {
		t.Run(testName, func(t *testing.T) {
			r, err := NewRegistry(uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			dbName := "db"
			r.DatabaseDrop(ctx, dbName)

			db, err := r.DatabaseGetOrCreate(ctx, dbName)
			require.NoError(t, err)
			require.NotNil(t, db)

			t.Cleanup(func() {
				r.DatabaseDrop(ctx, dbName)
			})

			collectionName := "collection"

			var i, createdTotal atomic.Int32

			teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
				id := i.Add(1)

				ready <- struct{}{}
				<-start

				created, err := r.CollectionCreate(ctx, dbName, collectionName)
				require.NoError(t, err)
				if created {
					createdTotal.Add(1)
				}

				created, err = r.CollectionCreate(ctx, dbName, collectionName)
				require.NoError(t, err)
				require.False(t, created)

				c := r.CollectionGet(ctx, dbName, collectionName)
				require.NotNil(t, c)
				require.Equal(t, collectionName, c.Name)

				list, err := r.CollectionList(ctx, dbName)
				require.NoError(t, err)
				require.Contains(t, list, c)

				q := fmt.Sprintf("INSERT INTO %q (%s) VALUES(?)", c.TableName, DefaultColumn)
				doc := fmt.Sprintf(`{"$s": {"p": {"_id": {"t": "int"}}, "$k": ["_id"]}, "_id": %d}`, id)
				_, err = db.ExecContext(ctx, q, doc)
				require.NoError(t, err)
			})

			require.Equal(t, int32(teststress.NumGoroutines), i.Load())
			require.Equal(t, int32(1), createdTotal.Load())
		})
	}
}

func TestDropSameStress(t *testing.T) {
	// Otherwise, the test might fail with "database schema has changed".
	// That error code is SQLITE_SCHEMA (17).
	// See https://www.sqlite.org/rescode.html#schema and https://www.sqlite.org/compile.html#max_schema_retry
	require.Less(t, teststress.NumGoroutines, 50)

	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	for testName, uri := range map[string]string{
		"file":             "file:./",
		"file-immediate":   "file:./?_txlock=immediate",
		"memory":           "file:./?mode=memory",
		"memory-immediate": "file:./?mode=memory&_txlock=immediate",
	} {
		t.Run(testName, func(t *testing.T) {
			r, err := NewRegistry(uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			dbName := "db"
			r.DatabaseDrop(ctx, dbName)

			db, err := r.DatabaseGetOrCreate(ctx, dbName)
			require.NoError(t, err)
			require.NotNil(t, db)

			t.Cleanup(func() {
				r.DatabaseDrop(ctx, dbName)
			})

			collectionName := "collection"

			created, err := r.CollectionCreate(ctx, dbName, collectionName)
			require.NoError(t, err)
			require.True(t, created)

			var droppedTotal atomic.Int32

			teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
				ready <- struct{}{}
				<-start

				dropped, err := r.CollectionDrop(ctx, dbName, collectionName)
				require.NoError(t, err)
				if dropped {
					droppedTotal.Add(1)
				}
			})

			require.Equal(t, int32(1), droppedTotal.Load())
		})
	}
}

func TestCreateDropSameStress(t *testing.T) {
	// Otherwise, the test might fail with "database schema has changed".
	// That error code is SQLITE_SCHEMA (17).
	// See https://www.sqlite.org/rescode.html#schema and https://www.sqlite.org/compile.html#max_schema_retry
	require.Less(t, teststress.NumGoroutines, 50)

	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	for testName, uri := range map[string]string{
		"file":             "file:./",
		"file-immediate":   "file:./?_txlock=immediate",
		"memory":           "file:./?mode=memory",
		"memory-immediate": "file:./?mode=memory&_txlock=immediate",
	} {
		t.Run(testName, func(t *testing.T) {
			r, err := NewRegistry(uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			dbName := "db"
			r.DatabaseDrop(ctx, dbName)

			db, err := r.DatabaseGetOrCreate(ctx, dbName)
			require.NoError(t, err)
			require.NotNil(t, db)

			t.Cleanup(func() {
				r.DatabaseDrop(ctx, dbName)
			})

			collectionName := "collection"

			var i, createdTotal, droppedTotal atomic.Int32

			teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
				id := i.Add(1)

				ready <- struct{}{}
				<-start

				if id%2 == 0 {
					created, err := r.CollectionCreate(ctx, dbName, collectionName)
					require.NoError(t, err)
					if created {
						createdTotal.Add(1)
					}
				} else {
					dropped, err := r.CollectionDrop(ctx, dbName, collectionName)
					require.NoError(t, err)
					if dropped {
						droppedTotal.Add(1)
					}
				}
			})

			require.Equal(t, int32(teststress.NumGoroutines), i.Load())
			require.Less(t, int32(1), createdTotal.Load())
			require.Less(t, int32(1), droppedTotal.Load())
		})
	}
}

func TestIndexesCreateDrop(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	r, err := NewRegistry("file:./?mode=memory", testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName := testutil.DatabaseName(t)

	db, err := r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, db)

	t.Cleanup(func() {
		r.DatabaseDrop(ctx, dbName)
	})

	collectionName := testutil.CollectionName(t)

	toCreate := []IndexInfo{{
		Name: "index_non_unique",
		Key: []IndexKeyPair{{
			Field:      "f1",
			Descending: false,
		}, {
			Field:      "f2",
			Descending: true,
		}},
	}, {
		Name: "index_unique",
		Key: []IndexKeyPair{{
			Field:      "foo",
			Descending: false,
		}},
		Unique: true,
	}}

	err = r.IndexesCreate(ctx, dbName, collectionName, toCreate)
	require.NoError(t, err)

	collection := r.CollectionGet(ctx, dbName, collectionName)

	t.Run("NonUniqueIndex", func(t *testing.T) {
		indexName := collection.TableName + "_index_non_unique"
		q := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type = 'index' AND name = '%s'", indexName)
		row := db.QueryRowContext(ctx, q)

		var sql string
		require.NoError(t, row.Scan(&sql))

		expected := fmt.Sprintf(
			`CREATE INDEX "%s" ON "%s" (_ferretdb_sjson->'$.f1', _ferretdb_sjson->'$.f2' DESC)`,
			indexName, collection.TableName,
		)
		require.Equal(t, expected, sql)
	})

	t.Run("UniqueIndex", func(t *testing.T) {
		indexName := collection.TableName + "_index_unique"
		q := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type = 'index' AND name = '%s'", indexName)
		row := db.QueryRowContext(ctx, q)

		var sql string
		require.NoError(t, row.Scan(&sql))

		expected := fmt.Sprintf(
			`CREATE UNIQUE INDEX "%s" ON "%s" (_ferretdb_sjson->'$.foo')`,
			indexName, collection.TableName,
		)
		require.Equal(t, expected, sql)
	})

	t.Run("DefaultIndex", func(t *testing.T) {
		indexName := collection.TableName + "__id_"
		q := "SELECT sql FROM sqlite_master WHERE type = 'index' AND name = ?"
		row := db.QueryRowContext(ctx, q, indexName)

		var sql string
		require.NoError(t, row.Scan(&sql))

		expected := fmt.Sprintf(
			`CREATE UNIQUE INDEX "%s" ON "%s" (_ferretdb_sjson->'$._id')`,
			indexName, collection.TableName,
		)
		require.Equal(t, expected, sql)
	})

	t.Run("CheckSettingsAfterCreation", func(t *testing.T) {
		err = r.initCollections(ctx, dbName, db)
		require.NoError(t, err)

		collection = r.CollectionGet(ctx, dbName, collectionName)
		require.Equal(t, 3, len(collection.Settings.Indexes))
	})

	t.Run("DropIndexes", func(t *testing.T) {
		toDrop := []string{"index_non_unique", "index_unique"}
		err = r.IndexesDrop(ctx, dbName, collectionName, toDrop)
		require.NoError(t, err)

		q := "SELECT count(*) FROM sqlite_master WHERE type = 'index' AND tbl_name = ?"
		row := db.QueryRowContext(ctx, q, collection.TableName)

		var count int
		require.NoError(t, row.Scan(&count))
		require.Equal(t, 1, count) // only default index
	})

	t.Run("CheckSettingsAfterDrop", func(t *testing.T) {
		err = r.initCollections(ctx, dbName, db)
		require.NoError(t, err)

		collection = r.CollectionGet(ctx, dbName, collectionName)
		require.Equal(t, 1, len(collection.Settings.Indexes))
	})
}
