package tigrisdb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

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

		inserted := make([]*types.Document, 0)
		for i := 0; i < 10; i++ {
			doc := must.NotFail(types.NewDocument("_id", int64(i)))
			err := tdb.InsertDocument(ctx, dbName, collName, doc)
			require.NoError(t, err)

			inserted = append(inserted, doc)
		}

		iter, err := tdb.QueryDocuments(ctx, &FetchParam{
			DB:         dbName,
			Collection: collName,
		})
		require.NoError(t, err)

		require.NotNil(t, iter)

		defer iter.Close()

		queried := make([]*types.Document, 0)

		i := 0
		for {
			n, doc, err := iter.Next()
			if err != nil {
				if errors.Is(err, iterator.ErrIteratorDone) {
					break
				}

				require.NoError(t, err)
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

		iter, err := tdb.QueryDocuments(ctx, &FetchParam{
			DB:         dbName,
			Collection: collName,
		})
		require.NoError(t, err)

		require.Nil(t, iter)
	})

	t.Run("CollectionEmpty", func(t *testing.T) {
		t.Parallel()

		dbName, collName, ctx, tdb := setup(t)

		_, err := tdb.CreateCollectionIfNotExist(ctx, dbName, collName, driver.Schema(strings.TrimSpace(fmt.Sprintf(
			`{"title": "%s","properties": {"_id": {"type": "string","format": "byte"}},"primary_key": ["_id"]}`,
			collName,
		))))
		require.NoError(t, err)

		iter, err := tdb.QueryDocuments(ctx, &FetchParam{
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

		iter, err := tdb.QueryDocuments(ctx, &FetchParam{
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

		inserted := make([]*types.Document, 0)
		for i := 0; i < 10; i++ {
			doc := must.NotFail(types.NewDocument("_id", int64(i)))
			err := tdb.InsertDocument(ctx, dbName, collName, doc)
			require.NoError(t, err)

			inserted = append(inserted, doc)
		}

		iter, err := tdb.QueryDocuments(ctx, &FetchParam{
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
