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

	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/teststress"
)

func TestCreateDrop(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	// that also tests that query parameters are preserved by using non-writable directory
	p, _, err := New("file:/?mode=memory&_pragma=journal_mode(wal)", testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(p.Close)

	dbName := testutil.DatabaseName(t)

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

	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE %q (id INT) STRICT", t.Name()))
	require.NoError(t, err)

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

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	for testName, uri := range map[string]string{
		"file":             "file:./",
		"file-immediate":   "file:./?_txlock=immediate",
		"memory":           "file:./?mode=memory",
		"memory-immediate": "file:./?mode=memory&_txlock=immediate",
	} {
		t.Run(testName, func(t *testing.T) {
			p, _, err := New(uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(p.Close)

			var i atomic.Int32

			n := teststress.Stress(t, func(ready chan<- struct{}, start <-chan struct{}) {
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

				_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE TABLE %q (id INT) STRICT", t.Name()))
				require.NoError(t, err)

				dropped := p.Drop(ctx, dbName)
				require.True(t, dropped)

				dropped = p.Drop(ctx, dbName)
				require.False(t, dropped)

				db = p.GetExisting(ctx, dbName)
				require.Nil(t, db)
			})

			require.Equal(t, int32(n), i.Load())
		})
	}
}

func TestDefaults(t *testing.T) {
	t.Parallel()
	ctx := testutil.Ctx(t)

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	p, _, err := New("file:./", testutil.Logger(t), sp)
	require.NoError(t, err)

	dbName := testutil.DatabaseName(t)

	t.Cleanup(func() {
		p.Drop(ctx, dbName)
		p.Close()
	})

	db, _, err := p.GetOrCreate(ctx, dbName)
	require.NoError(t, err)

	rows, err := db.QueryContext(ctx, "PRAGMA compile_options")
	require.NoError(t, err)

	var options []string

	for rows.Next() {
		var o string
		require.NoError(t, rows.Scan(&o))
		t.Logf("option: %s", o)
		options = append(options, o)
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())

	require.Contains(t, options, "THREADSAFE=1")        // for it to work with database/sql
	require.Contains(t, options, "ENABLE_DBSTAT_VTAB")  // for dbStats/collStats/etc
	require.NotContains(t, options, "MAX_SCHEMA_RETRY") // see comments for SQLITE_SCHEMA in tests

	for q, expected := range map[string]string{
		"SELECT sqlite_version()":   "3.41.2",
		"SELECT sqlite_source_id()": "2023-03-22 11:56:21 0d1fc92f94cb6b76bffe3ec34d69cffde2924203304e8ffc4155597af0c191da",
		"PRAGMA busy_timeout":       "10000",
		"PRAGMA encoding":           "UTF-8",
		"PRAGMA journal_mode":       "wal",
	} {
		q, expected := q, expected
		t.Run(q, func(t *testing.T) {
			t.Parallel()

			var actual string
			err := db.QueryRowContext(ctx, q).Scan(&actual)
			require.NoError(t, err)
			require.Equal(t, expected, actual, q)
		})
	}
}
