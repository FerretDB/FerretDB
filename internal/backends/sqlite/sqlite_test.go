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

	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCollectionsStats(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	dbName := testutil.DatabaseName(t)
	r, err := metadata.NewRegistry("file:./?mode=memory", testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	d, err := r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)

	cNames := []string{"collectionOne", "collectionTwo"}
	colls := make([]*metadata.Collection, len(cNames))

	for i, cName := range cNames {
		_, err = r.CollectionCreate(ctx, dbName, cName)
		require.NoError(t, err)
		colls[i] = r.CollectionGet(ctx, dbName, cName)
	}

	t.Run("RefreshTwoTables", func(t *testing.T) {
		_, err = collectionsStats(ctx, d, colls, true)
		require.NoError(t, err)
	})

	t.Run("RefreshNonExistentFirstTable", func(t *testing.T) {
		_, err = collectionsStats(ctx, d, []*metadata.Collection{{TableName: "non-existent"}, colls[0]}, true)
		require.ErrorContains(t, err, "SQL logic error: no such table: non-existent")
	})

	t.Run("RefreshNonExistentSecondTable", func(t *testing.T) {
		_, err = collectionsStats(ctx, d, []*metadata.Collection{colls[0], {TableName: "non-existent"}}, true)
		require.ErrorContains(t, err, "SQL logic error: no such table: non-existent")
	})

	t.Run("NoRefreshWithNonExistentTable", func(t *testing.T) {
		_, err = collectionsStats(ctx, d, []*metadata.Collection{{TableName: "non-existent"}, colls[0]}, false)
		require.NoError(t, err)
	})
}
