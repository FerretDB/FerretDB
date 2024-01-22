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
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

// testCollection creates, tests, and drops a unique collection in the existing database.
func testCollection(t *testing.T, ctx context.Context, r *Registry, p *fsql.DB, dbName, collectionName string) {
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
		`INSERT INTO %s.%s (%s) VALUES (?)`,
		dbName, c.TableName,
		DefaultColumn,
	)
	doc := `{"$s": {"p": {"_id": {"t": "int"}}, "$k": ["_id"]}, "_id": 42}`
	_, err = p.ExecContext(ctx, q, doc)
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
func createDatabase(t *testing.T, ctx context.Context) (r *Registry, db *fsql.DB, dbName string) {
	t.Helper()

	u := testutil.TestMySQLURI(t, ctx, "")
	require.NotEmpty(t, u)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	r, err = NewRegistry(u, testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName = testutil.DatabaseName(t)
	db, err = r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, db)

	t.Cleanup(func() {
		_, err = r.DatabaseDrop(ctx, dbName)
		require.NoError(t, err)
	})

	return r, db, dbName
}

func TestCheckAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	t.Run("Auth", func(t *testing.T) {
		t.Parallel()

		r, err := NewRegistry("mysql://username:password@127.0.0.1:3306/ferretdb", testutil.Logger(t), sp)
		require.NoError(t, err)
		t.Cleanup(r.Close)

		_, err = r.getPool(ctx)
		require.NoError(t, err)
	})

	t.Run("WrongUser", func(t *testing.T) {
		t.Parallel()

		r, err := NewRegistry(
			"mysql://wrong-user:wrong-password@127.0.0.1:3306/ferretdb?allowNativePasswords=true",
			testutil.Logger(t),
			sp,
		)
		require.NoError(t, err)
		t.Cleanup(r.Close)

		_, err = r.getPool(ctx)

		expected := `Error 1045 \(28000\): Access denied for user 'wrong-user*'@'[\d\.]+' \(using password: YES\)`
		assert.Regexp(t, expected, err)
	})

	t.Run("WrongDatabase", func(t *testing.T) {
		t.Parallel()

		r, err := NewRegistry("mysql://username:password@127.0.0.1:3306/wrong-database", testutil.Logger(t), sp)
		require.NoError(t, err)
		t.Cleanup(r.Close)

		_, err = r.getPool(ctx)

		expected := `Error 1049 (42000): Unknown database 'wrong-database'`
		require.ErrorContains(t, err, expected)
	})
}

func TestDefaultEmptySchema(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, _, dbName := createDatabase(t, ctx)

	list, err := r.DatabaseList(ctx)

	require.NoError(t, err)
	assert.Equal(t, []string{dbName}, list)
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
			`INSERT INTO %s.%s (%s) VALUES(?)`,
			dbName, c.TableName,
			DefaultColumn,
		)
		doc := fmt.Sprintf(`{"$s": {"p": {"_id": {"t": "int"}}, "$k": ["_id"]}, "_id": %d}`, id)
		_, err = db.ExecContext(ctx, q, doc)
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

func TestCheckDatabaseUpdated(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	r, db, dbName := createDatabase(t, ctx)

	var err error

	t.Run("CheckDatabaseCreate", func(t *testing.T) {
		err = r.initCollections(ctx, dbName, db)
		require.NoError(t, err)

		var p *fsql.DB
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
		require.ErrorContains(t, err, "Unknown database 'TestCheckDatabaseUpdated'")
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
			collectionName:   "Collections" + strings.Repeat("cD", 75),
			tablePartInIndex: "collections" + strings.Repeat("cd", 10),
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
					assert.Equal(t, tc.tablePartInIndex+"__id__67399184_idx", index.Index)
				case batch1[0].Name:
					assert.Equal(t, tc.tablePartInIndex+"_abababababababababa_12fa1dfe_idx", index.Index)
				case batch1[1].Name:
					assert.Equal(t, tc.tablePartInIndex+"_abababababababababa_ca7ee610_idx", index.Index)
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
					assert.Equal(t, tc.tablePartInIndex+"__id__67399184_idx", index.Index)
				case batch1[0].Name:
					assert.Equal(t, tc.tablePartInIndex+"_abababababababababa_12fa1dfe_idx", index.Index)
				case batch1[1].Name:
					assert.Equal(t, tc.tablePartInIndex+"_abababababababababa_ca7ee610_idx", index.Index)
				case batch2[0].Name:
					assert.Equal(t, tc.tablePartInIndex+"_abababababababababa_aaf0d99c_idx", index.Index)
				default:
					t.Errorf("unexpected index: %s", index.Name)
				}
			}
		})
	}
}
