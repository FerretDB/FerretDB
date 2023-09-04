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
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestStats(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	r, err := metadata.NewRegistry("file:./?mode=memory", testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName := t.Name()
	db := newDatabase(r, dbName)

	t.Cleanup(func() {
		db.Close()
	})

	var res *backends.DatabaseStatsResult

	t.Run("NonExistingDatabase", func(t *testing.T) {
		res, err = db.Stats(ctx, new(backends.DatabaseStatsParams))
		require.NoError(t, err)
		require.Equal(t, new(backends.DatabaseStatsResult), res)
	})

	collectionName := t.Name()
	err = db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: collectionName})
	require.NoError(t, err)
	require.NotNil(t, db)

	t.Cleanup(func() {
		r.DatabaseDrop(ctx, dbName)
	})

	t.Run("EmptyCollection", func(t *testing.T) {
		res, err = db.Stats(ctx, new(backends.DatabaseStatsParams))
		require.NoError(t, err)
		require.NotZero(t, res.SizeTotal)
		require.NotZero(t, res.CountCollections)
		require.NotZero(t, res.SizeCollections)
		require.Zero(t, res.CountObjects)
	})
}
