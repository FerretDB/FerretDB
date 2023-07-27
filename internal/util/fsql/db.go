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

	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// sqlDB is a subset of [*database/sql.DB] methods that we use.
//
// It mainly exist to check interfaces.
type sqlDB interface {
	Close() error
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

// DB wraps [*database/sql.DB] with tracing, metrics, logging, and resource tracking.
type DB struct {
	*metricsCollector

	sqlDB sqlDB
	l     *zap.Logger
	token *resource.Token
}

// WrapDB creates a new DB.
//
// Name is used for metric label values, etc.
// Logger (that will be named) is used for query logging.
func WrapDB(db *sql.DB, name string, l *zap.Logger) *DB {
	res := &DB{
		metricsCollector: newMetricsCollector(name, db.Stats),
		sqlDB:            db,
		l:                l.Named(name),
		token:            resource.NewToken(),
	}

	resource.Track(res, res.token)

	return res
}

// Close implements sqlDB.
func (db *DB) Close() error {
	resource.Untrack(db, db.token)
	return db.sqlDB.Close()
}

// QueryContext implements sqlDB.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	defer observability.FuncCall(ctx)()

	start := time.Now()

	fields := []any{zap.Any("args", args)}
	db.l.Sugar().With(fields...).Debugf(">>> %s", query)

	rows, err := db.sqlDB.QueryContext(ctx, query, args...)

	fields = append(fields, zap.Duration("duration", time.Since(start)), zap.Error(err))
	db.l.Sugar().With(fields...).Debugf("<<< %s", query)

	return rows, err
}

// QueryRowContext implements sqlDB.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	defer observability.FuncCall(ctx)()

	start := time.Now()

	fields := []any{zap.Any("args", args)}
	db.l.Sugar().With(fields...).Debugf(">>> %s", query)

	row := db.sqlDB.QueryRowContext(ctx, query, args...)

	fields = append(fields, zap.Duration("duration", time.Since(start)), zap.Error(row.Err()))
	db.l.Sugar().With(fields...).Debugf("<<< %s", query)

	return row
}

// ExecContext implements sqlDB.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	defer observability.FuncCall(ctx)()

	start := time.Now()

	fields := []any{zap.Any("args", args)}
	db.l.Sugar().With(fields...).Debugf(">>> %s", query)

	res, err := db.sqlDB.ExecContext(ctx, query, args...)

	fields = append(fields, zap.Duration("duration", time.Since(start)), zap.Error(err))
	db.l.Sugar().With(fields...).Debugf("<<< %s", query)

	return res, err
}

// check interfaces
var (
	_ sqlDB                = (*sql.DB)(nil)
	_ sqlDB                = (*DB)(nil)
	_ prometheus.Collector = (*DB)(nil)
)
