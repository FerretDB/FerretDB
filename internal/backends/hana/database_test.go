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

package hana

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/stretchr/testify/require"
)

func TestDatabaseStats(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	hanaURL, err := testutil.TestHanaURI()
	require.NoError(t, err)

	params := NewBackendParams{
		URI: hanaURL,
		L:   testutil.Logger(t),
	}

	b, err := NewBackend(&params)
	require.NoError(t, err)
	t.Cleanup(b.Close)

	dbName := testutil.DatabaseName(t)
	db, err := b.Database(dbName)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = b.DropDatabase(ctx, &backends.DropDatabaseParams{Name: dbName})
		require.NoError(t, err)
	})

	cNames := []string{"collectionOne", "collectionTwo"}
	for _, cName := range cNames {
		err = db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: cName})
		require.NoError(t, err)
		require.NotNil(t, db)
	}

	t.Run("DatabaseWithCollections", func(t *testing.T) {
		res, err := db.Stats(ctx, &backends.DatabaseStatsParams{
			Refresh: true,
		})
		require.NoError(t, err)
		// Should this be non zero if we don't have data in collections?
		// require.NotZero(t, res.SizeTotal)
		// require.NotZero(t, res.SizeCollections)
		require.Zero(t, res.CountDocuments)
	})
}

func TestDatabaseStatsFreeStorage(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	hanaURL, err := testutil.TestHanaURI()
	require.NoError(t, err)

	for name, _ := range map[string]string{
		"dir":    "",
		"memory": "" + "/?mode=memory",
	} {
		name := name
		t.Run(name, func(t *testing.T) {
			params := NewBackendParams{
				URI: hanaURL,
				L:   testutil.Logger(t),
			}

			b, err := NewBackend(&params)
			require.NoError(t, err)

			t.Cleanup(b.Close)

			dbName := testutil.DatabaseName(t)
			db, err := b.Database(dbName)
			require.NoError(t, err)

			t.Cleanup(func() {
				err = b.DropDatabase(ctx, &backends.DropDatabaseParams{Name: dbName})
				require.NoError(t, err)
			})

			cNames := []string{"collectionOne", "collectionTwo"}
			for _, cName := range cNames {
				err = db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: cName})
				require.NoError(t, err)
				require.NotNil(t, db)
			}

			res, err := db.Stats(ctx, new(backends.DatabaseStatsParams))
			require.NoError(t, err)

			t.Logf("freeStorage size: %d", res.SizeFreeStorage)
			require.Zero(t, res.SizeFreeStorage)

			c, err := db.Collection(cNames[0])
			require.NoError(t, err)

			nInsert, deleteFromIndex, deleteToIndex := 50, 10, 40
			ids := make([]any, nInsert)
			toInsert := make([]*types.Document, nInsert)
			for i := 0; i < nInsert; i++ {
				ids[i] = types.NewObjectID()
				toInsert[i] = must.NotFail(types.NewDocument("_id", ids[i], "v", "foo"))
			}

			_, err = c.InsertAll(ctx, &backends.InsertAllParams{Docs: toInsert})
			require.NoError(t, err)

			_, err = c.DeleteAll(ctx, &backends.DeleteAllParams{IDs: ids[deleteFromIndex:deleteToIndex]})
			require.NoError(t, err)

			res, err = db.Stats(ctx, new(backends.DatabaseStatsParams))
			require.NoError(t, err)

			t.Logf("freeStorage size: %d", res.SizeFreeStorage)
			require.NotZero(t, res.SizeFreeStorage)
		})
	}
}
