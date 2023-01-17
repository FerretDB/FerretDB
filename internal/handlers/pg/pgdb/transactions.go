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
	"math/rand"
	"time"

	"github.com/jackc/pgx/v4"

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
// If f returns an error, the transaction is rolled back.
// Errors are wrapped with lazyerrors.Error,
// so the caller needs to use errors.Is to check the error,
// for example, errors.Is(err, ErrSchemaNotExist).
func (pgPool *Pool) InTransaction(ctx context.Context, f func(pgx.Tx) error) (err error) {
	var tx pgx.Tx

	if tx, err = pgPool.Begin(ctx); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	defer func() {
		if err == nil {
			return
		}

		if rerr := tx.Rollback(ctx); rerr != nil {
			pgPool.Config().ConnConfig.Logger.Log(
				ctx, pgx.LogLevelError, "failed to perform rollback",
				map[string]any{"error": rerr},
			)
		}
	}()

	if err = f(tx); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	if err = tx.Commit(ctx); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	return
}

// InTransactionRetry wraps the given function f in a transaction.
// If f returns (possibly wrapped) *transactionConflictError, the transaction is retried multiple times with delays.
// If the transaction still fails after that, the last error is returned.
func (pgPool *Pool) InTransactionRetry(ctx context.Context, f func(pgx.Tx) error) error {
	// TODO use exponential backoff with jitter instead
	// https://github.com/FerretDB/FerretDB/issues/1720
	const (
		retriesMax    = 20
		retryDelayMin = 100 * time.Millisecond
		retryDelayMax = 200 * time.Millisecond
	)

	var err error
	var tcErr *transactionConflictError

	for retry := 0; retry < retriesMax; retry++ {
		err = pgPool.InTransaction(ctx, f)

		switch {
		case err == nil:
			return nil
		case errors.As(err, &tcErr):
			deltaMS := rand.Int63n((retryDelayMax - retryDelayMin).Milliseconds())
			ctxutil.Sleep(ctx, retryDelayMin+time.Duration(deltaMS)*time.Millisecond)
		default:
			return lazyerrors.Error(err)
		}
	}

	return lazyerrors.Error(err)
}
