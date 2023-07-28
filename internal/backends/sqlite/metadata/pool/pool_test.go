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

package pool

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

func TestCreateDrop(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	// that also tests that query parameters are preserved by using non-writable directory
	p, err := New("file:./?mode=memory&_pragma=journal_mode(wal)", testutil.Logger(t))
	require.NoError(t, err)
	t.Cleanup(p.Close)

	dbName := t.Name()

	db := p.GetExisting(ctx, dbName)
	require.Nil(t, db)

	db, created, err := p.GetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, db)
	require.True(t, created)

	db2, created, err := p.GetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.Same(t, db, db2)
	require.False(t, created)

	db2 = p.GetExisting(ctx, dbName)
	require.Same(t, db, db2)

	require.Contains(t, p.List(ctx), dbName)

	// journal_mode is silently ignored for mode=memory
	var res string
	err = db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&res)
	require.NoError(t, err)
	require.Equal(t, "memory", res)

	dropped := p.Drop(ctx, dbName)
	require.True(t, dropped)

	dropped = p.Drop(ctx, dbName)
	require.False(t, dropped)

	db = p.GetExisting(ctx, dbName)
	require.Nil(t, db)
}

func TestCreateDropStress(t *testing.T) {
	ctx := testutil.Ctx(t)

	for testName, uri := range map[string]string{
		"file":   "file:./",
		"memory": "file:./?mode=memory",
	} {
		t.Run(testName, func(t *testing.T) {
			p, err := New(uri, testutil.Logger(t))
			require.NoError(t, err)
			t.Cleanup(p.Close)

			var i atomic.Int32

			teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
				dbName := fmt.Sprintf("db_%03d", i.Add(1))

				t.Cleanup(func() {
					p.Drop(ctx, dbName)
				})

				ready <- struct{}{}
				<-start

				db := p.GetExisting(ctx, dbName)
				require.Nil(t, db)

				db, created, err := p.GetOrCreate(ctx, dbName)
				require.NoError(t, err)
				require.NotNil(t, db)
				require.True(t, created)

				db2, created, err := p.GetOrCreate(ctx, dbName)
				require.NoError(t, err)
				require.Same(t, db, db2)
				require.False(t, created)

				db2 = p.GetExisting(ctx, dbName)
				require.Same(t, db, db2)

				require.Contains(t, p.List(ctx), dbName)

				err = db.QueryRowContext(ctx, "SELECT 1").Err()
				require.NoError(t, err)

				dropped := p.Drop(ctx, dbName)
				require.True(t, dropped)

				dropped = p.Drop(ctx, dbName)
				require.False(t, dropped)

				db = p.GetExisting(ctx, dbName)
				require.Nil(t, db)
			})
		})
	}
}

func TestPragmas(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	p, err := New("file:./", testutil.Logger(t))
	require.NoError(t, err)
	t.Cleanup(func() {
		p.Drop(ctx, t.Name())
		p.Close()
	})

	db, _, err := p.GetOrCreate(ctx, t.Name())
	require.NoError(t, err)

	for pragma, expected := range map[string]string{
		"busy_timeout": "5000",
		"journal_mode": "wal",
	} {
		pragma, expected := pragma, expected
		t.Run(pragma, func(t *testing.T) {
			t.Parallel()

			var actual string
			err = db.QueryRowContext(ctx, "PRAGMA "+pragma).Scan(&actual)
			require.NoError(t, err)
			require.Equal(t, expected, actual, pragma)
		})
	}
}
