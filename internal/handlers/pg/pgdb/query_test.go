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

		err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
			err := InsertDocument(ctx, tx, databaseName, collection, expectedDoc)
			if err != nil {
				return err
			}

			sp := &SQLParam{DB: databaseName, Collection: collection}
			iter, err := GetDocuments(ctx, tx, sp)
			require.NoError(t, err)
			require.NotNil(t, iter)

			defer iter.Close()

			n, doc, err := iter.Next()
			assert.NoError(t, err)
			assert.Equal(t, uint32(0), n)
			assert.Equal(t, expectedDoc, doc)

			n, doc, err = iter.Next()
			assert.Equal(t, iterator.ErrIteratorDone, err)
			assert.Equal(t, uint32(0), n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("cancel-context", func(t *testing.T) {
		t.Parallel()

		collection := collectionName + "-two"
		expectedDocs := []*types.Document{
			must.NotFail(types.NewDocument("_id", "bar", "id", "1")),
			must.NotFail(types.NewDocument("_id", "foo", "id", "2")),
		}

		err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
			for _, doc := range expectedDocs {
				err := InsertDocument(ctx, tx, databaseName, collection, doc)
				if err != nil {
					return err
				}
			}

			ctxTest, cancel := context.WithCancel(ctx)
			sp := &SQLParam{DB: databaseName, Collection: collection}
			iter, err := GetDocuments(ctxTest, tx, sp)
			require.NoError(t, err)
			require.NotNil(t, iter)

			defer iter.Close()

			n, doc, err := iter.Next()
			assert.NoError(t, err)
			assert.Equal(t, uint32(0), n)
			assert.Equal(t, expectedDocs[0], doc)

			cancel()
			n, doc, err = iter.Next()
			assert.Equal(t, context.Canceled, err)
			assert.Equal(t, uint32(0), n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("empty-collection", func(t *testing.T) {
		t.Parallel()

		collection := collectionName + "-empty"

		err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
			err := CreateCollection(ctx, tx, databaseName, collection)
			if err != nil {
				return err
			}

			sp := &SQLParam{DB: databaseName, Collection: collection}
			iter, err := GetDocuments(ctx, tx, sp)
			require.NoError(t, err)
			require.NotNil(t, iter)

			defer iter.Close()

			n, doc, err := iter.Next()
			assert.Equal(t, iterator.ErrIteratorDone, err)
			assert.Equal(t, uint32(0), n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("non-existent-collection", func(t *testing.T) {
		t.Parallel()

		tx, err := pool.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		sp := &SQLParam{DB: databaseName, Collection: collectionName + "-non-existent"}
		iter, err := GetDocuments(ctx, tx, sp)
		require.NoError(t, err)
		require.NotNil(t, iter)

		defer iter.Close()

		n, doc, err := iter.Next()
		assert.Equal(t, iterator.ErrIteratorDone, err)
		assert.Equal(t, uint32(0), n)
		assert.Nil(t, doc)

		require.NoError(t, tx.Commit(ctx))
	})
}

func TestPrepareWhereClause(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		sqlFilters  *types.Document
		whereClause string
		args        []any
	}{
		"QueryID": {
			sqlFilters:  must.NotFail(types.NewDocument("_id", "foo")),
			whereClause: " WHERE ((_jsonb->$1)::jsonb = $2)",
			args:        []any{"_id", `"foo"`},
		},
		"QueryIDNested": {
			sqlFilters:  must.NotFail(types.NewDocument("_id.v", "foo")),
			whereClause: " WHERE ((_jsonb#>$1)::jsonb = $2)",
			args:        []any{[]string{"_id", "v"}, `"foo"`},
		},
		"QueryTwoFields": {
			sqlFilters: must.NotFail(types.NewDocument("v", "foo", "v2", "bar")),
			whereClause: " WHERE ((_jsonb->$1)::jsonb = $2) OR (_jsonb->$1)::jsonb @> $2" +
				" AND ((_jsonb->$3)::jsonb = $4) OR (_jsonb->$3)::jsonb @> $4",
			args: []any{"v", `"foo"`, "v2", `"bar"`},
		},
		"DeepNestedDocument": {
			sqlFilters:  must.NotFail(types.NewDocument("v.foo.bar", "baz")),
			whereClause: " WHERE ((_jsonb#>$1)::jsonb = $2) OR (_jsonb#>$1)::jsonb @> $2",
			args:        []any{[]string{"v", "foo", "bar"}, `"baz"`},
		},
		"QueryArray": {
			sqlFilters:  must.NotFail(types.NewDocument("v.0", "foo")),
			whereClause: " WHERE ((_jsonb#>$1)::jsonb = $2) OR (_jsonb#>$1)::jsonb @> $2",
			args:        []any{[]string{"v", "0"}, `"foo"`},
		},
	}
	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, got1 := prepareWhereClause(tc.sqlFilters)
			assert.Equalf(t, tc.whereClause, got, "prepareWhereClause(%v)", tc.sqlFilters)
			assert.Equalf(t, tc.args, got1, "prepareWhereClause(%v)", tc.sqlFilters)
		})
	}
}
