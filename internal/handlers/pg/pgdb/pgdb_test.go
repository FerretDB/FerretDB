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
	"errors"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// getPool creates a new connection's connection pool for testing.
func getPool(ctx context.Context, tb testing.TB) *Pool {
	tb.Helper()

	logger := testutil.Logger(tb, zap.NewAtomicLevelAt(zap.DebugLevel))

	p, err := state.NewProvider("")
	require.NoError(tb, err)

	pool, err := NewPool(ctx, testutil.PostgreSQLURL(tb, nil), logger, p)
	require.NoError(tb, err)
	tb.Cleanup(pool.Close)

	return pool
}

// setupDatabase ensures that test-specific FerretDB database / PostgreSQL schema does not exist
// before and after the test.
func setupDatabase(ctx context.Context, tb testing.TB, pool *Pool, db string) {
	dropDatabase := func() {
		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropDatabase(ctx, tx, db)
		})
	}

	dropDatabase()
	tb.Cleanup(dropDatabase)
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
	pool := getPool(ctx, t)

	t.Run("NoDatabase", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)
		setupDatabase(ctx, t, pool, databaseName)

		err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropCollection(ctx, tx, databaseName, collectionName)
		})
		require.ErrorIs(t, err, ErrTableNotExist)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropDatabase(ctx, tx, databaseName)
		})
		require.ErrorIs(t, err, ErrSchemaNotExist)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return CreateCollection(ctx, tx, databaseName, collectionName)
		})
		require.NoError(t, err)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return CreateDatabaseIfNotExists(ctx, tx, databaseName)
		})
		require.NoError(t, err)

		var exists bool
		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			exists, err = CollectionExists(ctx, tx, databaseName, collectionName)
			return err
		})
		require.NoError(t, err)
		assert.True(t, exists)

		var collections []string
		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			collections, err = Collections(ctx, tx, databaseName)
			return err
		})
		require.NoError(t, err)
		assert.Equal(t, []string{collectionName}, collections)
	})

	t.Run("NoCollection", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)
		setupDatabase(ctx, t, pool, databaseName)

		err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
			if err := CreateDatabaseIfNotExists(ctx, tx, databaseName); err != nil && !errors.Is(err, ErrAlreadyExist) {
				return err
			}
			return nil
		})
		require.NoError(t, err)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			if err = CreateDatabaseIfNotExists(ctx, tx, databaseName); err != nil && !errors.Is(err, ErrAlreadyExist) {
				return err
			}
			return nil
		})
		require.NoError(t, err)

		var exists bool
		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			exists, err = CollectionExists(ctx, tx, databaseName, collectionName)
			return err
		})
		require.NoError(t, err)
		assert.False(t, exists)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropCollection(ctx, tx, databaseName, collectionName)
		})
		require.ErrorIs(t, err, ErrTableNotExist)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return CreateCollection(ctx, tx, databaseName, collectionName)
		})
		require.NoError(t, err)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			exists, err = CollectionExists(ctx, tx, databaseName, collectionName)
			return err
		})
		require.NoError(t, err)
		assert.True(t, exists)

		var collections []string
		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			collections, err = Collections(ctx, tx, databaseName)
			return err
		})
		assert.Equal(t, []string{collectionName}, collections)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropDatabase(ctx, tx, databaseName)
		})
		require.NoError(t, err)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropDatabase(ctx, tx, databaseName)
		})
		require.ErrorIs(t, err, ErrSchemaNotExist)
	})

	t.Run("Collection", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)
		setupDatabase(ctx, t, pool, databaseName)

		err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
			if err := CreateDatabaseIfNotExists(ctx, tx, databaseName); err != nil && !errors.Is(err, ErrAlreadyExist) {
				return err
			}
			return nil
		})
		require.NoError(t, err)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return CreateCollection(ctx, tx, databaseName, collectionName)
		})
		require.NoError(t, err)

		var collections []string
		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			collections, err = Collections(ctx, tx, databaseName)
			return err
		})
		require.NoError(t, err)
		assert.Equal(t, []string{collectionName}, collections)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return CreateCollection(ctx, tx, databaseName, collectionName)
		})
		require.ErrorIs(t, err, ErrAlreadyExist)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropCollection(ctx, tx, databaseName, collectionName)
		})
		require.NoError(t, err)

		err = pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return DropCollection(ctx, tx, databaseName, collectionName)
		})
		require.ErrorIs(t, err, ErrTableNotExist)
	})
}

func TestCreateCollectionIfNotExists(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t)

	t.Run("NoDatabase", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)
		setupDatabase(ctx, t, pool, databaseName)

		err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
			return CreateCollectionIfNotExists(ctx, tx, databaseName, collectionName)
		})
		require.NoError(t, err)
	})

	t.Run("Database", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)
		setupDatabase(ctx, t, pool, databaseName)

		err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
			if err := CreateDatabaseIfNotExists(ctx, tx, databaseName); err != nil {
				return err
			}

			return CreateCollectionIfNotExists(ctx, tx, databaseName, collectionName)
		})
		require.NoError(t, err)
	})

	t.Run("Collection", func(t *testing.T) {
		t.Parallel()

		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)
		setupDatabase(ctx, t, pool, databaseName)

		err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
			if err := CreateDatabaseIfNotExists(ctx, tx, databaseName); err != nil {
				return err
			}

			if err := CreateCollection(ctx, tx, databaseName, collectionName); err != nil {
				return err
			}

			return CreateCollectionIfNotExists(ctx, tx, databaseName, collectionName)
		})
		require.NoError(t, err)
	})
}
