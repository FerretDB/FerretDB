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

package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestListDatabases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	b, err := NewBackend(&NewBackendParams{URI: testutil.TestSQLiteURI(t, ""), L: testutil.Logger(t), P: sp})
	require.NoError(t, err)
	t.Cleanup(b.Close)

	dbNames := []string{"testDB1", "testDB2", "testDB3"}

	testDB, err := b.Database(dbNames[0])
	require.NoError(t, err)
	err = testDB.CreateCollection(ctx, &backends.CreateCollectionParams{Name: "testCollection1"})
	require.NoError(t, err)
	defer b.DropDatabase(ctx, &backends.DropDatabaseParams{Name: dbNames[0]})

	testDB, err = b.Database(dbNames[1])
	require.NoError(t, err)
	err = testDB.CreateCollection(ctx, &backends.CreateCollectionParams{Name: "testCollection1"})
	require.NoError(t, err)
	defer b.DropDatabase(ctx, &backends.DropDatabaseParams{Name: dbNames[1]})

	testDB, err = b.Database(dbNames[2])
	require.NoError(t, err)
	err = testDB.CreateCollection(ctx, &backends.CreateCollectionParams{Name: "testCollection1"})
	require.NoError(t, err)
	defer b.DropDatabase(ctx, &backends.DropDatabaseParams{Name: dbNames[2]})

	t.Run("ListDatabases with specific name", func(t *testing.T) {
		res, err := b.ListDatabases(ctx, &backends.ListDatabasesParams{
			Name: dbNames[0],
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(res.Databases))
		require.Equal(t, dbNames[0], res.Databases[0].Name)
	})

	t.Run("ListDatabases with wrong name", func(t *testing.T) {
		res, err := b.ListDatabases(ctx, &backends.ListDatabasesParams{
			Name: "not-existing",
		})
		require.NoError(t, err)
		require.Equal(t, 0, len(res.Databases))
	})

	t.Run("ListDatabases with nil param", func(t *testing.T) {
		res, err := b.ListDatabases(ctx, nil)
		require.NoError(t, err)
		require.Equal(t, 3, len(res.Databases))
		require.Equal(t, dbNames[0], res.Databases[0].Name)
		require.Equal(t, dbNames[1], res.Databases[1].Name)
		require.Equal(t, dbNames[2], res.Databases[2].Name)
	})

	t.Run("ListDatabases with nil param", func(t *testing.T) {
		res, err := b.ListDatabases(ctx, &backends.ListDatabasesParams{})
		require.NoError(t, err)
		require.Equal(t, 3, len(res.Databases))
		require.Equal(t, dbNames[0], res.Databases[0].Name)
		require.Equal(t, dbNames[1], res.Databases[1].Name)
		require.Equal(t, dbNames[2], res.Databases[2].Name)
	})
}
