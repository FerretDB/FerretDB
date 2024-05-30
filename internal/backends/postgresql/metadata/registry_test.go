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
	"os/user"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

// testCollection creates, tests, and drops a unique collection in the existing database.
func testCollection(t *testing.T, ctx context.Context, r *Registry, p *pgxpool.Pool, dbName, collectionName string) {
	t.Helper()

	c, err := r.CollectionGet(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.Nil(t, c)

	created, err := r.CollectionCreate(ctx, &CollectionCreateParams{DBName: dbName, Name: collectionName})
	require.NoError(t, err)
	require.True(t, created)

	created, err = r.CollectionCreate(ctx, &CollectionCreateParams{DBName: dbName, Name: collectionName})
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
	_, err = p.Exec(ctx, q, doc)
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
func createDatabase(t *testing.T, ctx context.Context) (*Registry, *pgxpool.Pool, string) {
	t.Helper()

	u := testutil.TestPostgreSQLURI(t, ctx, "")

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	r, err := NewRegistry(u, 100, testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName := testutil.DatabaseName(t)
	p, err := r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, p)

	return r, p, dbName
}

func TestAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	var username string
	if u, _ := user.Current(); u != nil {
		username = u.Username
	}

	for name, tc := range map[string]struct {
		uri string
		err string
	}{
		"NoAuth": {
			uri: "postgres://127.0.0.1:5432/ferretdb",
			err: `server error: FATAL: role "` + username + `" does not exist (SQLSTATE 28000)`,
		},
		"WrongUser": {
			uri: "postgres://wrong-user:wrong-password@127.0.0.1:5432/ferretdb",
			err: `server error: FATAL: role "wrong-user" does not exist (SQLSTATE 28000)`,
		},
		"WrongDatabase": {
			uri: "postgres://username:password@127.0.0.1:5432/wrong-database",
			err: `server error: FATAL: database "wrong-database" does not exist (SQLSTATE 3D000)`,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			sp, err := state.NewProvider("")
			require.NoError(t, err)

			r, err := NewRegistry(tc.uri, 100, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			_, err = r.getPool(ctx)
			require.ErrorContains(t, err, tc.err)
		})
	}
}

func TestCreateDropStress(t *testing.T) {
	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, db, dbName := createDatabase(t, ctx)

	var i atomic.Int32

	teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
		collectionName := fmt.Sprintf("collection_%03d", i.Add(1))

		ready <- struct{}{}
		<-start

		testCollection(t, ctx, r, db, dbName, collectionName)
	})
}

func TestCreateSameStress(t *testing.T) {
	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, db, dbName := createDatabase(t, ctx)
	collectionName := testutil.CollectionName(t)

	var i, createdTotal atomic.Int32

	teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
		id := i.Add(1)

		ready <- struct{}{}
		<-start

		created, err := r.CollectionCreate(ctx, &CollectionCreateParams{DBName: dbName, Name: collectionName})
		require.NoError(t, err)
		if created {
			createdTotal.Add(1)
		}

		created, err = r.CollectionCreate(ctx, &CollectionCreateParams{DBName: dbName, Name: collectionName})
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
	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, _, dbName := createDatabase(t, ctx)
	collectionName := testutil.CollectionName(t)

	created, err := r.CollectionCreate(ctx, &CollectionCreateParams{DBName: dbName, Name: collectionName})
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
	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, _, dbName := createDatabase(t, ctx)
	collectionName := testutil.CollectionName(t)

	var i, createdTotal, droppedTotal atomic.Int32

	teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
		id := i.Add(1)

		ready <- struct{}{}
		<-start

		if id%2 == 0 {
			created, err := r.CollectionCreate(ctx, &CollectionCreateParams{DBName: dbName, Name: collectionName})
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

func TestDefaultEmptySchema(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, _, dbName := createDatabase(t, ctx)

	list, err := r.DatabaseList(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{dbName}, list)

	created, err := r.CollectionCreate(ctx, &CollectionCreateParams{DBName: "public", Name: testutil.CollectionName(t)})
	require.NoError(t, err)
	assert.True(t, created)

	list, err = r.DatabaseList(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{dbName, "public"}, list)
}

func TestCheckDatabaseUpdated(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, db, dbName := createDatabase(t, ctx)

	var err error

	t.Run("CheckDatabaseCreate", func(t *testing.T) {
		err = r.initCollections(ctx, dbName, db)
		require.NoError(t, err)

		var p *pgxpool.Pool
		p, err = r.DatabaseGetExisting(ctx, dbName)
		require.NoError(t, err)
		require.NotNil(t, p)
	})

	collectionName := testutil.CollectionName(t)
	created, err := r.CollectionCreate(ctx, &CollectionCreateParams{DBName: dbName, Name: collectionName})
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
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, db, dbName := createDatabase(t, ctx)

	oldCollectionName := testutil.CollectionName(t)
	newCollectionName := "new"

	created, err := r.CollectionCreate(ctx, &CollectionCreateParams{DBName: dbName, Name: oldCollectionName})
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
			UUID:      oldCollection.UUID,
			TableName: oldCollection.TableName,
			Indexes:   oldCollection.Indexes,
		}

		actual, err := r.CollectionGet(ctx, dbName, newCollectionName)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
	})
}

