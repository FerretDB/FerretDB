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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
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
				return lazyerrors.Error(err)
			}

			if err := InsertDocument(ctx, tx, databaseName, collectionName, doc2); err != nil {
				return lazyerrors.Error(err)
			}

			qp := &QueryParam{DB: databaseName, Collection: collectionName}
			iter, err := QueryDocuments(ctxGet, tx, qp)
			if err != nil {
				return lazyerrors.Error(err)
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
				return lazyerrors.Error(err)
			}

			qp := &QueryParam{DB: databaseName, Collection: collectionName}
			iter, err := QueryDocuments(ctxGet, tx, qp)
			if err != nil {
				return lazyerrors.Error(err)
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
				return lazyerrors.Error(err)
			}

			qp := &QueryParam{DB: databaseName, Collection: collectionName}
			iter, err := QueryDocuments(ctxGet, tx, qp)
			if err != nil {
				return lazyerrors.Error(err)
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
				return lazyerrors.Error(err)
			}

			qp := &QueryParam{DB: databaseName, Collection: collectionName}
			iter, err := QueryDocuments(ctxGet, tx, qp)
			if err != nil {
				return lazyerrors.Error(err)
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
			qp := &QueryParam{DB: databaseName, Collection: collectionName}
			iter, err := QueryDocuments(ctxGet, tx, qp)
			if err != nil {
				return lazyerrors.Error(err)
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

func TestPrepareWhereClause(t *testing.T) {
	t.Parallel()
	objectID := types.ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0xff, 0xff, 0xff}

	// WHERE clauses occurring frequently in tests
	whereEq := " WHERE (_jsonb->$1)::jsonb = $2"
	whereContain := " WHERE (_jsonb->$1)::jsonb @> $2"

	for name, tc := range map[string]struct {
		filter   *types.Document
		expected string
	}{
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
			expected: " WHERE (_jsonb#>$1)::jsonb = $2",
		},

		"DotNotation": {
			filter:   must.NotFail(types.NewDocument("v.doc", "foo")),
			expected: " WHERE (_jsonb#>$1)::jsonb @> $2",
		},
		"DotNotationArrayIndex": {
			filter:   must.NotFail(types.NewDocument("v.arr.0", "foo")),
			expected: " WHERE (_jsonb#>$1)::jsonb @> $2",
		},

		"Comment": {
			filter: must.NotFail(types.NewDocument("$comment", "I'm comment")),
		},

		"ImplicitString": {
			filter:   must.NotFail(types.NewDocument("v", "foo")),
			expected: whereContain,
		},
		"ImplicitEmptyString": {
			filter:   must.NotFail(types.NewDocument("v", "")),
			expected: whereContain,
		},
		"ImplicitInt32": {
			filter:   must.NotFail(types.NewDocument("v", int32(42))),
			expected: whereContain,
		},
		"ImplicitInt64": {
			filter:   must.NotFail(types.NewDocument("v", int64(42))),
			expected: whereContain,
		},
		"ImplicitFloat64": {
			filter:   must.NotFail(types.NewDocument("v", float64(42.13))),
			expected: whereContain,
		},
		"ImplicitMaxFloat64": {
			filter:   must.NotFail(types.NewDocument("v", math.MaxFloat64)),
			expected: whereContain,
		},
		"ImplicitBool": {
			filter: must.NotFail(types.NewDocument("v", true)),
		},
		"ImplicitObjectID": {
			filter:   must.NotFail(types.NewDocument("v", objectID)),
			expected: whereContain,
		},

		"EqString": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", "foo")),
			)),
			expected: whereContain,
		},
		"EqEmptyString": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", "")),
			)),
			expected: whereContain,
		},
		"EqInt32": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", int32(42))),
			)),
			expected: whereContain,
		},
		"EqInt64": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", int64(42))),
			)),
			expected: whereContain,
		},
		"EqFloat64": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", float64(42.13))),
			)),
			expected: whereContain,
		},
		"EqMaxFloat64": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", math.MaxFloat64)),
			)),
			expected: whereContain,
		},
		"EqBool": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", true)),
			)),
		},
		"EqObjectID": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", objectID)),
			)),
			expected: whereContain,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, _, err := prepareWhereClause(tc.filter)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, actual)
		})
	}
}
