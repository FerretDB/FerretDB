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
	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestStats(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	r, err := metadata.NewRegistry("file:./?mode=memory", testutil.Logger(t))
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName := t.Name()
	db := newDatabase(r, dbName)

	t.Cleanup(func() {
		db.Close()
	})

	t.Run("NonExistingDatabase", func(t *testing.T) {
		var res *backends.StatsResult
		res, err = db.Stats(ctx, new(backends.StatsParams))
		require.NoError(t, err)
		require.Equal(t, new(backends.StatsResult), res)
	})

	collectionOne := "collectionOne"
	err = db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: collectionOne})
	require.NoError(t, err)
	require.NotNil(t, db)

	collectionTwo := "collectionTwo"
	err = db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: collectionTwo})
	require.NoError(t, err)
	require.NotNil(t, db)

	t.Cleanup(func() {
		r.DatabaseDrop(ctx, dbName)
	})

	var dbStatsRes *backends.StatsResult
	t.Run("EmptyCollection", func(t *testing.T) {
		dbStatsRes, err = db.Stats(ctx, new(backends.StatsParams))
		require.NoError(t, err)
		require.NotZero(t, dbStatsRes.SizeTotal)
		require.NotZero(t, dbStatsRes.CountCollections)
		require.NotZero(t, dbStatsRes.SizeCollections)
		require.Zero(t, dbStatsRes.CountObjects)
	})

	t.Run("CollectionOne", func(t *testing.T) {
		res, err := db.Stats(ctx, &backends.StatsParams{Collection: collectionOne})
		require.NoError(t, err)
		require.NotZero(t, res.SizeTotal)
		require.Less(t, res.SizeTotal, dbStatsRes.SizeTotal)
		require.Equal(t, res.CountCollections, int64(1))
		require.NotZero(t, res.SizeCollections)
		require.Less(t, res.SizeCollections, dbStatsRes.SizeCollections)
		require.Zero(t, res.CountObjects)
	})
}
