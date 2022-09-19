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

package pgdb

import (
	"context"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// getPool creates a new connection's connection pool for testing.
func getPool(ctx context.Context, tb testing.TB, l *zap.Logger) *Pool {
	tb.Helper()

	pool, err := NewPool(ctx, testutil.PostgreSQLURL(tb, nil), l, false)
	require.NoError(tb, err)
	tb.Cleanup(pool.Close)

	return pool
}

func TestValidUTF8Locale(t *testing.T) {
	t.Parallel()

	cases := []struct {
		locale   string
		expected bool
	}{
		{"en_US.utf8", true},
		{"en_US.utf-8", true},
		{"en_US.UTF8", true},
		{"en_US.UTF-8", true},
		{"en_UK.UTF-8", false},
		{"en_UK.utf--8", false},
		{"en_US", false},
		{"utf8", false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.locale, func(t *testing.T) {
			t.Parallel()

			actual := isValidUTF8Locale(tc.locale)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCreateDrop(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t, zaptest.NewLogger(t))

	t.Run("SchemaDoesNotExistTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)

		t.Cleanup(func() {
			pool.InTransaction(ctx, func(tx pgx.Tx) error {
				DropDatabase(ctx, tx, databaseName)
				return nil
			})
		})

		// Schema does not exist ->
		// - table drop is not possible
		// - schema drop is not possible
		// - table creation is not possible
		// - schema creation is possible

		err := DropCollection(ctx, pool, databaseName, collectionName)
		require.Equal(t, ErrSchemaNotExist, err)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropDatabase(ctx, tx, databaseName)
		})
		require.ErrorIs(t, err, ErrSchemaNotExist)

		err = CreateCollection(ctx, pool, databaseName, collectionName)
		require.ErrorIs(t, err, ErrSchemaNotExist)

		err = CreateDatabase(ctx, pool, databaseName)
		require.NoError(t, err)

		tables, err := Collections(ctx, pool, databaseName)
		require.NoError(t, err)
		assert.Empty(t, tables)
	})

	t.Run("SchemaExistsTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)

		t.Cleanup(func() {
			pool.InTransaction(ctx, func(tx pgx.Tx) error {
				DropDatabase(ctx, tx, databaseName)
				return nil
			})
		})

		err := CreateDatabase(ctx, pool, databaseName)
		require.NoError(t, err)

		// Schema exists ->
		// - schema creation is not possible
		// - table drop is not possible
		// - table creation is possible
		// - schema drop is possible (only once)

		err = CreateDatabase(ctx, pool, databaseName)
		require.ErrorIs(t, err, ErrAlreadyExist)

		err = DropCollection(ctx, pool, databaseName, collectionName)
		require.ErrorIs(t, err, ErrTableNotExist)

		err = CreateCollection(ctx, pool, databaseName, collectionName)
		require.NoError(t, err)

		tables, err := Collections(ctx, pool, databaseName)
		require.NoError(t, err)
		assert.Equal(t, []string{collectionName}, tables)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropDatabase(ctx, tx, databaseName)
		})
		require.NoError(t, err)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropDatabase(ctx, tx, databaseName)
		})
		require.ErrorIs(t, err, ErrSchemaNotExist)
	})

	t.Run("SchemaExistsTableExists", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)

		t.Cleanup(func() {
			pool.InTransaction(ctx, func(tx pgx.Tx) error {
				DropDatabase(ctx, tx, databaseName)
				return nil
			})
		})

		err := CreateDatabase(ctx, pool, databaseName)
		require.NoError(t, err)

		err = CreateCollection(ctx, pool, databaseName, collectionName)
		require.NoError(t, err)

		tables, err := Collections(ctx, pool, databaseName)
		require.NoError(t, err)
		assert.Equal(t, []string{collectionName}, tables)

		// Table exists ->
		// - table creation is not possible
		// - schema creation is not possible
		// - table drop is possible (only once)
		// - schema drop is possible

		err = CreateCollection(ctx, pool, databaseName, collectionName)
		require.ErrorIs(t, err, ErrAlreadyExist)

		err = CreateDatabase(ctx, pool, databaseName)
		require.ErrorIs(t, err, ErrAlreadyExist)

		err = DropCollection(ctx, pool, databaseName, collectionName)
		require.NoError(t, err)

		err = DropCollection(ctx, pool, databaseName, collectionName)
		require.ErrorIs(t, err, ErrTableNotExist)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropDatabase(ctx, tx, databaseName)
		})
		require.NoError(t, err)
	})
}

