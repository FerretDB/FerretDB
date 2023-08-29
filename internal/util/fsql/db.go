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

// Package fsql provides [database/sql] utilities.
package fsql

import (
	"context"
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// DB wraps [*database/sql.DB] with tracing, metrics, logging, and resource tracking.
//
// It exposes the subset of *sql.DB methods we use except that it returns *Rows instead of *sql.Rows.
// It also exposes additional methods.
type DB struct {
	*metricsCollector

	sqlDB *sql.DB
	l     *zap.Logger
	token *resource.Token
}

// WrapDB creates a new DB.
//
// Name is used for metric label values, etc.
// Logger (that will be named) is used for query logging.
func WrapDB(db *sql.DB, name string, l *zap.Logger) *DB {
	if db == nil {
		return nil
	}

	res := &DB{
		metricsCollector: newMetricsCollector(name, db.Stats),
		sqlDB:            db,
		l:                l.Named(name),
		token:            resource.NewToken(),
	}

	resource.Track(res, res.token)

	return res
}

// Close calls [*sql.DB.Close].
func (db *DB) Close() error {
	resource.Untrack(db, db.token)
	return db.sqlDB.Close()
}

// QueryContext calls [*sql.DB.QueryContext].
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*Rows, error) {
	defer observability.FuncCall(ctx)()

	start := time.Now()

	fields := []any{zap.Any("args", args)}
	db.l.Sugar().With(fields...).Debugf(">>> %s", query)

	rows, err := db.sqlDB.QueryContext(ctx, query, args...)

	fields = append(fields, zap.Duration("time", time.Since(start)), zap.Error(err))
	db.l.Sugar().With(fields...).Debugf("<<< %s", query)

	return wrapRows(rows), err
}

// QueryRowContext calls [*sql.DB.QueryRowContext].
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	defer observability.FuncCall(ctx)()

	start := time.Now()

	fields := []any{zap.Any("args", args)}
	db.l.Sugar().With(fields...).Debugf(">>> %s", query)

	row := db.sqlDB.QueryRowContext(ctx, query, args...)

	fields = append(fields, zap.Duration("time", time.Since(start)), zap.Error(row.Err()))
	db.l.Sugar().With(fields...).Debugf("<<< %s", query)

	return row
}

// ExecContext calls [*sql.DB.ExecContext].
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	defer observability.FuncCall(ctx)()

	start := time.Now()

	fields := []any{zap.Any("args", args)}
	db.l.Sugar().With(fields...).Debugf(">>> %s", query)

	res, err := db.sqlDB.ExecContext(ctx, query, args...)

	// to differentiate between 0 and nil
	var ra *int64

	if res != nil {
		rav, _ := res.RowsAffected()
		ra = &rav
	}

	fields = append(fields, zap.Int64p("rows", ra), zap.Duration("time", time.Since(start)), zap.Error(err))
	db.l.Sugar().With(fields...).Debugf("<<< %s", query)

	return res, err
}

// InTransaction wraps the given function f in a transaction.
//
// If f returns an error or context is canceled, the transaction is rolled back.
func (db *DB) InTransaction(ctx context.Context, f func(*Tx) error) (err error) {
	defer observability.FuncCall(ctx)()

	var sqlTx *sql.Tx

	if sqlTx, err = db.sqlDB.BeginTx(ctx, nil); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	tx := wrapTx(sqlTx, db.l)

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

		_ = tx.Rollback()
	}()

	if err = f(tx); err != nil {
		// do not wrap f's error because the caller depends on it in some cases
		return
	}

	if err = tx.Commit(); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	done = true

	return
}

// check interfaces
var (
	_ prometheus.Collector = (*DB)(nil)
)
