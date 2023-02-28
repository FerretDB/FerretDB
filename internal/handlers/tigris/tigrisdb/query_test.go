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

package tigrisdb

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestQueryDocuments(t *testing.T) {
	t.Parallel()

	t.Run("QueryDocuments", func(t *testing.T) {
		t.Parallel()

		dbName, collName, ctx, tdb := setup(t)

		var inserted []*types.Document
		for i := 0; i < 10; i++ {
			doc := must.NotFail(types.NewDocument("_id", int64(i)))
			err := tdb.InsertDocument(ctx, dbName, collName, doc)
			require.NoError(t, err)

			inserted = append(inserted, doc)
		}

		iter, err := tdb.QueryDocuments(ctx, &QueryParams{
			DB:         dbName,
			Collection: collName,
		})
		require.NoError(t, err)

		require.NotNil(t, iter)

		defer iter.Close()

		var queried []*types.Document

		i := 0
		for {
			var n int
			var doc *types.Document

			n, doc, err = iter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			require.NoError(t, err)
			require.Equal(t, i, n)

			queried = append(queried, doc)
			i++
		}

		require.Equal(t, len(inserted), len(queried))

		n, doc, err := iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		require.Nil(t, doc)
		require.Zero(t, n)

		n, doc, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		require.Nil(t, doc)
		require.Zero(t, n)
	})

	t.Run("CollectionNotExist", func(t *testing.T) {
		t.Parallel()

		dbName, collName, ctx, tdb := setup(t)

		_, err := tdb.createDatabaseIfNotExists(ctx, dbName)
		require.NoError(t, err)

		iter, err := tdb.QueryDocuments(ctx, &QueryParams{
			DB:         dbName,
			Collection: collName,
		})
		require.NoError(t, err)

		require.NotNil(t, iter)

		n, doc, err := iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		require.Nil(t, doc)
		require.Zero(t, n)
	})

	t.Run("CollectionEmpty", func(t *testing.T) {
		t.Parallel()

		dbName, collName, ctx, tdb := setup(t)

		_, err := tdb.CreateCollectionIfNotExist(ctx, dbName, collName, driver.Schema(strings.TrimSpace(fmt.Sprintf(
			`{"title": "%s","properties": {"_id": {"type": "string","format": "byte"}},"primary_key": ["_id"]}`,
			collName,
		))))
		require.NoError(t, err)

		iter, err := tdb.QueryDocuments(ctx, &QueryParams{
			DB:         dbName,
			Collection: collName,
		})
		require.NoError(t, err)

		require.NotNil(t, iter)

		n, doc, err := iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		require.Nil(t, doc)
		require.Zero(t, n)
	})

	t.Run("EarlyClose", func(t *testing.T) {
		t.Parallel()

		dbName, collName, ctx, tdb := setup(t)

		for i := 0; i < 10; i++ {
			doc := must.NotFail(types.NewDocument("_id", int64(i)))
			err := tdb.InsertDocument(ctx, dbName, collName, doc)
			require.NoError(t, err)
		}

		iter, err := tdb.QueryDocuments(ctx, &QueryParams{
			DB:         dbName,
			Collection: collName,
		})
		require.NoError(t, err)

		require.NotNil(t, iter)

		iter.Close()

		n, doc, err := iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		require.Nil(t, doc)
		require.Zero(t, n)

		n, doc, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err)
		require.Nil(t, doc)
		require.Zero(t, n)
	})

	t.Run("CancelContext", func(t *testing.T) {
		t.Parallel()

		dbName, collName, ctx, tdb := setup(t)

		ctx, cancel := context.WithCancel(ctx)

		for i := 0; i < 10; i++ {
			doc := must.NotFail(types.NewDocument("_id", int64(i)))
			err := tdb.InsertDocument(ctx, dbName, collName, doc)
			require.NoError(t, err)
		}

		iter, err := tdb.QueryDocuments(ctx, &QueryParams{
			DB:         dbName,
			Collection: collName,
		})
		require.NoError(t, err)

		require.NotNil(t, iter)

		cancel()

		n, doc, err := iter.Next()
		require.ErrorIs(t, err, context.Canceled, "%v", err)
		require.Zero(t, n)
		require.Nil(t, doc)

		// still canceled
		n, doc, err = iter.Next()
		require.ErrorIs(t, err, context.Canceled, "%v", err)
		require.Zero(t, n)
		require.Nil(t, doc)

		iter.Close()

		// done now
		n, doc, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
		require.Zero(t, n)
		require.Nil(t, doc)

		// still done
		n, doc, err = iter.Next()
		require.Equal(t, iterator.ErrIteratorDone, err, "%v", err)
		require.Zero(t, n)
		require.Nil(t, doc)
	})
}