func TestConcurrentCreate(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	databaseName := testutil.DatabaseName(t)
	collectionName := testutil.CollectionName(t)

	// Create PostgreSQL database with the same name as FerretDB database / PostgreSQL schema
	// because it is good enough.
	createPool := getPool(ctx, t, zaptest.NewLogger(t))
	_, err := createPool.Exec(ctx, `CREATE DATABASE `+databaseName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := createPool.Exec(ctx, `DROP DATABASE `+databaseName)
		require.NoError(t, err)
	})

	n := 10
	dsn := testutil.PostgreSQLURL(t, &testutil.PostgreSQLURLOpts{
		DatabaseName: databaseName,
		Params: map[string]string{
			"pool_min_conns": strconv.Itoa(n),
			"pool_max_conns": strconv.Itoa(n),
		},
	})
	pool, err := NewPool(ctx, dsn, zaptest.NewLogger(t), false)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	for _, tc := range []struct {
		name        string
		f           func() error
		compareFunc func(*testing.T, int) bool
	}{
		{
			name: "CreateDatabase",
			f: func() error {
				return CreateDatabase(ctx, pool, databaseName)
			},
			compareFunc: func(t *testing.T, errors int) bool {
				return assert.Equal(t, n-1, errors)
			},
		}, {
			name: "CreateCollection",
			f: func() error {
				return CreateCollection(ctx, pool, databaseName, collectionName)
			},
			compareFunc: func(t *testing.T, errors int) bool {
				return assert.LessOrEqual(t, errors, n-1)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start := make(chan struct{})
			res := make(chan error, n)
			for i := 0; i < n; i++ {
				go func() {
					<-start
					res <- tc.f()
				}()
			}

			close(start)

			var errors int
			for i := 0; i < n; i++ {
				err := <-res
				if err == nil {
					continue
				}

				errors++
				assert.ErrorIs(t, err, ErrAlreadyExist)
			}

			tc.compareFunc(t, errors)

			// one more time to check "normal" error (DuplicateSchema, DuplicateTable)
			assert.ErrorIs(t, tc.f(), ErrAlreadyExist)
		})
	}
}

func TestTableExists(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t, zaptest.NewLogger(t))

	t.Run("SchemaDoesNotExistTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)

		t.Cleanup(func() {
			pool.InTransaction(ctx, func(tx pgx.Tx) error {
				DropDatabase(ctx, tx, databaseName)
				return nil
			})
		})

		ok, err := CollectionExists(ctx, pool, databaseName, collectionName)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("SchemaExistsTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)

		CreateDatabase(ctx, pool, databaseName)

		t.Cleanup(func() {
			pool.InTransaction(ctx, func(tx pgx.Tx) error {
				DropDatabase(ctx, tx, databaseName)
				return nil
			})
		})

		ok, err := CollectionExists(ctx, pool, databaseName, collectionName)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("SchemaExistsTableExists", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)

		CreateDatabase(ctx, pool, databaseName)
		CreateCollection(ctx, pool, databaseName, collectionName)

		t.Cleanup(func() {
			pool.InTransaction(ctx, func(tx pgx.Tx) error {
				DropDatabase(ctx, tx, databaseName)
				return nil
			})
		})

		ok, err := CollectionExists(ctx, pool, databaseName, collectionName)
		require.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestCreateTableIfNotExist(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t, zaptest.NewLogger(t))

	t.Run("SchemaDoesNotExistTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)

		t.Cleanup(func() {
			pool.InTransaction(ctx, func(tx pgx.Tx) error {
				DropDatabase(ctx, tx, databaseName)
				return nil
			})
		})

		ok, err := CreateCollectionIfNotExist(ctx, pool, databaseName, collectionName)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("SchemaExistsTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)

		CreateDatabase(ctx, pool, databaseName)

		t.Cleanup(func() {
			pool.InTransaction(ctx, func(tx pgx.Tx) error {
				DropDatabase(ctx, tx, databaseName)
				return nil
			})
		})

		created, err := CreateCollectionIfNotExist(ctx, pool, databaseName, collectionName)
		require.NoError(t, err)
		assert.True(t, created)
	})

	t.Run("SchemaExistsTableExists", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)

		CreateDatabase(ctx, pool, databaseName)
		CreateCollection(ctx, pool, databaseName, collectionName)

		t.Cleanup(func() {
			pool.InTransaction(ctx, func(tx pgx.Tx) error {
				DropDatabase(ctx, tx, databaseName)
				return nil
			})
		})

		created, err := CreateCollectionIfNotExist(ctx, pool, databaseName, collectionName)
		require.NoError(t, err)
		assert.False(t, created)
	})
}
