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

	"github.com/stretchr/testify/assert"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

// testCollection creates, tests, and drops an unique collection in existing database.
func testCollection(t *testing.T, ctx context.Context, r *Registry, db *pgxpool.Pool, dbName, collectionName string) {
	t.Helper()

	c, err := r.CollectionGet(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.Nil(t, c)

	created, err := r.CollectionCreate(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.True(t, created)

	created, err = r.CollectionCreate(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.False(t, created)

	c, err = r.CollectionGet(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.Equal(t, collectionName, c.Name)

	list, err := r.CollectionList(ctx, dbName)
	require.NoError(t, err)
	require.Contains(t, list, c)

	q := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES($1)`,
		pgx.Identifier{dbName, c.TableName}.Sanitize(),
		DefaultColumn,
	)
	doc := `{"$s": {"p": {"_id": {"t": "int"}}, "$k": ["_id"]}, "_id": 42}`
	_, err = db.Exec(ctx, q, doc)
	require.NoError(t, err)

	dropped, err := r.CollectionDrop(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.True(t, dropped)

	dropped, err = r.CollectionDrop(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.False(t, dropped)

	c, err = r.CollectionGet(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.Nil(t, c)
}

// createDatabase creates a new provider and registry required for creating a database and
// returns registry, db pool and created database name.
func createDatabase(t *testing.T, ctx context.Context) (r *Registry, db *pgxpool.Pool, dbName string) {
	sp, err := state.NewProvider("")
	require.NoError(t, err)

	u := "postgres://username:password@127.0.0.1:5432/ferretdb?pool_min_conns=1"
	r, err = NewRegistry(u, testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName = testutil.DatabaseName(t)
	db, err = r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, db)

	t.Cleanup(func() {
		_, _ = r.DatabaseDrop(ctx, dbName)
	})

	return r, db, dbName
}

func TestCheckAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	connInfo := conninfo.New()
	ctx := conninfo.Ctx(testutil.Ctx(t), connInfo)

	for name, tc := range map[string]struct {
		uri string
		err string
	}{
		"Auth": {
			uri: "postgres://username:password@127.0.0.1:5432/ferretdb?pool_min_conns=1",
			err: "",
		},
		"NoAuth": {
			uri: "postgres://127.0.0.1:5432/ferretdb?pool_min_conns=1",
			err: "failed to connect to `host=127.0.0.1 user=", // username is the current user running the test
		},
		"NonExistingUser": {
			uri: "postgres://wrong-user:wrong-password@127.0.0.1:5432/ferretdb?pool_min_conns=1",
			err: "failed to connect to `host=127.0.0.1 user=wrong-user database=ferretdb`",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			sp, err := state.NewProvider("")
			require.NoError(t, err)

			r, err := NewRegistry(tc.uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			_, err = r.getPool(ctx)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestCreateDrop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	connInfo := conninfo.New()
	ctx := conninfo.Ctx(testutil.Ctx(t), connInfo)

	r, db, dbName := createDatabase(t, ctx)
	collectionName := testutil.CollectionName(t)
	testCollection(t, ctx, r, db, dbName, collectionName)
}

func TestCreateDropStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	connInfo := conninfo.New()
	ctx := conninfo.Ctx(testutil.Ctx(t), connInfo)

	r, db, dbName := createDatabase(t, ctx)

	t.Cleanup(func() {
		_, _ = r.DatabaseDrop(ctx, dbName)
	})

	var i atomic.Int32

	teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
		collectionName := fmt.Sprintf("collection_%03d", i.Add(1))

		ready <- struct{}{}
		<-start

		testCollection(t, ctx, r, db, dbName, collectionName)
	})
}

func TestCreateSameStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	connInfo := conninfo.New()
	ctx := conninfo.Ctx(testutil.Ctx(t), connInfo)

	r, db, dbName := createDatabase(t, ctx)
	collectionName := testutil.CollectionName(t)

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

		c, err := r.CollectionGet(ctx, dbName, collectionName)
		require.NoError(t, err)
		require.NotNil(t, c)
		require.Equal(t, collectionName, c.Name)

		list, err := r.CollectionList(ctx, dbName)
		require.NoError(t, err)
		require.Contains(t, list, c)

		q := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES($1)",
			pgx.Identifier{dbName, c.TableName}.Sanitize(),
			DefaultColumn,
		)
		doc := fmt.Sprintf(`{"$s": {"p": {"_id": {"t": "int"}}, "$k": ["_id"]}, "_id": %d}`, id)
		_, err = db.Exec(ctx, q, doc)
		require.NoError(t, err)
	})

	require.Equal(t, int32(1), createdTotal.Load())
}

func TestDropSameStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	connInfo := conninfo.New()
	ctx := conninfo.Ctx(testutil.Ctx(t), connInfo)

	r, _, dbName := createDatabase(t, ctx)
	collectionName := testutil.CollectionName(t)
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
}

func TestCreateDropSameStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	connInfo := conninfo.New()
	ctx := conninfo.Ctx(testutil.Ctx(t), connInfo)

	r, _, dbName := createDatabase(t, ctx)
	collectionName := testutil.CollectionName(t)
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

	require.Less(t, int32(1), createdTotal.Load())
	require.Less(t, int32(1), droppedTotal.Load())
}

func TestCheckDatabaseUpdated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	connInfo := conninfo.New()
	ctx := conninfo.Ctx(testutil.Ctx(t), connInfo)

	r, db, dbName := createDatabase(t, ctx)

	var err error

	t.Run("CheckDatabaseCreate", func(t *testing.T) {
		err = r.initCollections(ctx, dbName, db)
		require.NoError(t, err)

		var dbPool *pgxpool.Pool
		dbPool, err = r.DatabaseGetExisting(ctx, dbName)
		require.NoError(t, err)
		require.NotNil(t, dbPool)
	})

	collectionName := testutil.CollectionName(t)
	created, err := r.CollectionCreate(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.True(t, created)

	t.Run("CheckCollectionCreated", func(t *testing.T) {
		metadataCollection, err := r.CollectionGet(ctx, dbName, collectionName)
		require.NoError(t, err)
		require.NotNil(t, metadataCollection)
		require.Equal(t, collectionName, metadataCollection.Name)

		err = r.initCollections(ctx, dbName, db)
		require.NoError(t, err)

		dbCollection, err := r.CollectionGet(ctx, dbName, collectionName)
		require.NoError(t, err)
		require.Equal(t, metadataCollection, dbCollection)
	})

	t.Run("CheckCollectionDropped", func(t *testing.T) {
		dropped, err := r.CollectionDrop(ctx, dbName, collectionName)
		require.NoError(t, err)
		require.True(t, dropped)

		err = r.initCollections(ctx, dbName, db)
		require.NoError(t, err)

		dbCollection, err := r.CollectionGet(ctx, dbName, collectionName)
		require.NoError(t, err)
		require.Nil(t, dbCollection)
	})

	t.Run("CheckDatabaseDropped", func(t *testing.T) {
		dropped, err := r.DatabaseDrop(ctx, dbName)
		require.NoError(t, err)
		require.True(t, dropped)

		err = r.initCollections(ctx, dbName, db)
		require.Error(t, err)
		require.ErrorContains(t, err, "relation \"TestCheckDatabaseUpdated._ferretdb_database_metadata\" does not exist")
	})
}

func TestRenameCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	connInfo := conninfo.New()
	ctx := conninfo.Ctx(testutil.Ctx(t), connInfo)

	r, db, dbName := createDatabase(t, ctx)

	oldCollectionName := testutil.CollectionName(t)
	newCollectionName := "new"

	created, err := r.CollectionCreate(ctx, dbName, oldCollectionName)
	require.NoError(t, err)
	require.True(t, created)

	oldCollection, err := r.CollectionGet(ctx, dbName, oldCollectionName)
	require.NoError(t, err)

	t.Run("CollectionRename", func(t *testing.T) {
		var renamed bool
		renamed, err = r.CollectionRename(ctx, dbName, oldCollectionName, newCollectionName)
		require.NoError(t, err)
		require.True(t, renamed)
	})

	t.Run("CheckCollectionRenamed", func(t *testing.T) {
		err = r.initCollections(ctx, dbName, db)
		require.NoError(t, err)

		expected := &Collection{
			Name:      newCollectionName,
			TableName: oldCollection.TableName,
		}

		actual, err := r.CollectionGet(ctx, dbName, newCollectionName)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	})
}