func TestMetadataIndexes(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	_, db, dbName := createDatabase(t, ctx)

	var sql string
	err := db.QueryRow(
		ctx,
		"SELECT indexdef FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 AND indexname = $3",
		dbName, metadataTableName, metadataTableName+"_id_idx",
	).Scan(&sql)
	require.NoError(t, err)

	expected := fmt.Sprintf(
		`CREATE UNIQUE INDEX %s ON %q.%s USING btree (((_jsonb -> '_id'::text)))`,
		metadataTableName+"_id_idx", dbName, metadataTableName,
	)
	assert.Equal(t, expected, sql)

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
}

func TestIndexesCreateDrop(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, db, dbName := createDatabase(t, ctx)
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
	}, {
		Name: "nested_fields",
		Key: []IndexKeyPair{{
			Field: "foo.bar",
		}, {
			Field:      "foo.baz",
			Descending: true,
		}},
	}}

	err := r.IndexesCreate(ctx, dbName, collectionName, toCreate)
	require.NoError(t, err)

	collection, err := r.CollectionGet(ctx, dbName, collectionName)
	require.NoError(t, err)

	t.Run("CreateIndexes", func(t *testing.T) {
		t.Run("NonUniqueIndex", func(t *testing.T) {
			t.Parallel()

			i := slices.IndexFunc(collection.Indexes, func(ii IndexInfo) bool {
				return ii.Name == "index_non_unique"
			})
			require.GreaterOrEqual(t, i, 0)
			tableIndexName := collection.Indexes[i].PgIndex

			var sql string
			err := db.QueryRow(
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
			t.Parallel()

			i := slices.IndexFunc(collection.Indexes, func(ii IndexInfo) bool {
				return ii.Name == "index_unique"
			})
			require.GreaterOrEqual(t, i, 0)
			tableIndexName := collection.Indexes[i].PgIndex

			var sql string
			err := db.QueryRow(
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
			t.Parallel()

			i := slices.IndexFunc(collection.Indexes, func(ii IndexInfo) bool {
				return ii.Name == "nested_fields"
			})
			require.GreaterOrEqual(t, i, 0)
			tableIndexName := collection.Indexes[i].PgIndex

			var sql string
			err := db.QueryRow(
				ctx,
				"SELECT indexdef FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 AND indexname = $3",
				dbName, collection.TableName, tableIndexName,
			).Scan(&sql)
			require.NoError(t, err)

			expected := fmt.Sprintf(
				`CREATE INDEX %s ON %q.%s USING btree`+
					` ((((_jsonb -> 'foo'::text) -> 'bar'::text)), (((_jsonb -> 'foo'::text) -> 'baz'::text)) DESC)`,
				tableIndexName, dbName, collection.TableName,
			)
			require.Equal(t, expected, sql)
		})

		t.Run("DefaultIndex", func(t *testing.T) {
			t.Parallel()

			i := slices.IndexFunc(collection.Indexes, func(ii IndexInfo) bool {
				return ii.Name == "_id_"
			})
			require.GreaterOrEqual(t, i, 0)
			tableIndexName := collection.Indexes[i].PgIndex

			var sql string
			err := db.QueryRow(
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
	})

	t.Run("CheckSettingsAfterCreation", func(t *testing.T) {
		// Force DBs and collection initialization to check that indexes metadata is stored correctly in the database.
		_, err = r.getPool(ctx)
		require.NoError(t, err)

		var refreshedCollection *Collection
		refreshedCollection, err = r.CollectionGet(ctx, dbName, collectionName)
		require.NoError(t, err)

		require.Equal(t, 4, len(refreshedCollection.Indexes))

		for _, index := range refreshedCollection.Indexes {
			switch index.Name {
			case "_id_":
				assert.Equal(t, 1, len(index.Key))
			case "index_non_unique":
				assert.Equal(t, 2, len(index.Key))
			case "index_unique":
				assert.Equal(t, 1, len(index.Key))
			case "nested_fields":
				assert.Equal(t, 2, len(index.Key))
			default:
				t.Errorf("unexpected index: %s", index.Name)
			}
		}
	})

	t.Run("DropIndexes", func(t *testing.T) {
		toDrop := []string{"index_non_unique", "nested_fields"}
		err := r.IndexesDrop(ctx, dbName, collectionName, toDrop)
		require.NoError(t, err)

		q := "SELECT count(indexdef) FROM pg_indexes WHERE schemaname = $1 AND tablename = $2"
		row := db.QueryRow(ctx, q, dbName, collection.TableName)

		var count int
		require.NoError(t, row.Scan(&count))
		require.Equal(t, 2, count) // only default index and index_unique should be left

		// Force DBs and collection initialization to check index metadata after deletion.
		_, err = r.getPool(ctx)
		require.NoError(t, err)

		collection, err = r.CollectionGet(ctx, dbName, collectionName)
		require.NoError(t, err)
		require.Equal(t, 2, len(collection.Indexes))

		for _, index := range collection.Indexes {
			switch index.Name {
			case "_id_":
				assert.Equal(t, 1, len(index.Key))
			case "index_unique":
				assert.Equal(t, 1, len(index.Key))
			default:
				t.Errorf("unexpected index: %s", index.Name)
			}
		}
	})
}

func TestLongIndexNames(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, _, dbName := createDatabase(t, ctx)

	batch1 := []IndexInfo{{
		Name: strings.Repeat("aB", 75),
		Key: []IndexKeyPair{{
			Field:      "foo",
			Descending: false,
		}, {
			Field:      "bar",
			Descending: true,
		}},
	}, {
		Name: strings.Repeat("aB", 75) + "_unique",
		Key: []IndexKeyPair{{
			Field:      "foo",
			Descending: false,
		}},
		Unique: true,
	}}

	batch2 := []IndexInfo{{
		Name: strings.Repeat("aB", 75) + "_bar",
		Key: []IndexKeyPair{{
			Field:      "bar",
			Descending: false,
		}},
	}}

	for name, tc := range map[string]struct {
		collectionName   string
		tablePartInIndex string
	}{
		"ShortCollectionName": {
			collectionName:   testutil.CollectionName(t),
			tablePartInIndex: "testlongindexnames_47546aa3",
		},
		"LongCollectionName": {
			collectionName:   "Collection" + strings.Repeat("cD", 75),
			tablePartInIndex: "collection" + strings.Repeat("cd", 10),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := r.IndexesCreate(ctx, dbName, tc.collectionName, batch1)
			require.NoError(t, err)

			collection, err := r.CollectionGet(ctx, dbName, tc.collectionName)
			require.NoError(t, err)
			require.Equal(t, 3, len(collection.Indexes))

			for _, index := range collection.Indexes {
				switch index.Name {
				case "_id_":
					assert.Equal(t, tc.tablePartInIndex+"__id__67399184_idx", index.PgIndex)
				case batch1[0].Name:
					assert.Equal(t, tc.tablePartInIndex+"_ababababababababab_12fa1dfe_idx", index.PgIndex)
				case batch1[1].Name:
					assert.Equal(t, tc.tablePartInIndex+"_ababababababababab_ca7ee610_idx", index.PgIndex)
				default:
					t.Errorf("unexpected index: %s", index.Name)
				}
			}

			err = r.IndexesCreate(ctx, dbName, tc.collectionName, batch2)
			require.NoError(t, err)

			// Force DBs and collection initialization to check that indexes metadata is stored correctly in the database.
			_, err = r.getPool(ctx)
			require.NoError(t, err)

			collection, err = r.CollectionGet(ctx, dbName, tc.collectionName)
			require.NoError(t, err)
			require.Equal(t, 4, len(collection.Indexes))

			for _, index := range collection.Indexes {
				switch index.Name {
				case "_id_":
					assert.Equal(t, tc.tablePartInIndex+"__id__67399184_idx", index.PgIndex)
				case batch1[0].Name:
					assert.Equal(t, tc.tablePartInIndex+"_ababababababababab_12fa1dfe_idx", index.PgIndex)
				case batch1[1].Name:
					assert.Equal(t, tc.tablePartInIndex+"_ababababababababab_ca7ee610_idx", index.PgIndex)
				case batch2[0].Name:
					assert.Equal(t, tc.tablePartInIndex+"_ababababababababab_aaf0d99c_idx", index.PgIndex)
				default:
					t.Errorf("unexpected index: %s", index.Name)
				}
			}
		})
	}
}
