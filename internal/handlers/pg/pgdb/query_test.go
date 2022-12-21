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
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestGetDocuments(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)

	pool := getPool(ctx, t)
	databaseName := testutil.DatabaseName(t)
	collectionName := testutil.CollectionName(t)
	setupDatabase(ctx, t, pool, databaseName)

	err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
		return CreateDatabaseIfNotExists(ctx, tx, databaseName)
	})
	require.NoError(t, err)

	t.Run("one-document", func(t *testing.T) {
		t.Parallel()

		collection := collectionName + "-one"
		expectedDoc := must.NotFail(types.NewDocument("_id", "foo", "id", "1"))

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		require.NoError(t, InsertDocument(ctx, tx, databaseName, collection, expectedDoc))

		sp := &SQLParam{DB: databaseName, Collection: collection}
		it, err := pool.GetDocuments(ctx, tx, sp)
		require.NoError(t, err)
		require.NotNil(t, it)

		defer it.Close()

		iter, doc, err := it.Next()
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), iter)
		assert.Equal(t, expectedDoc, doc)

		iter, doc, err = it.Next()
		assert.Equal(t, iterator.ErrIteratorDone, err)
		assert.Equal(t, uint32(0), iter)
		assert.Nil(t, doc)

		it.Close()
		require.NoError(t, tx.Commit(ctx))
	})

	t.Run("cancel-context", func(t *testing.T) {
		t.Parallel()

		collection := collectionName + "-two"
		expectedDocs := []*types.Document{
			must.NotFail(types.NewDocument("_id", "bar", "id", "1")),
			must.NotFail(types.NewDocument("_id", "foo", "id", "2")),
		}

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		require.NoError(t, InsertDocument(ctx, tx, databaseName, collection, expectedDocs[0]))
		require.NoError(t, InsertDocument(ctx, tx, databaseName, collection, expectedDocs[1]))

		ctxTest, cancel := context.WithCancel(ctx)
		sp := &SQLParam{DB: databaseName, Collection: collection}
		it, err := pool.GetDocuments(ctxTest, tx, sp)
		require.NoError(t, err)
		require.NotNil(t, it)

		defer it.Close()

		iter, doc, err := it.Next()
		assert.NoError(t, err)
		assert.Equal(t, uint32(0), iter)
		assert.Equal(t, expectedDocs[0], doc)

		cancel()
		iter, doc, err = it.Next()
		assert.Equal(t, context.Canceled, err)
		assert.Equal(t, uint32(0), iter)
		assert.Nil(t, doc)

		it.Close()
		require.NoError(t, tx.Commit(ctx))
	})

	t.Run("empty-collection", func(t *testing.T) {
		t.Parallel()

		collection := collectionName + "-empty"

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		require.NoError(t, CreateCollection(ctx, tx, databaseName, collection))

		sp := &SQLParam{DB: databaseName, Collection: collection}
		it, err := pool.GetDocuments(ctx, tx, sp)
		require.NoError(t, err)
		require.NotNil(t, it)

		defer it.Close()

		iter, doc, err := it.Next()
		assert.Equal(t, iterator.ErrIteratorDone, err)
		assert.Equal(t, uint32(0), iter)
		assert.Nil(t, doc)

		it.Close()
		require.NoError(t, tx.Commit(ctx))
	})

	t.Run("non-existent-collection", func(t *testing.T) {
		t.Parallel()

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		sp := &SQLParam{DB: databaseName, Collection: collectionName + "-non-existent"}
		it, err := pool.GetDocuments(ctx, tx, sp)
		require.NoError(t, err)
		require.NotNil(t, it)

		iter, doc, err := it.Next()
		assert.Equal(t, iterator.ErrIteratorDone, err)
		assert.Equal(t, uint32(0), iter)
		assert.Nil(t, doc)

		it.Close()
		require.NoError(t, tx.Commit(ctx))
	})
}
