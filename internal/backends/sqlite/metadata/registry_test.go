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

package metadata

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

// testCollection creates, tests, and drops an unique collection in existing database.
func testCollection(t *testing.T, ctx context.Context, r *Registry, db *fsql.DB, dbName, collectionName string) {
	t.Helper()

	c := r.CollectionGet(ctx, dbName, collectionName)
	require.Nil(t, c)

	created, err := r.CollectionCreate(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.True(t, created)

	created, err = r.CollectionCreate(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.False(t, created)

	c = r.CollectionGet(ctx, dbName, collectionName)
	require.NotNil(t, c)
	require.Equal(t, collectionName, c.Name)

	list, err := r.CollectionList(ctx, dbName)
	require.NoError(t, err)
	require.Contains(t, list, collectionName)

	q := fmt.Sprintf("INSERT INTO %q (%s) VALUES(?)", c.TableName, DefaultColumn)
	doc := `{"$s": {"p": {"_id": {"t": "int"}}, "$k": ["_id"]}, "_id": 42}`
	_, err = db.ExecContext(ctx, q, doc)
	require.NoError(t, err)

	dropped, err := r.CollectionDrop(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.True(t, dropped)

	dropped, err = r.CollectionDrop(ctx, dbName, collectionName)
	require.NoError(t, err)
	require.False(t, dropped)

	c = r.CollectionGet(ctx, dbName, collectionName)
	require.Nil(t, c)
}

func TestCreateDrop(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	r, err := NewRegistry("file:./?mode=memory", testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName := testutil.DatabaseName(t)

	db, err := r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, db)

	t.Cleanup(func() {
		r.DatabaseDrop(ctx, dbName)
	})

	collectionName := testutil.CollectionName(t)

	testCollection(t, ctx, r, db, dbName, collectionName)
}

func TestCreateDropStress(t *testing.T) {
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	for testName, uri := range map[string]string{
		"file":             "file:./",
		"file-immediate":   "file:./?_txlock=immediate",
		"memory":           "file:./?mode=memory",
		"memory-immediate": "file:./?mode=memory&_txlock=immediate",
	} {
		t.Run(testName, func(t *testing.T) {
			r, err := NewRegistry(uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			dbName := "db"
			r.DatabaseDrop(ctx, dbName)

			db, err := r.DatabaseGetOrCreate(ctx, dbName)
			require.NoError(t, err)
			require.NotNil(t, db)

			t.Cleanup(func() {
				r.DatabaseDrop(ctx, dbName)
			})

			var i atomic.Int32

			teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
				collectionName := fmt.Sprintf("collection_%03d", i.Add(1))

				ready <- struct{}{}
				<-start

				testCollection(t, ctx, r, db, dbName, collectionName)
			})
		})
	}
}

func TestCreateSameStress(t *testing.T) {
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	for testName, uri := range map[string]string{
		"file":             "file:./",
		"file-immediate":   "file:./?_txlock=immediate",
		"memory":           "file:./?mode=memory",
		"memory-immediate": "file:./?mode=memory&_txlock=immediate",
	} {
		t.Run(testName, func(t *testing.T) {
			r, err := NewRegistry(uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			dbName := "db"
			r.DatabaseDrop(ctx, dbName)

			db, err := r.DatabaseGetOrCreate(ctx, dbName)
			require.NoError(t, err)
			require.NotNil(t, db)

			t.Cleanup(func() {
				r.DatabaseDrop(ctx, dbName)
			})

			collectionName := "collection"

			var i, createdTotal atomic.Int32

			teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
				id := i.Add(1)

				ready <- struct{}{}
				<-start

				created, err := r.CollectionCreate(ctx, dbName, collectionName)
				require.NoError(t, err)
				if created {
					createdTotal.Add(1)
				}

				created, err = r.CollectionCreate(ctx, dbName, collectionName)
				require.NoError(t, err)
				require.False(t, created)

				c := r.CollectionGet(ctx, dbName, collectionName)
				require.NotNil(t, c)
				require.Equal(t, collectionName, c.Name)

				list, err := r.CollectionList(ctx, dbName)
				require.NoError(t, err)
				require.Contains(t, list, collectionName)

				q := fmt.Sprintf("INSERT INTO %q (%s) VALUES(?)", c.TableName, DefaultColumn)
				doc := fmt.Sprintf(`{"$s": {"p": {"_id": {"t": "int"}}, "$k": ["_id"]}, "_id": %d}`, id)
				_, err = db.ExecContext(ctx, q, doc)
				require.NoError(t, err)
			})

			require.Equal(t, int32(1), createdTotal.Load())
		})
	}
}

func TestDropSameStress(t *testing.T) {
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	for testName, uri := range map[string]string{
		"file":             "file:./",
		"file-immediate":   "file:./?_txlock=immediate",
		"memory":           "file:./?mode=memory",
		"memory-immediate": "file:./?mode=memory&_txlock=immediate",
	} {
		t.Run(testName, func(t *testing.T) {
			r, err := NewRegistry(uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			dbName := "db"
			r.DatabaseDrop(ctx, dbName)

			db, err := r.DatabaseGetOrCreate(ctx, dbName)
			require.NoError(t, err)
			require.NotNil(t, db)

			t.Cleanup(func() {
				r.DatabaseDrop(ctx, dbName)
			})

			collectionName := "collection"

			created, err := r.CollectionCreate(ctx, dbName, collectionName)
			require.NoError(t, err)
			require.True(t, created)

			var droppedTotal atomic.Int32

			teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
				ready <- struct{}{}
				<-start

				dropped, err := r.CollectionDrop(ctx, dbName, collectionName)
				require.NoError(t, err)
				if dropped {
					droppedTotal.Add(1)
				}
			})

			require.Equal(t, int32(1), droppedTotal.Load())
		})
	}
}

