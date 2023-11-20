package postgresql

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"

	"github.com/FerretDB/FerretDB/internal/util/state"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCappedCollectionInsertAllDeleteAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	params := NewBackendParams{
		URI: testutil.TestPostgreSQLURI(t, ctx, ""),
		L:   testutil.Logger(t),
		P:   sp,
	}

	// create a backend without backendContract wrap
	r, err := metadata.NewRegistry(params.URI, params.L, params.P)
	require.NoError(t, err)

	b := backend{r: r}
	t.Cleanup(b.Close)

	dbName := testutil.DatabaseName(t)
	collName := testutil.CollectionName(t)

	db, err := b.Database(dbName)
	require.NoError(t, err)

	err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
		Name:       collName,
		CappedSize: 8192,
	})
	require.NoError(t, err)

	coll, err := db.Collection(collName)
	require.NoError(t, err)

	doc1 := must.NotFail(types.NewDocument("_id", int32(1)))
	doc1.SetRecordID(1)

	docMax := must.NotFail(types.NewDocument("_id", int32(2)))
	docMax.SetRecordID(math.MaxInt64)

	docMaxUint := must.NotFail(types.NewDocument("_id", int32(3)))
	docMaxUint.SetRecordID(math.MaxUint64)

	docEpochalypse := must.NotFail(types.NewDocument("_id", int32(4)))
	date := time.Date(2038, time.January, 19, 3, 14, 6, 0, time.UTC)
	docEpochalypse.SetRecordID(types.Timestamp(date.Unix()))

	insertDocs := []*types.Document{doc1, docMax, docMaxUint, docEpochalypse}

	_, err = coll.InsertAll(ctx, &backends.InsertAllParams{Docs: insertDocs})
	require.NoError(t, err)

	res, err := coll.Query(ctx, nil)
	require.NoError(t, err)

	docs, err := iterator.ConsumeValues[struct{}, *types.Document](res.Iter)
	require.NoError(t, err)
	require.Equal(t, 4, len(docs))

	assert.Equal(t, doc1.RecordID(), docs[0].RecordID())
	assert.Equal(t, docMax.RecordID(), docs[1].RecordID())
	assert.Equal(t, docMaxUint.RecordID(), docs[2].RecordID())
	assert.Equal(t, docEpochalypse.RecordID(), docs[3].RecordID())

	deletePaarams := &backends.DeleteAllParams{
		RecordIDs: []types.Timestamp{docMax.RecordID(), docMaxUint.RecordID(), docEpochalypse.RecordID()},
	}
	del, err := coll.DeleteAll(ctx, deletePaarams)
	require.NoError(t, err)
	require.Equal(t, int32(3), del.Deleted)

	res, err = coll.Query(ctx, nil)
	require.NoError(t, err)

	docs, err = iterator.ConsumeValues[struct{}, *types.Document](res.Iter)
	require.NoError(t, err)
	// assertEqualRecordID(t, []*types.Document{doc1}, docs)
}
