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

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/handlers/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

// testCollection creates, tests, and drops a collection in existing database.
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
	d := must.NotFail(types.NewDocument("_id", int32(42)))
	b := must.NotFail(sjson.Marshal(d))
	_, err = db.ExecContext(ctx, q, string(b))
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

	r, err := NewRegistry("file:./?mode=memory", testutil.Logger(t))
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName := t.Name()

	db, err := r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, db)

	t.Cleanup(func() {
		r.DatabaseDrop(ctx, dbName)
	})

	collectionName := t.Name()

	testCollection(t, ctx, r, db, dbName, collectionName)
}

func TestCreateDropSameStress(t *testing.T) {
	ctx := testutil.Ctx(t)

	for testName, uri := range map[string]string{
		"file":   "file:./",
		"memory": "file:./?mode=memory",
	} {
		t.Run(testName, func(t *testing.T) {
			r, err := NewRegistry(uri, testutil.Logger(t))
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

			teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
				ready <- struct{}{}
				<-start

				testCollection(t, ctx, r, db, dbName, collectionName)
			})
		})
	}
}

func TestCreateDropDifferentStress(t *testing.T) {
	ctx := testutil.Ctx(t)

	for testName, uri := range map[string]string{
		"file":   "file:./",
		"memory": "file:./?mode=memory",
	} {
		t.Run(testName, func(t *testing.T) {
			r, err := NewRegistry(uri, testutil.Logger(t))
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