func TestBuildFilter(t *testing.T) {
	t.Parallel()
	objectID := types.ObjectID{0x62, 0x56, 0xc5, 0xba, 0x0b, 0xad, 0xc0, 0xff, 0xee, 0xff, 0xff, 0xff}

	for name, tc := range map[string]struct {
		filter   *types.Document
		expected string
		skip     string // defaults to `{}`
	}{
		"IDObjectID": {
			filter:   must.NotFail(types.NewDocument("_id", objectID)),
			expected: `{"_id":"YlbFugutwP/u////"}`,
		},
		"IDString": {
			filter:   must.NotFail(types.NewDocument("_id", "foo")),
			expected: `{"_id":"foo"}`,
		},
		"IDDotNotation": {
			filter:   must.NotFail(types.NewDocument("_id.doc", "foo")),
			expected: `{"_id.doc":"foo"}`,
		},

		"DotNotation": {
			filter:   must.NotFail(types.NewDocument("v.doc", "foo")),
			expected: `{"v.doc":"foo"}`,
		},
		"DotNotationArrayIndex": {
			filter: must.NotFail(types.NewDocument("v.arr.0", "foo")),
		},

		"ImplicitString": {
			filter:   must.NotFail(types.NewDocument("v", "foo")),
			expected: `{"v":"foo"}`,
		},
		"ImplicitEmptyString": {
			filter:   must.NotFail(types.NewDocument("v", "")),
			expected: `{"v":""}`,
			skip:     "https://github.com/FerretDB/FerretDB/issues/1940",
		},
		"ImplicitInt32": {
			filter:   must.NotFail(types.NewDocument("v", int32(42))),
			expected: `{"v":42}`,
		},
		"ImplicitInt64": {
			filter:   must.NotFail(types.NewDocument("v", int64(42))),
			expected: `{"v":42}`,
		},
		"ImplicitFloat64": {
			filter:   must.NotFail(types.NewDocument("v", float64(42.13))),
			expected: `{"v":42.13}`,
		},
		"ImplicitMaxFloat64": {
			filter:   must.NotFail(types.NewDocument("v", math.MaxFloat64)),
			expected: `{"v":1.7976931348623157e+308}`,
		},
		"ImplicitBool": {
			filter: must.NotFail(types.NewDocument("v", true)),
		},
		"ImplicitObjectID": {
			filter:   must.NotFail(types.NewDocument("v", objectID)),
			expected: `{"v":"YlbFugutwP/u////"}`,
		},

		"EqString": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", "foo")),
			)),
			expected: `{"v":"foo"}`,
		},
		"EqEmptyString": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", "foo")),
			)),
			expected: `{"v":""}`,
			skip:     "https://github.com/FerretDB/FerretDB/issues/1940",
		},
		"EqInt32": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", int32(42))),
			)),
			expected: `{"v":42}`,
		},
		"EqInt64": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", int64(42))),
			)),
			expected: `{"v":42}`,
		},
		"EqFloat64": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", float64(42.13))),
			)),
			expected: `{"v":42.13}`,
		},
		"EqMaxFloat64": {
			filter: must.NotFail(types.NewDocument(
				"v", must.NotFail(types.NewDocument("$eq", math.MaxFloat64)),
			)),
			expected: `{"v":1.7976931348623157e+308}`,
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
			expected: `{"v":"YlbFugutwP/u////"}`,
		},

		"Comment": {
			filter: must.NotFail(types.NewDocument("$comment", "I'm comment")),
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			// replace default value with default json
			if tc.expected == "" {
				tc.expected = "{}"
			}

			actual, err := BuildFilter(tc.filter)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, actual)
		})
	}
}

func setup(t *testing.T) (string, string, context.Context, *TigrisDB) {
	t.Helper()

	dbName := testutil.DatabaseName(t)
	collName := testutil.CollectionName(t)

	ctx := testutil.Ctx(t)
	cfg := &config.Driver{
		URL: testutil.TigrisURL(t),
	}

	logger := testutil.Logger(t, zap.NewAtomicLevelAt(zap.DebugLevel))
	tdb, err := New(ctx, cfg, logger)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, e := tdb.Driver.DeleteProject(ctx, dbName)
		require.NoError(t, e)
	})

	return dbName, collName, ctx, tdb
}
