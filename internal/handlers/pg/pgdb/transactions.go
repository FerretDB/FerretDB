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
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"

	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// transactionConflictError is returned when one of the queries in the transaction returned an error because
// of an unexpected conflict. The caller could retry such a transaction.
type transactionConflictError struct {
	err error // underlying error
}

// newTransactionConflictError creates a new *transactionConflictError with the given underlying error.
func newTransactionConflictError(err error) error {
	return &transactionConflictError{err: err}
}

// Error implements the error interface.
func (e *transactionConflictError) Error() string {
	return "transactionConflictError: " + e.err.Error()
}

// InTransaction wraps the given function f in a transaction.
//
// If f returns an error, the transaction is rolled back.
//
// Passed context will be used for BEGIN/ROLLBACK/COMMIT statements.
// Context cancellation does not rollback the transaction.
// In practice, f should use the same ctx for Query/QueryRow/Exec that would return an error if context is canceled.
//
// Errors are wrapped with lazyerrors.Error,
// so the caller needs to use errors.Is to check the error,
// for example, errors.Is(err, ErrSchemaNotExist).
func (pgPool *Pool) InTransaction(ctx context.Context, f func(pgx.Tx) error) (err error) {
	var keepTx pgx.Tx

	err = pgPool.InTransactionKeep(ctx, func(tx pgx.Tx) error {
		keepTx = tx
		return f(tx)
	})
	if err != nil {
		err = lazyerrors.Error(err)
		return
	}

	err = keepTx.Commit(ctx)
	if err != nil {
		err = lazyerrors.Error(err)
		return
	}

	return
}

// InTransactionKeep is a variant of InTransaction that keeps transaction open if there is no error.
func (pgPool *Pool) InTransactionKeep(ctx context.Context, f func(pgx.Tx) error) (err error) {
	var tx pgx.Tx

	if tx, err = pgPool.p.Begin(ctx); err != nil {
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

		if e := tx.Rollback(ctx); e != nil {
			pgPool.logger.Log(
				ctx, tracelog.LogLevelError, "failed to perform rollback",
				map[string]any{"err": e},
			)
		}
	}()

	if err = f(tx); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	done = true
	return
}

// InTransactionRetry wraps the given function f in a transaction.
// If f returns (possibly wrapped) *transactionConflictError, the transaction is retried multiple times with delays.
// If the transaction still fails after that, the last error is returned.
func (pgPool *Pool) InTransactionRetry(ctx context.Context, f func(pgx.Tx) error) error {
	const (
		retriesMax    = 30
		retryDelayMax = 200 * time.Millisecond
	)

	var retry int64

	for {
		err := pgPool.InTransaction(ctx, f)
		var tcErr *transactionConflictError

		switch {
		case err == nil:
			return nil

		case errors.As(err, &tcErr):
			if retry >= retriesMax {
				return lazyerrors.Errorf("giving up after %d retries: %w", retry, err)
			}

			retry++
			pgPool.logger.Log(
				ctx, tracelog.LogLevelWarn, "attempt failed, retrying",
				map[string]any{"err": err, "retry": retry},
			)

			ctxutil.SleepWithJitter(ctx, retryDelayMax, retry)

		default:
			return lazyerrors.Errorf("non-retriable error: %w", err)
		}
	}
}
