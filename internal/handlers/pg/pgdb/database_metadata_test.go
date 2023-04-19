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
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestDatabaseMetadata(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)

	pool := getPool(ctx, t)
	databaseName := testutil.DatabaseName(t)
	collectionName := testutil.CollectionName(t)
	setupDatabase(ctx, t, pool, databaseName)

	err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		ms := newMetadataStorage(tx, databaseName, collectionName)
		nameCreated, _, err := ms.store(ctx)
		// In this case error is possible: if this test is run in parallel with other tests,
		// ensureMetadata may fail to create the index or insert data due to concurrent requests to PostgreSQL.
		// In such case, we expect InTransactionRetry to handle the error and retry the transaction if neede.
		if err != nil {
			return err
		}

		var nameFound string

		nameFound, err = ms.getTableName(ctx)
		require.NoError(t, err)

		assert.Equal(t, nameCreated, nameFound)

		// adding metadata that already exist should not fail
		_, _, err = ms.store(ctx)
		require.NoError(t, err)

		err = ms.remove(ctx)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)
}

func TestRenameCollection(t *testing.T) {
	ctx := testutil.Ctx(t)

	pool := getPool(ctx, t)

	t.Run("Simple", func(t *testing.T) {
		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)
		const newCollectionName = "new_name"
		setupDatabase(ctx, t, pool, databaseName)

		var tableName string
		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			err := CreateCollection(ctx, tx, databaseName, collectionName)
			require.NoError(t, err)

			md, err := newMetadataStorage(tx, databaseName, collectionName).get(ctx, false)
			require.NoError(t, err)
			assert.Equal(t, collectionName, md.collection)

			tableName = md.table
			require.NotEmpty(t, tableName)

			return nil
		})

		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			ms := newMetadataStorage(tx, databaseName, collectionName)
			err := ms.renameCollection(ctx, newCollectionName)
			require.NoError(t, err)

			return nil
		})

		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			exists, err := CollectionExists(ctx, tx, databaseName, newCollectionName)
			require.NoError(t, err)
			assert.True(t, exists)

			exists, err = CollectionExists(ctx, tx, databaseName, collectionName)
			require.NoError(t, err)
			assert.False(t, exists)

			ms := newMetadataStorage(tx, databaseName, newCollectionName)
			md, err := ms.get(ctx, false)
			require.NoError(t, err)

			assert.Equal(t, newCollectionName, md.collection)
			assert.Equal(t, tableName, md.table)

			ms = newMetadataStorage(tx, databaseName, collectionName)
			_, err = ms.get(ctx, false)
			require.Equal(t, ErrTableNotExist, err)

			return nil
		})
	})

	t.Run("AlreadyExist", func(t *testing.T) {
		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)
		const existingCollectionName = "existing_name"
		setupDatabase(ctx, t, pool, databaseName)

		var existingTableName, tableName string
		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			err := CreateCollection(ctx, tx, databaseName, existingCollectionName)
			require.NoError(t, err)

			md, err := newMetadataStorage(tx, databaseName, existingCollectionName).get(ctx, false)
			require.NoError(t, err)
			assert.Equal(t, existingCollectionName, md.collection)

			existingTableName = md.table
			require.NotEmpty(t, existingTableName)

			err = CreateCollection(ctx, tx, databaseName, collectionName)
			require.NoError(t, err)

			md, err = newMetadataStorage(tx, databaseName, collectionName).get(ctx, false)
			require.NoError(t, err)
			assert.Equal(t, collectionName, md.collection)

			tableName = md.table
			require.NotEmpty(t, tableName)

			return nil
		})

		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			ms := newMetadataStorage(tx, databaseName, collectionName)
			err := ms.renameCollection(ctx, existingCollectionName)
			require.Equal(t, ErrAlreadyExist, err)

			return nil
		})

		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			exists, err := CollectionExists(ctx, tx, databaseName, existingCollectionName)
			require.NoError(t, err)
			assert.True(t, exists)

			exists, err = CollectionExists(ctx, tx, databaseName, collectionName)
			require.NoError(t, err)
			assert.True(t, exists)

			ms := newMetadataStorage(tx, databaseName, existingCollectionName)
			md, err := ms.get(ctx, false)
			require.NoError(t, err)

			assert.Equal(t, existingCollectionName, md.collection)
			assert.Equal(t, existingTableName, md.table)

			ms = newMetadataStorage(tx, databaseName, collectionName)
			md, err = ms.get(ctx, false)
			require.NoError(t, err)

			assert.Equal(t, collectionName, md.collection)
			assert.Equal(t, tableName, md.table)

			return nil
		})
	})

	t.Run("NotExist", func(t *testing.T) {
		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)
		const newCollectionName = "new_name"
		setupDatabase(ctx, t, pool, databaseName)

		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			ms := newMetadataStorage(tx, databaseName, collectionName)
			err := ms.renameCollection(ctx, newCollectionName)
			require.ErrorIs(t, err, ErrTableNotExist)

			return nil
		})

		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			exists, err := CollectionExists(ctx, tx, databaseName, newCollectionName)
			require.NoError(t, err)
			assert.False(t, exists)

			exists, err = CollectionExists(ctx, tx, databaseName, collectionName)
			require.NoError(t, err)
			assert.False(t, exists)

			return nil
		})
	})

	t.Run("Serial", func(t *testing.T) {
		databaseName := testutil.DatabaseName(t)
		collectionName := testutil.CollectionName(t)
		const newCollectionName = "existing_name"
		setupDatabase(ctx, t, pool, databaseName)

		var tableName string
		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			err := CreateCollection(ctx, tx, databaseName, collectionName)
			require.NoError(t, err)

			md, err := newMetadataStorage(tx, databaseName, collectionName).get(ctx, false)
			require.NoError(t, err)
			assert.Equal(t, collectionName, md.collection)

			tableName = md.table
			require.NotEmpty(t, tableName)

			return nil
		})

		var newTableName string
		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			ms := newMetadataStorage(tx, databaseName, collectionName)
			err := ms.renameCollection(ctx, newCollectionName)
			require.NoError(t, err)

			err = CreateCollection(ctx, tx, databaseName, collectionName)
			require.NoError(t, err)

			md, err := newMetadataStorage(tx, databaseName, collectionName).get(ctx, false)
			require.NoError(t, err)
			assert.Equal(t, collectionName, md.collection)

			newTableName = md.table
			require.NotEmpty(t, newTableName)
			require.NotEqual(t, tableName, newTableName)

			return nil
		})

		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			exists, err := CollectionExists(ctx, tx, databaseName, newCollectionName)
			require.NoError(t, err)
			assert.True(t, exists)

			exists, err = CollectionExists(ctx, tx, databaseName, collectionName)
			require.NoError(t, err)
			assert.True(t, exists)

			ms := newMetadataStorage(tx, databaseName, newCollectionName)
			md, err := ms.get(ctx, false)
			require.NoError(t, err)

			assert.Equal(t, newCollectionName, md.collection)
			assert.Equal(t, tableName, md.table)

			ms = newMetadataStorage(tx, databaseName, collectionName)
			md, err = ms.get(ctx, false)
			require.NoError(t, err)

			assert.Equal(t, collectionName, md.collection)
			assert.Equal(t, newTableName, md.table)

			return nil
		})

	})
}
