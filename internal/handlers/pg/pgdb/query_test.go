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
	setupDatabase(ctx, t, pool, databaseName)

	doc1 := must.NotFail(types.NewDocument("_id", int32(1)))
	doc2 := must.NotFail(types.NewDocument("_id", int32(1)))

	t.Run("Normal", func(t *testing.T) {
		t.Parallel()

		collectionName := testutil.CollectionName(t)

		err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
			err := InsertDocument(ctx, tx, databaseName, collectionName, doc1)
			require.NoError(t, err)

			err = InsertDocument(ctx, tx, databaseName, collectionName, doc2)
			require.NoError(t, err)

			sp := &SQLParam{DB: databaseName, Collection: collectionName}
			iter, err := GetDocuments(ctx, tx, sp)
			require.NoError(t, err)
			require.NotNil(t, iter)

			defer iter.Close()

			n, doc, err := iter.Next()
			require.NoError(t, err)
			assert.Equal(t, 0, n)
			assert.Equal(t, doc1, doc)

			n, doc, err = iter.Next()
			require.NoError(t, err)
			assert.Equal(t, 1, n)
			assert.Equal(t, doc2, doc)

			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("EarlyClose", func(t *testing.T) {
		t.Parallel()

		collectionName := testutil.CollectionName(t)

		err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
			err := InsertDocument(ctx, tx, databaseName, collectionName, doc1)
			require.NoError(t, err)

			sp := &SQLParam{DB: databaseName, Collection: collectionName}
			iter, err := GetDocuments(ctx, tx, sp)
			require.NoError(t, err)
			require.NotNil(t, iter)

			iter.Close()

			n, doc, err := iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("CancelContext", func(t *testing.T) {
		t.Parallel()

		ctxGet, cancelGet := context.WithCancel(ctx)

		collectionName := testutil.CollectionName(t)

		err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
			err := InsertDocument(ctx, tx, databaseName, collectionName, doc1)
			require.NoError(t, err)

			sp := &SQLParam{DB: databaseName, Collection: collectionName}
			iter, err := GetDocuments(ctxGet, tx, sp)
			require.NoError(t, err)
			require.NotNil(t, iter)

			cancelGet()

			// FIXME context.Canceled first, done later?

			n, doc, err := iter.Next()
			require.Equal(t, context.Canceled, err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("EmptyCollection", func(t *testing.T) {
		t.Parallel()

		collection := testutil.CollectionName(t)

		err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
			err := CreateCollection(ctx, tx, databaseName, collection)
			require.NoError(t, err)

			sp := &SQLParam{DB: databaseName, Collection: collection}
			iter, err := GetDocuments(ctx, tx, sp)
			require.NoError(t, err)
			require.NotNil(t, iter)

			defer iter.Close()

			n, doc, err := iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("NonExistentCollection", func(t *testing.T) {
		t.Parallel()

		collection := testutil.CollectionName(t)

		err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
			sp := &SQLParam{DB: databaseName, Collection: collection}
			iter, err := GetDocuments(ctx, tx, sp)
			require.NoError(t, err)
			require.NotNil(t, iter)

			defer iter.Close()

			n, doc, err := iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})
}
