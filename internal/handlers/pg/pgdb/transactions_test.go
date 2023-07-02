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

package pgdb

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestInTransaction(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t)

	assert.Panics(t, func() {
		pool.InTransaction(ctx, func(pgx.Tx) error {
			// We can't test `runtime.Goexit()`, but `panic(nil)` hangs the buggy code as well;
			// see comments inside `InTransaction`.
			//
			//nolint:vet // we need it for testing
			panic(nil)
		})
	})
}

func TestInTransactionKeep(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t)

	t.Run("Commit", func(t *testing.T) {
		t.Parallel()

		var keepTx pgx.Tx
		err := pool.InTransactionKeep(ctx, func(tx pgx.Tx) error {
			keepTx = tx
			return nil
		})
		require.NoError(t, err)

		var res int
		err = keepTx.QueryRow(ctx, "SELECT 1").Scan(&res)
		require.NoError(t, keepTx.Commit(ctx))
		require.NoError(t, err)
		assert.Equal(t, 1, res)
	})

	t.Run("Rollback", func(t *testing.T) {
		t.Parallel()

		var keepTx pgx.Tx
		err := pool.InTransactionKeep(ctx, func(tx pgx.Tx) error {
			keepTx = tx
			return errors.New("boom")
		})
		require.Error(t, err)

		var res int
		err = keepTx.QueryRow(ctx, "SELECT 1").Scan(&res)
		require.Equal(t, pgx.ErrTxClosed, keepTx.Commit(ctx))
		require.Equal(t, pgx.ErrTxClosed, keepTx.Rollback(ctx))
		require.Equal(t, pgx.ErrTxClosed, err)
		assert.Equal(t, 0, res)
	})
}