func TestIndexesCreateDrop(t *testing.T) {
	t.Parallel()

	connInfo := conninfo.New()
	ctx := conninfo.Ctx(testutil.Ctx(t), connInfo)

	r, db, dbName := createDatabase(t, ctx)
	collectionName := testutil.CollectionName(t)

	t.Cleanup(func() {
		_, _ = r.DatabaseDrop(ctx, dbName)
	})

	toCreate := []IndexInfo{
		{
			Name: "index_non_unique",
			Key: []IndexKeyPair{
				{
					Field:      "f1",
					Descending: false,
				},
				{
					Field:      "f2",
					Descending: true,
				},
			},
		},
		{
			Name: "index_unique",
			Key: []IndexKeyPair{
				{
					Field:      "foo",
					Descending: false,
				},
			},
			Unique: true,
		},
		{
			Name: "nested_fields",
			Key: []IndexKeyPair{
				{
					Field: "foo.bar",
				},
				{
					Field:      "foo.baz",
					Descending: true,
				},
			},
		},
	}

	err := r.IndexesCreate(ctx, dbName, collectionName, toCreate)
	require.NoError(t, err)

	collection, err := r.CollectionGet(ctx, dbName, collectionName)
	require.NoError(t, err)

	t.Run("NonUniqueIndex", func(t *testing.T) {
		tableIndexName := "index_non_unique_13064d07"
		var sql string
		err = db.QueryRow(
			ctx,
			"SELECT indexdef FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 AND indexname = $3",
			dbName, collection.TableName, tableIndexName,
		).Scan(&sql)
		require.NoError(t, err)

		expected := fmt.Sprintf(
			`CREATE INDEX %s ON %q.%s USING btree (((_jsonb -> 'f1'::text)), ((_jsonb -> 'f2'::text)) DESC)`,
			tableIndexName, dbName, collection.TableName,
		)
		require.Equal(t, expected, sql)
	})

	t.Run("UniqueIndex", func(t *testing.T) {
		tableIndexName := "index_unique_d29d5863"
		var sql string
		err = db.QueryRow(
			ctx,
			"SELECT indexdef FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 AND indexname = $3",
			dbName, collection.TableName, tableIndexName,
		).Scan(&sql)
		require.NoError(t, err)

		expected := fmt.Sprintf(
			`CREATE UNIQUE INDEX %s ON %q.%s USING btree (((_jsonb -> 'foo'::text)))`,
			tableIndexName, dbName, collection.TableName,
		)
		require.Equal(t, expected, sql)
	})

	t.Run("NestedFields", func(t *testing.T) {
		tableIndexName := "nested_fields_21fa0586"
		var sql string
		err = db.QueryRow(
			ctx,
			"SELECT indexdef FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 AND indexname = $3",
			dbName, collection.TableName, tableIndexName,
		).Scan(&sql)
		require.NoError(t, err)

		expected := fmt.Sprintf(
			`CREATE INDEX %s ON %q.%s USING btree ((((_jsonb -> 'foo'::text) -> 'bar'::text)), (((_jsonb -> 'foo'::text) -> 'baz'::text)) DESC)`,
			tableIndexName, dbName, collection.TableName,
		)
		require.Equal(t, expected, sql)
	})

	t.Run("DefaultIndex", func(t *testing.T) {
		tableIndexName := "_id__67399184"
		var sql string
		err = db.QueryRow(
			ctx,
			"SELECT indexdef FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 AND indexname = $3",
			dbName, collection.TableName, tableIndexName,
		).Scan(&sql)
		require.NoError(t, err)

		expected := fmt.Sprintf(
			`CREATE UNIQUE INDEX %s ON %q.%s USING btree (((_jsonb -> '_id'::text)))`,
			tableIndexName, dbName, collection.TableName,
		)
		require.Equal(t, expected, sql)
	})

	t.Run("CheckSettingsAfterCreation", func(t *testing.T) {
		err = r.initCollections(ctx, dbName, db)
		require.NoError(t, err)

		collection, err = r.CollectionGet(ctx, dbName, collectionName)
		require.NoError(t, err)
		require.Equal(t, 4, len(collection.Settings.Indexes))
	})

	t.Run("DropIndexes", func(t *testing.T) {
		toDrop := []string{"index_non_unique", "nested_fields"}
		err = r.IndexesDrop(ctx, dbName, collectionName, toDrop)
		require.NoError(t, err)

		q := "SELECT count(indexdef) FROM pg_indexes WHERE schemaname = $1 AND tablename = $2"
		row := db.QueryRow(ctx, q, dbName, collection.TableName)

		var count int
		require.NoError(t, row.Scan(&count))
		require.Equal(t, 2, count) // only default index and index_unique should be left
	})

	t.Run("CheckSettingsAfterDrop", func(t *testing.T) {
		err = r.initCollections(ctx, dbName, db)
		require.NoError(t, err)

		collection, err = r.CollectionGet(ctx, dbName, collectionName)
		require.NoError(t, err)
		require.Equal(t, 2, len(collection.Settings.Indexes))
	})

	t.Run("MetadataIndexes", func(t *testing.T) {
		t.Parallel()

		var sql string
		err = db.QueryRow(
			ctx,
			"SELECT indexdef FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 AND indexname = $3",
			dbName, metadataTableName, metadataTableName+"_id_idx",
		).Scan(&sql)
		require.NoError(t, err)

		expected := fmt.Sprintf(
			`CREATE UNIQUE INDEX %s ON %q.%s USING btree (((_jsonb -> '_id'::text)))`,
			metadataTableName+"_id_idx", dbName, metadataTableName,
		)
		require.Equal(t, expected, sql)

		err = db.QueryRow(
			ctx,
			"SELECT indexdef FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 AND indexname = $3",
			dbName, metadataTableName, metadataTableName+"_table_idx",
		).Scan(&sql)
		assert.NoError(t, err)

		expected = fmt.Sprintf(
			`CREATE UNIQUE INDEX %s ON %q.%s USING btree (((_jsonb -> 'table'::text)))`,
			metadataTableName+"_table_idx", dbName, metadataTableName,
		)
		assert.Equal(t, expected, sql)
	})
}
