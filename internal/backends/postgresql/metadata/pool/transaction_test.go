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
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestInTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())
	username, password := conninfo.Get(ctx).Auth()

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	u := "postgres://username:password@127.0.0.1:5432/ferretdb?pool_min_conns=1"

	pp, err := New(u, testutil.Logger(t), sp)
	require.NoError(t, err)

	t.Cleanup(pp.Close)

	p, err := pp.Get(username, password)
	require.NoError(t, err)

	t.Cleanup(p.Close)

	tableName := t.Name()
	_, err = p.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %[1]s; CREATE TABLE %[1]s(s TEXT);`, tableName))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err = p.Exec(ctx, fmt.Sprintf(`DROP TABLE %s`, tableName))
		require.NoError(t, err)
	})

	t.Run("Commit", func(t *testing.T) {
		t.Parallel()

		ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New()) // create new instance of ctx to avoid using canceled ctx
		err := err                                           // avoid data race

		v := testutil.CollectionName(t)
		err = InTransaction(ctx, p, func(tx pgx.Tx) error {
			_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s(s) VALUES ($1)`, tableName), v)
			require.NoError(t, err)
			return nil
		})
		require.NoError(t, err)

		var res string
		err = p.QueryRow(ctx, fmt.Sprintf(`SELECT s FROM %s WHERE s = $1`, tableName), v).Scan(&res)
		require.NoError(t, err)
		require.Equal(t, v, res)
	})

	t.Run("Rollback", func(t *testing.T) {
		t.Parallel()

		ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New()) // create new instance of ctx to avoid using canceled ctx
		err := err                                           // avoid data race

		v := testutil.CollectionName(t)
		err = InTransaction(ctx, p, func(tx pgx.Tx) error {
			_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s(s) VALUES ($1)`, tableName), v)
			require.NoError(t, err)
			return errors.New("boom")
		})
		require.Error(t, err)

		var res string
		err = p.QueryRow(ctx, fmt.Sprintf(`SELECT s FROM %s WHERE s = $1`, tableName), v).Scan(&res)
		require.Equal(t, pgx.ErrNoRows, err)
		require.Empty(t, res)
	})

	t.Run("ContextCancelRollback", func(t *testing.T) {
		t.Parallel()

		ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New()) // create new instance of ctx to avoid using canceled ctx
		err := err                                           // avoid data race

		var cancel func()
		ctx, cancel = context.WithCancel(ctx)

		v := testutil.CollectionName(t)
		err = InTransaction(ctx, p, func(tx pgx.Tx) error {
			_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s(s) VALUES ($1)`, tableName), v)
			require.NoError(t, err)

			cancel()

			return nil
		})
		require.Error(t, err)

		var res string
		err = p.QueryRow(context.WithoutCancel(ctx), fmt.Sprintf(`SELECT s FROM %s WHERE s = $1`, tableName), v).Scan(&res)
		require.Equal(t, pgx.ErrNoRows, err)
		require.Empty(t, res)
	})

	t.Run("Panic", func(t *testing.T) {
		t.Parallel()

		ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New()) // create new instance of ctx to avoid using canceled ctx
		err := err                                           // avoid data race

		v := testutil.CollectionName(t)
		assert.Panics(t, func() {
			err = InTransaction(ctx, p, func(tx pgx.Tx) error {
				_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s(s) VALUES ($1)`, tableName), v)
				require.NoError(t, err)

				//nolint:vet // need it for testing
				panic(nil)
			})
			require.Equal(t, pgx.ErrNoRows, err)
		})

		var res string
		err = p.QueryRow(ctx, fmt.Sprintf(`SELECT s FROM %s WHERE s = $1`, tableName), v).Scan(&res)
		require.Equal(t, pgx.ErrNoRows, err)
		require.Empty(t, res)
	})
}
