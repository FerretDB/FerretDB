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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/observability"
)

// InTransaction uses pool p and wraps the given function f in a transaction.
//
// If f returns an error or context is canceled, the transaction is rolled back.
func InTransaction(ctx context.Context, p *pgxpool.Pool, f func(tx *pgx.Tx) error) (err error) {
	defer observability.FuncCall(ctx)()

	var pgTx pgx.Tx

	if pgTx, err = p.BeginTx(ctx, pgx.TxOptions{}); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	var done bool

	defer func() {
		// It is not enough to check `err == nil` there,
		// because in tests `f` could contain testify/require.XXX or `testing.TB.FailNow()` calls
		// that call `runtime.Goexit()`, leaving `err` unset in `err = f(tx)` below.
		// This situation would hang a test.
		//
		// As a bonus, checking a separate variable also handles any panics in `f`,
		// including `panic(nil)` that is problematic for tests too.
		if done {
			return
		}

		if err == nil {
			err = lazyerrors.Errorf("transaction was not committed")
		}

		_ = pgTx.Rollback(ctx)
	}()

	if err = f(&pgTx); err != nil {
		// do not wrap f's error because the caller depends on it in some cases
		return
	}

	if err = pgTx.Commit(ctx); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	done = true

	return
}
