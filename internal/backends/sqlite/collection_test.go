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

package sqlite

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCollectionInsertAllQueryExplain(t *testing.T) {
	// remove this test
	// TODO https://github.com/FerretDB/FerretDB/issues/3181

	t.Parallel()

	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	b, err := NewBackend(&NewBackendParams{URI: "file:" + t.TempDir() + "/", L: testutil.Logger(t), P: sp})
	require.NoError(t, err)
	t.Cleanup(b.Close)

	db, err := b.Database(testutil.DatabaseName(t))
	require.NoError(t, err)

	collName, cappedCollName := testutil.CollectionName(t), testutil.CollectionName(t)+"capped"
	coll, err := db.Collection(collName)
	require.NoError(t, err)

	err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
		Name:       cappedCollName,
		CappedSize: 8192,
	})
	require.NoError(t, err)

	cappedColl, err := db.Collection(cappedCollName)
	require.NoError(t, err)

	insertDocs := []*types.Document{
		must.NotFail(types.NewDocument("_id", int32(2))),
		must.NotFail(types.NewDocument("_id", int32(3))),
		must.NotFail(types.NewDocument("_id", int32(1))),
	}

	_, err = coll.InsertAll(ctx, &backends.InsertAllParams{Docs: insertDocs})
	require.NoError(t, err)

	_, err = cappedColl.InsertAll(ctx, &backends.InsertAllParams{Docs: insertDocs})
	require.NoError(t, err)

	t.Run("CappedCollectionSort", func(t *testing.T) {
		t.Parallel()

		sort := backends.SortField{Key: "_id"}
		queryRes, err := cappedColl.Query(ctx, &backends.QueryParams{Sort: &sort})
		require.NoError(t, err)

		docs, err := iterator.ConsumeValues[struct{}, *types.Document](queryRes.Iter)
		require.NoError(t, err)

		require.Len(t, docs, len(insertDocs))
		for i, doc := range docs {
			assert.NotEmpty(t, doc.RecordID())
			assert.Equal(t, insertDocs[i].Keys(), doc.Keys())
			assert.Equal(t, insertDocs[i].Values(), doc.Values())
		}

		explainRes, err := cappedColl.Explain(ctx, new(backends.ExplainParams))
		require.NoError(t, err)
		assert.True(t, explainRes.SortPushdown)
	})
}
