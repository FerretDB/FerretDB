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

// Use _test package to avoid import cycle with testutil.
package pgdb_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

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

			actual := pgdb.IsValidUTF8Locale(tc.locale)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCreateDrop(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t, nil, zaptest.NewLogger(t))

	t.Run("SchemaDoesNotExistTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		schemaName := testutil.SchemaName(t)
		tableName := testutil.TableName(t)

		t.Cleanup(func() {
			pool.DropDatabase(ctx, schemaName)
		})

		// Schema does not exist ->
		// - table drop is not possible
		// - schema drop is not possible
		// - table creation is not possible
		// - schema creation is possible

		err := pool.DropCollection(ctx, schemaName, tableName)
		require.Equal(t, pgdb.ErrSchemaNotExist, err)

		err = pool.DropDatabase(ctx, schemaName)
		require.Equal(t, pgdb.ErrSchemaNotExist, err)

		err = pgdb.CreateCollection(ctx, pool, schemaName, tableName)
		require.ErrorIs(t, err, pgdb.ErrSchemaNotExist)

		err = pool.CreateDatabase(ctx, schemaName)
		require.NoError(t, err)

		tables, err := pool.Collections(ctx, schemaName)
		require.NoError(t, err)
		assert.Empty(t, tables)
	})

	t.Run("SchemaExistsTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		schemaName := testutil.SchemaName(t)
		tableName := testutil.TableName(t)

		t.Cleanup(func() {
			pool.DropDatabase(ctx, schemaName)
		})

		err := pool.CreateDatabase(ctx, schemaName)
		require.NoError(t, err)

		// Schema exists ->
		// - schema creation is not possible
		// - table drop is not possible
		// - table creation is possible
		// - schema drop is possible (only once)

		err = pgdb.CreateDatabase(ctx, pool, schemaName)
		require.ErrorIs(t, err, pgdb.ErrAlreadyExist)

		err = pool.DropCollection(ctx, schemaName, tableName)
		require.ErrorIs(t, err, pgdb.ErrTableNotExist)

		err = pgdb.CreateCollection(ctx, pool, schemaName, tableName)
		require.NoError(t, err)

		tables, err := pool.Collections(ctx, schemaName)
		require.NoError(t, err)
		assert.Equal(t, []string{tableName}, tables)

		err = pool.DropDatabase(ctx, schemaName)
		require.NoError(t, err)

		err = pool.DropDatabase(ctx, schemaName)
		require.Equal(t, pgdb.ErrSchemaNotExist, err)
	})

	t.Run("SchemaExistsTableExists", func(t *testing.T) {
		t.Parallel()

		schemaName := testutil.SchemaName(t)
		tableName := testutil.TableName(t)

		t.Cleanup(func() {
			pool.DropDatabase(ctx, schemaName)
		})

		err := pgdb.CreateDatabase(ctx, pool, schemaName)
		require.NoError(t, err)

		err = pgdb.CreateCollection(ctx, pool, schemaName, tableName)
		require.NoError(t, err)

		tables, err := pool.Collections(ctx, schemaName)
		require.NoError(t, err)
		assert.Equal(t, []string{tableName}, tables)

		// Table exists ->
		// - table creation is not possible
		// - schema creation is not possible
		// - table drop is possible (only once)
		// - schema drop is possible

		err = pgdb.CreateCollection(ctx, pool, schemaName, tableName)
		require.ErrorIs(t, err, pgdb.ErrAlreadyExist)

		err = pgdb.CreateDatabase(ctx, pool, schemaName)
		require.ErrorIs(t, err, pgdb.ErrAlreadyExist)

		err = pool.DropCollection(ctx, schemaName, tableName)
		require.NoError(t, err)

		err = pool.DropCollection(ctx, schemaName, tableName)
		require.ErrorIs(t, err, pgdb.ErrTableNotExist)

		err = pool.DropDatabase(ctx, schemaName)
		require.NoError(t, err)
	})
}

