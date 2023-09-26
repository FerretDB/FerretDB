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
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestInTransaction(t *testing.T) {
	// do not run parallel, there is something weird about context canceling and parallel test
	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	username, password := conninfo.Get(ctx).Auth()

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	u := "postgres://username:password@127.0.0.1:5432/ferretdb?pool_min_conns=1"

	pp, err := New(u, testutil.Logger(t), sp)
	require.NoError(t, err)

	p, err := pp.Get(username, password)
	require.NoError(t, err)

	_, err = p.Exec(ctx, "DROP TABLE IF EXISTS test_transaction; CREATE TABLE test_transaction(s TEXT);")
	require.NoError(t, err)

	t.Run("Commit", func(t *testing.T) {
		ctx = conninfo.Ctx(testutil.Ctx(t), conninfo.New())

		v := testutil.CollectionName(t)
		err = InTransaction(ctx, p, func(tx pgx.Tx) error {
			_, err = tx.Exec(ctx, "INSERT INTO test_transaction(s) VALUES ($1)", v)
			require.NoError(t, err)
			return nil
		})
		require.NoError(t, err)

		var res string
		err = p.QueryRow(ctx, "SELECT s FROM test_transaction WHERE s = $1", v).Scan(&res)
		require.NoError(t, err)
		require.Equal(t, v, res)
	})

	t.Run("Rollback", func(t *testing.T) {
		ctx = conninfo.Ctx(testutil.Ctx(t), conninfo.New())

		v := testutil.CollectionName(t)
		err = InTransaction(ctx, p, func(tx pgx.Tx) error {
			_, err = tx.Exec(ctx, "INSERT INTO test_transaction(s) VALUES ($1)", v)
			require.NoError(t, err)
			return errors.New("boom")
		})
		require.Error(t, err)

		var res string
		err = p.QueryRow(ctx, "SELECT s FROM test_transaction WHERE s = $1", v).Scan(&res)
		require.ErrorContains(t, err, "no rows in result set")
	})

	t.Run("ContextCancelRollback", func(t *testing.T) {
		ctx = conninfo.Ctx(testutil.Ctx(t), conninfo.New())

		var cancel func()
		ctx, cancel = context.WithCancel(ctx)

		v := testutil.CollectionName(t)
		err = InTransaction(ctx, p, func(tx pgx.Tx) error {
			_, err = tx.Exec(ctx, "INSERT INTO test_transaction(s) VALUES ($1)", v)
			require.NoError(t, err)

			cancel()

			return nil
		})
		require.Error(t, err)

		var res string
		err = p.QueryRow(context.WithoutCancel(ctx), "SELECT s FROM test_transaction WHERE s = $1", v).Scan(&res)
		require.ErrorContains(t, err, "no rows in result set")
	})

	t.Run("Panic", func(t *testing.T) {
		ctx = conninfo.Ctx(testutil.Ctx(t), conninfo.New())

		v := testutil.CollectionName(t)
		assert.Panics(t, func() {
			err = InTransaction(ctx, p, func(tx pgx.Tx) error {
				_, err = tx.Exec(ctx, "INSERT INTO test_transaction(s) VALUES ($1)", v)
				require.NoError(t, err)

				panic("boom")
			})
			require.ErrorContains(t, err, "no rows in result set")
		})

		var res string
		err = p.QueryRow(ctx, "SELECT s FROM test_transaction WHERE s = $1", v).Scan(&res)
		require.ErrorContains(t, err, "no rows in result set")
	})
}
