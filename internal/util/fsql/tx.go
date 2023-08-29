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

package fsql

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/observability"
	"github.com/FerretDB/FerretDB/internal/util/resource"
)

// Tz wraps [*database/sql.Tx] with resource tracking.
//
// It exposes the subset of *sql.Tx methods we use.
type Tx struct {
	sqlTx *sql.Tx
	l     *zap.Logger
	token *resource.Token
}

// wrapTx creates new Tx.
func wrapTx(tx *sql.Tx, l *zap.Logger) *Tx {
	if tx == nil {
		return nil
	}

	res := &Tx{
		sqlTx: tx,
		l:     l,
		token: resource.NewToken(),
	}

	resource.Track(res, res.token)

	return res
}

// Commit calls [*sql.Tx.Commit].
func (tx *Tx) Commit() error {
	resource.Untrack(tx, tx.token)
	return tx.sqlTx.Commit()
}

// Rollback calls [*sql.Tx.Rollback].
func (tx *Tx) Rollback() error {
	resource.Untrack(tx, tx.token)
	return tx.sqlTx.Rollback()
}

// QueryContext calls [*sql.Tx.QueryContext].
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...any) (*Rows, error) {
	defer observability.FuncCall(ctx)()

	start := time.Now()

	fields := []any{zap.Any("args", args)}
	tx.l.Sugar().With(fields...).Debugf(">>> %s", query)

	rows, err := tx.sqlTx.QueryContext(ctx, query, args...)

	fields = append(fields, zap.Duration("time", time.Since(start)), zap.Error(err))
	tx.l.Sugar().With(fields...).Debugf("<<< %s", query)

	return wrapRows(rows), err
}

// QueryRowContext calls [*sql.Tx.QueryRowContext].
func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	defer observability.FuncCall(ctx)()

	start := time.Now()

	fields := []any{zap.Any("args", args)}
	tx.l.Sugar().With(fields...).Debugf(">>> %s", query)

	row := tx.sqlTx.QueryRowContext(ctx, query, args...)

	fields = append(fields, zap.Duration("time", time.Since(start)), zap.Error(row.Err()))
	tx.l.Sugar().With(fields...).Debugf("<<< %s", query)

	return row
}

// ExecContext calls [*sql.Tx.ExecContext].
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	defer observability.FuncCall(ctx)()

	start := time.Now()

	fields := []any{zap.Any("args", args)}
	tx.l.Sugar().With(fields...).Debugf(">>> %s", query)

	res, err := tx.sqlTx.ExecContext(ctx, query, args...)

	// to differentiate between 0 and nil
	var ra *int64

	if res != nil {
		rav, _ := res.RowsAffected()
		ra = &rav
	}

	fields = append(fields, zap.Int64p("rows", ra), zap.Duration("time", time.Since(start)), zap.Error(err))
	tx.l.Sugar().With(fields...).Debugf("<<< %s", query)

	return res, err
}