func TestConcurrentCreate(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	createPool := getPool(ctx, t, nil, zaptest.NewLogger(t))
	dbName := testutil.SchemaName(t) // using schema name helper for database name is good enough
	_, err := createPool.Exec(ctx, `CREATE DATABASE `+dbName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := createPool.Exec(ctx, `DROP DATABASE `+dbName)
		require.NoError(t, err)
	})

	n := 10
	dsn := fmt.Sprintf("postgres://postgres@127.0.0.1:5432/%[1]s?pool_min_conns=%[2]d&pool_max_conns=%[2]d", dbName, n)
	pool, err := pgdb.NewPool(ctx, dsn, zaptest.NewLogger(t), false)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	schemaName := testutil.SchemaName(t)
	tableName := testutil.TableName(t)

	for _, tc := range []struct {
		name        string
		f           func() error
		compareFunc func(*testing.T, int) bool
	}{
		{
			name: "CreateDatabase",
			f: func() error {
				return pgdb.CreateDatabase(ctx, pool, schemaName)
			},
			compareFunc: func(t *testing.T, errors int) bool {
				return assert.Equal(t, n-1, errors)
			},
		}, {
			name: "CreateCollection",
			f: func() error {
				return pgdb.CreateCollection(ctx, pool, schemaName, tableName)
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
				assert.ErrorIs(t, err, pgdb.ErrAlreadyExist)
			}

			tc.compareFunc(t, errors)

			// one more time to check "normal" error (DuplicateSchema, DuplicateTable)
			assert.ErrorIs(t, tc.f(), pgdb.ErrAlreadyExist)
		})
	}
}

func TestTableExists(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t, nil, zaptest.NewLogger(t))

	t.Run("SchemaDoesNotExistTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		schemaName := testutil.SchemaName(t)
		tableName := testutil.TableName(t)

		t.Cleanup(func() {
			pool.DropDatabase(ctx, schemaName)
		})

		ok, err := pool.CollectionExists(ctx, schemaName, tableName)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("SchemaExistsTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		schemaName := testutil.SchemaName(t)
		tableName := testutil.TableName(t)

		pool.CreateDatabase(ctx, schemaName)

		t.Cleanup(func() {
			pool.DropDatabase(ctx, schemaName)
		})

		ok, err := pool.CollectionExists(ctx, schemaName, tableName)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("SchemaExistsTableExists", func(t *testing.T) {
		t.Parallel()

		schemaName := testutil.SchemaName(t)
		tableName := testutil.TableName(t)

		pool.CreateDatabase(ctx, schemaName)
		pgdb.CreateCollection(ctx, pool, schemaName, tableName)

		t.Cleanup(func() {
			pool.DropDatabase(ctx, schemaName)
		})

		ok, err := pool.CollectionExists(ctx, schemaName, tableName)
		require.NoError(t, err)
		assert.True(t, ok)
	})
}

func TestCreateTableIfNotExist(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t, nil, zaptest.NewLogger(t))

	t.Run("SchemaDoesNotExistTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		schemaName := testutil.SchemaName(t)
		tableName := testutil.TableName(t)

		t.Cleanup(func() {
			pool.DropDatabase(ctx, schemaName)
		})

		ok, err := pool.CreateTableIfNotExist(ctx, schemaName, tableName)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("SchemaExistsTableDoesNotExist", func(t *testing.T) {
		t.Parallel()

		schemaName := testutil.SchemaName(t)
		tableName := testutil.TableName(t)

		pool.CreateDatabase(ctx, schemaName)

		t.Cleanup(func() {
			pool.DropDatabase(ctx, schemaName)
		})

		created, err := pool.CreateTableIfNotExist(ctx, schemaName, tableName)
		require.NoError(t, err)
		assert.True(t, created)
	})

	t.Run("SchemaExistsTableExists", func(t *testing.T) {
		t.Parallel()

		schemaName := testutil.SchemaName(t)
		tableName := testutil.TableName(t)

		pool.CreateDatabase(ctx, schemaName)
		pgdb.CreateCollection(ctx, pool, schemaName, tableName)

		t.Cleanup(func() {
			pool.DropDatabase(ctx, schemaName)
		})

		created, err := pool.CreateTableIfNotExist(ctx, schemaName, tableName)
		require.NoError(t, err)
		assert.False(t, created)
	})
}
