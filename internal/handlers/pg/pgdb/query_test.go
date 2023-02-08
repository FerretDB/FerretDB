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
	"math"
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

		ctxGet, cancelGet := context.WithCancel(ctx)
		collectionName := testutil.CollectionName(t)

		err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
			if err := InsertDocument(ctx, tx, databaseName, collectionName, doc1); err != nil {
				return err
			}

			if err := InsertDocument(ctx, tx, databaseName, collectionName, doc2); err != nil {
				return err
			}

			sp := &SQLParam{DB: databaseName, Collection: collectionName}
			iter, err := GetDocuments(ctxGet, tx, sp)
			if err != nil {
				return err
			}
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
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			cancelGet()

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("EarlyClose", func(t *testing.T) {
		t.Parallel()

		ctxGet, cancelGet := context.WithCancel(ctx)
		collectionName := testutil.CollectionName(t)

		err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
			if err := InsertDocument(ctx, tx, databaseName, collectionName, doc1); err != nil {
				return err
			}

			sp := &SQLParam{DB: databaseName, Collection: collectionName}
			iter, err := GetDocuments(ctxGet, tx, sp)
			if err != nil {
				return err
			}
			require.NotNil(t, iter)

			iter.Close()

			n, doc, err := iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			cancelGet()

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
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

		err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
			if err := InsertDocument(ctx, tx, databaseName, collectionName, doc1); err != nil {
				return err
			}

			sp := &SQLParam{DB: databaseName, Collection: collectionName}
			iter, err := GetDocuments(ctxGet, tx, sp)
			if err != nil {
				return err
			}
			require.NotNil(t, iter)

			cancelGet()

			n, doc, err := iter.Next()
			require.ErrorIs(t, err, context.Canceled, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still canceled
			n, doc, err = iter.Next()
			require.ErrorIs(t, err, context.Canceled, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			iter.Close()

			// done now
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("EmptyCollection", func(t *testing.T) {
		t.Parallel()

		ctxGet, cancelGet := context.WithCancel(ctx)
		collectionName := testutil.CollectionName(t)

		err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
			if err := CreateCollection(ctx, tx, databaseName, collectionName); err != nil {
				return err
			}

			sp := &SQLParam{DB: databaseName, Collection: collectionName}
			iter, err := GetDocuments(ctxGet, tx, sp)
			if err != nil {
				return err
			}
			require.NotNil(t, iter)

			defer iter.Close()

			n, doc, err := iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			cancelGet()

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})

	t.Run("NonExistentCollection", func(t *testing.T) {
		t.Parallel()

		ctxGet, cancelGet := context.WithCancel(ctx)
		collectionName := testutil.CollectionName(t)

		err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
			sp := &SQLParam{DB: databaseName, Collection: collectionName}
			iter, err := GetDocuments(ctxGet, tx, sp)
			if err != nil {
				return err
			}
			require.NotNil(t, iter)

			defer iter.Close()

			n, doc, err := iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			cancelGet()

			// still done
			n, doc, err = iter.Next()
			require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
			assert.Zero(t, n)
			assert.Nil(t, doc)

			return nil
		})

		require.NoError(t, err)
	})
}

// WHERE clauses occurring frequently in test.
var (
	whereEq          string = " WHERE ((_jsonb->$1)::jsonb = $2)"
	whereEqOrContain string = whereEq + " OR (_jsonb->$1)::jsonb @> $2"
)

func TestPrepareWhereClause(t *testing.T) {
	t.Parallel()
	objectID := types.ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0xff, 0xff, 0xff}

	for name, tc := range map[string]struct {
		filter   *types.Document
		expected string
		skip     string
	}{
		"String": {
			filter:   must.NotFail(types.NewDocument("v", "foo")),
			expected: whereEqOrContain,
		},
		"EmptyString": {
			filter:   must.NotFail(types.NewDocument("v", "")),
			expected: whereEqOrContain,
		},
		"Int32": {
			filter:   must.NotFail(types.NewDocument("v", int32(42))),
			expected: whereEqOrContain,
		},
		"Int64": {
			filter:   must.NotFail(types.NewDocument("v", int64(42))),
			expected: whereEqOrContain,
		},
		"Float64": {
			filter:   must.NotFail(types.NewDocument("v", float64(42.13))),
			expected: whereEqOrContain,
		},
		"MaxFloat64": {
			filter:   must.NotFail(types.NewDocument("v", math.MaxFloat64)),
			expected: whereEqOrContain,
		},
		"Bool": {
			filter: must.NotFail(types.NewDocument("v", true)),
		},
		"Comment": {
			filter: must.NotFail(types.NewDocument("$comment", "I'm comment")),
		},
		"ObjectID": {
			filter:   must.NotFail(types.NewDocument("v", objectID)),
			expected: whereEqOrContain,
		},
		"IDObjectID": {
			filter:   must.NotFail(types.NewDocument("_id", objectID)),
			expected: whereEq,
		},
		"IDString": {
			filter:   must.NotFail(types.NewDocument("_id", "foo")),
			expected: whereEq,
		},
		"IDDotNotation": {
			filter:   must.NotFail(types.NewDocument("_id.doc", "foo")),
			expected: " WHERE ((_jsonb#>$1)::jsonb = $2)",
		},
		"DotNotation": {
			filter:   must.NotFail(types.NewDocument("v.doc", "foo")),
			expected: " WHERE ((_jsonb#>$1)::jsonb = $2) OR (_jsonb#>$1)::jsonb @> $2",
		},
		"DotNotationArrayIndex": {
			filter:   must.NotFail(types.NewDocument("v.arr.0", "foo")),
			expected: " WHERE ((_jsonb#>$1)::jsonb = $2) OR (_jsonb#>$1)::jsonb @> $2",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			actual, _ := prepareWhereClause(tc.filter)

			assert.Equal(t, tc.expected, actual)
		})
	}
}
