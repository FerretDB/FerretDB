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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

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

		r, err := NewRegistry("mysql://wrong-user:wrong-password@127.0.0.1:3306/ferretdb?allowNativePasswords=true", testutil.Logger(t), sp)
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