func TestCreateDropSameStress(t *testing.T) {
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	for testName, uri := range map[string]string{
		"file":             "file:./",
		"file-immediate":   "file:./?_txlock=immediate",
		"memory":           "file:./?mode=memory",
		"memory-immediate": "file:./?mode=memory&_txlock=immediate",
	} {
		t.Run(testName, func(t *testing.T) {
			r, err := NewRegistry(uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			dbName := "db"
			r.DatabaseDrop(ctx, dbName)

			db, err := r.DatabaseGetOrCreate(ctx, dbName)
			require.NoError(t, err)
			require.NotNil(t, db)

			t.Cleanup(func() {
				r.DatabaseDrop(ctx, dbName)
			})

			collectionName := "collection"

			var i, createdTotal, droppedTotal atomic.Int32

			teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
				id := i.Add(1)

				ready <- struct{}{}
				<-start

				if id%2 == 0 {
					created, err := r.CollectionCreate(ctx, dbName, collectionName)
					require.NoError(t, err)
					if created {
						createdTotal.Add(1)
					}
				} else {
					dropped, err := r.CollectionDrop(ctx, dbName, collectionName)
					require.NoError(t, err)
					if dropped {
						droppedTotal.Add(1)
					}
				}
			})

			require.Less(t, int32(1), createdTotal.Load())
			require.Less(t, int32(1), droppedTotal.Load())
		})
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	r, err := NewRegistry("file:./?mode=memory", testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName := t.Name()

	// trying to get the version while database does not exist
	version, err := r.Version(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", version)

	// database exists, so version can be queried
	db, err := r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, db)

	version, err = r.Version(ctx)
	require.NoError(t, err)
	assert.Equal(t, "3.41.2", version)

	// no databases available, but the version should be returned because it's stored in the registry
	r.DatabaseDrop(ctx, dbName)

	version, err = r.Version(ctx)
	require.NoError(t, err)
	assert.Equal(t, "3.41.2", version)
}
