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

package postgresql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCollectionInsertAllQueryExplain(t *testing.T) {
	// remove this test
	// TODO https://github.com/FerretDB/FerretDB/issues/3181

	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	params := NewBackendParams{
		URI: "postgres://username:password@127.0.0.1:5432/ferretdb",
		L:   testutil.Logger(t),
		P:   sp,
	}
	b, err := NewBackend(&params)
	require.NoError(t, err)
	t.Cleanup(b.Close)

	dbName := testutil.DatabaseName(t)
	db, err := b.Database(dbName)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = b.DropDatabase(ctx, &backends.DropDatabaseParams{Name: dbName})
		require.NoError(t, err)
	})

	cName := testutil.CollectionName(t)
	err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
		Name:       cName,
		CappedSize: 8192,
	})
	require.NoError(t, err)

	cappedColl, err := db.Collection(cName)
	require.NoError(t, err)

	insertDocs := []*types.Document{
		must.NotFail(types.NewDocument("_id", int32(2))),
		must.NotFail(types.NewDocument("_id", int32(3))),
		must.NotFail(types.NewDocument("_id", int32(1))),
	}

	_, err = cappedColl.InsertAll(ctx, &backends.InsertAllParams{Docs: insertDocs})
	require.NoError(t, err)

	t.Run("CappedCollectionSortAsc", func(t *testing.T) {
		t.Parallel()

		sort := backends.SortField{Key: "_id"}
		queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Sort: &sort})
		require.NoError(t, err)

		docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
		require.NoError(t, err)
		require.Len(t, docs, len(insertDocs))

		// inserted doc is frozen, queried doc is not frozen hence compare each value
		assert.Equal(t, insertDocs[2].RecordID(), docs[0].RecordID())
		assert.Equal(t, insertDocs[2].Values(), docs[0].Values())

		assert.Equal(t, insertDocs[0].RecordID(), docs[1].RecordID())
		assert.Equal(t, insertDocs[0].Values(), docs[1].Values())

		assert.Equal(t, insertDocs[1].RecordID(), docs[2].RecordID())
		assert.Equal(t, insertDocs[1].Values(), docs[2].Values())

		explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{Sort: &sort})
		require.NoError(t, err)
		assert.True(t, explainRes.SortPushdown)
	})

	t.Run("CappedCollectionSortDesc", func(t *testing.T) {
		t.Parallel()

		sort := backends.SortField{Key: "_id", Descending: true}
		queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Sort: &sort})
		require.NoError(t, err)

		docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
		require.NoError(t, err)
		require.Len(t, docs, len(insertDocs))

		// inserted doc is frozen, queried doc is not frozen hence compare each value
		assert.Equal(t, insertDocs[1].RecordID(), docs[0].RecordID())
		assert.Equal(t, insertDocs[1].Values(), docs[0].Values())

		assert.Equal(t, insertDocs[0].RecordID(), docs[1].RecordID())
		assert.Equal(t, insertDocs[0].Values(), docs[1].Values())

		assert.Equal(t, insertDocs[2].RecordID(), docs[2].RecordID())
		assert.Equal(t, insertDocs[2].Values(), docs[2].Values())

		explainRes, err := cappedColl.Explain(ctx, &backends.ExplainParams{Sort: &sort})
		require.NoError(t, err)
		assert.True(t, explainRes.SortPushdown)
	})
}
