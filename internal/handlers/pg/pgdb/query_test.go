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
	"fmt"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestQueryDocuments(t *testing.T) {
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

	cases := []struct {
		name       string
		collection string
		documents  []*types.Document

		// docsPerIteration represents how many documents should be fetched per each iteration,
		// use len(docsPerIteration) as the amount of fetch iterations.
		docsPerIteration []int
	}{{
		name:             "empty",
		collection:       collectionName,
		documents:        []*types.Document{},
		docsPerIteration: []int{},
	}, {
		name:             "one",
		collection:       collectionName + "_one",
		documents:        []*types.Document{must.NotFail(types.NewDocument("_id", "foo", "id", "1"))},
		docsPerIteration: []int{1},
	}, {
		name:       "two",
		collection: collectionName + "_two",
		documents: []*types.Document{
			must.NotFail(types.NewDocument("_id", "foo", "id", "1")),
			must.NotFail(types.NewDocument("_id", "foo", "id", "2")),
		},
		docsPerIteration: []int{2},
	}, {
		name:       "three",
		collection: collectionName + "_three",
		documents: []*types.Document{
			must.NotFail(types.NewDocument("_id", "foo", "id", "1")),
			must.NotFail(types.NewDocument("_id", "foo", "id", "2")),
			must.NotFail(types.NewDocument("_id", "foo", "id", "3")),
		},
		docsPerIteration: []int{2, 1},
	}}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tx, err := pool.Begin(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			for _, doc := range tc.documents {
				require.NoError(t, InsertDocument(ctx, tx, databaseName, tc.collection, doc))
			}

			sp := &SQLParam{DB: databaseName, Collection: tc.collection}
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
		defer tx.Rollback(ctx)

		for i := 1; i <= FetchedChannelBufSize*FetchedSliceCapacity+1; i++ {
			require.NoError(t, InsertDocument(ctx, tx, databaseName, collectionName+"_cancel",
				must.NotFail(types.NewDocument("_id", "foo", "id", fmt.Sprintf("%d", i))),
			))
		}

		sp := &SQLParam{DB: databaseName, Collection: collectionName + "_cancel"}
		ctx, cancel := context.WithCancel(ctx)
		fetchedChan, err := pool.QueryDocuments(ctx, tx, sp)
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
		defer tx.Rollback(ctx)

		sp := &SQLParam{DB: databaseName, Collection: collectionName + "_non-existing"}
		fetchedChan, err := pool.QueryDocuments(ctx, tx, sp)
		require.NoError(t, err)
		res, ok := <-fetchedChan
		require.False(t, ok)
		require.Nil(t, res.Docs)
		require.Nil(t, res.Err)

		require.NoError(t, tx.Commit(ctx))
	})
}
