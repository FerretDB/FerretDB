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

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCollectionStats(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	b, err := NewBackend(&NewBackendParams{URI: "file:./?mode=memory", L: testutil.Logger(t), P: sp})
	require.NoError(t, err)
	t.Cleanup(b.Close)

	db, err := b.Database(testutil.DatabaseName(t))
	require.NoError(t, err)

	cNames := []string{"collectionOne", "collectionTwo"}
	for _, cName := range cNames {
		err = db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: cName})
		require.NoError(t, err)
	}

	c, err := db.Collection(cNames[0])
	require.NoError(t, err)

	_, err = c.InsertAll(ctx, &backends.InsertAllParams{
		Docs: []*types.Document{must.NotFail(types.NewDocument("_id", types.NewObjectID()))},
	})
	require.NoError(t, err)

	dbStatsRes, err := db.Stats(ctx, new(backends.DatabaseStatsParams))
	require.NoError(t, err)

	res, err := c.Stats(ctx, new(backends.CollectionStatsParams))
	require.NoError(t, err)
	require.NotZero(t, res.SizeTotal)
	require.Less(t, res.SizeTotal, dbStatsRes.SizeTotal)
	require.NotZero(t, res.SizeCollection)
	require.Less(t, res.SizeCollection, dbStatsRes.SizeCollections)
	require.Equal(t, res.CountObjects, int64(1))
}
