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
	"fmt"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/net/context"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestQueryDocuments(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)

	pool := getPool(ctx, t, zaptest.NewLogger(t))
	dbName := testutil.DatabaseName(t)
	collectionName := testutil.CollectionName(t)

	t.Cleanup(func() {
		pool.InTransaction(ctx, func(tx pgx.Tx) error {
			DropDatabase(ctx, tx, dbName)
			return nil
		})
	})

	pool.InTransaction(ctx, func(tx pgx.Tx) error {
		DropDatabase(ctx, tx, dbName)
		return nil
	})
	require.NoError(t, CreateDatabase(ctx, pool, dbName))

	cases := []struct {
		name       string
		collection string
		documents  []*types.Document

		// docsPerIteration represents how many documents should be fetched per each iteration,
		// use len(docsPerIteration) as the amount of fetch iterations.
		docsPerIteration []int
	}{
		{
			name:             "empty",
			collection:       collectionName,
			documents:        []*types.Document{},
			docsPerIteration: []int{},
		},
		{
			name:             "one",
			collection:       collectionName + "_one",
			documents:        []*types.Document{must.NotFail(types.NewDocument("id", "1"))},
			docsPerIteration: []int{1},
		},
		{
			name:       "two",
			collection: collectionName + "_two",
			documents: []*types.Document{
				must.NotFail(types.NewDocument("id", "1")),
				must.NotFail(types.NewDocument("id", "2")),
			},
			docsPerIteration: []int{2},
		},
		{
			name:       "three",
			collection: collectionName + "_three",
			documents: []*types.Document{
				must.NotFail(types.NewDocument("id", "1")),
				must.NotFail(types.NewDocument("id", "2")),
				must.NotFail(types.NewDocument("id", "3")),
			},
			docsPerIteration: []int{2, 1},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tx, err := pool.Begin(ctx)
			require.NoError(t, err)

			sp := &SQLParam{DB: dbName, Collection: tc.collection}

			for _, doc := range tc.documents {
				require.NoError(t, InsertDocument(ctx, tx, sp, doc))
			}

			fetchedChan, err := pool.QueryDocuments(ctx, tx, sp)
			require.NoError(t, err)

			iter := 0
			for {
				fetched, ok := <-fetchedChan
				if !ok {
					break
				}

				assert.NoError(t, fetched.Err)
				assert.Equal(t, tc.docsPerIteration[iter], len(fetched.Docs))
				iter++
			}
			assert.Equal(t, len(tc.docsPerIteration), iter)

			require.NoError(t, tx.Commit(ctx))
		})
	}

	// Special case: cancel context before reading from channel.
	t.Run("cancel_context", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		sp := &SQLParam{DB: dbName, Collection: collectionName + "_cancel"}

		for i := 1; i <= FetchedChannelBufSize*FetchedSliceCapacity+1; i++ {
			require.NoError(t, InsertDocument(ctx, tx, sp,
				must.NotFail(types.NewDocument("id", fmt.Sprintf("%d", i))),
			))
		}

		ctx, cancel := context.WithCancel(context.Background())
		fetchedChan, err := pool.QueryDocuments(ctx, pool, sp)
		cancel()
		require.NoError(t, err)

		<-ctx.Done()
		countDocs := 0
		for {
			fetched, ok := <-fetchedChan
			if !ok {
				break
			}

			countDocs += len(fetched.Docs)
		}
		require.Less(t, countDocs, FetchedChannelBufSize*FetchedSliceCapacity+1)

		require.ErrorIs(t, tx.Rollback(ctx), context.Canceled)
	})

	// Special case: querying a non-existing collection.
	t.Run("non-existing_collection", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		sp := &SQLParam{DB: dbName, Collection: collectionName + "_non-existing"}
		fetchedChan, err := pool.QueryDocuments(context.Background(), tx, sp)
		require.NoError(t, err)
		res, ok := <-fetchedChan
		require.False(t, ok)
		require.Nil(t, res.Docs)
		require.Nil(t, res.Err)

		require.NoError(t, tx.Commit(ctx))
	})
}
