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
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestDatabaseStats(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	b, err := NewBackend(&NewBackendParams{URI: "file:./?mode=memory", L: testutil.Logger(t)})
	require.NoError(t, err)
	t.Cleanup(b.Close)

	db, err := b.Database(testutil.DatabaseName(t))
	require.NoError(t, err)
	t.Cleanup(db.Close)

	t.Run("NonExistingDatabase", func(t *testing.T) {
		res, err := db.Stats(ctx, new(backends.DatabaseStatsParams))
		require.NoError(t, err)
		require.Equal(t, new(backends.DatabaseStatsResult), res)
	})

	cNames := []string{"collectionOne", "collectionTwo"}
	for _, cName := range cNames {
		err = db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: cName})
		require.NoError(t, err)
		require.NotNil(t, db)
	}

	t.Run("DatabaseWithCollections", func(t *testing.T) {
		res, err := db.Stats(ctx, new(backends.DatabaseStatsParams))
		require.NoError(t, err)
		require.NotZero(t, res.SizeTotal)
		require.Equal(t, res.CountCollections, int64(len(cNames)))
		require.NotZero(t, res.SizeCollections)
		require.Zero(t, res.CountObjects)
	})
}
