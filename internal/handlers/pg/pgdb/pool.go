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
	"net/url"
	"strings"

	zapadapter "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/debugbuild"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/state"
)

var (
	// The only supported encoding in canonical form.
	supportedEncoding = "UTF8"

	// Supported locales in canonical forms.
	supportedLocales = []string{"POSIX", "C", "C.UTF8", "en_US.UTF8"}
)

// Pool represents PostgreSQL concurrency-safe connection pool.
type Pool struct {
	p      *pgxpool.Pool
	logger *zapadapter.Logger
}

// NewPool returns a new concurrency-safe connection pool.
//
// Passed context is used only by the first checking connection.
// Canceling it after that function returns does nothing.
func NewPool(ctx context.Context, uri string, logger *zap.Logger, p *state.Provider) (*Pool, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	values := u.Query()
	setDefaultValues(values)
	u.RawQuery = values.Encode()

	config, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		var v string
		if err := conn.QueryRow(ctx, `SHOW server_version`).Scan(&v); err != nil {
			return lazyerrors.Error(err)
		}

		if err := p.Update(func(s *state.State) { s.HandlerVersion = v }); err != nil {
			logger.Error("pgdb.Pool.AfterConnect: failed to update state", zap.Error(err))
		}

		return nil
	}

	pgdbLogger := zapadapter.NewLogger(logger.Named("pgdb"))

	tracers := []pgx.QueryTracer{
		// try to log everything; logger's configuration will skip extra levels if needed
		&tracelog.TraceLog{
			Logger:   pgdbLogger,
			LogLevel: tracelog.LogLevelTrace,
		},
	}

	if debugbuild.Enabled {
		tracers = append(tracers, new(debugTracer))
	}

	config.ConnConfig.Tracer = &multiQueryTracer{
		Tracers: tracers,
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := &Pool{
		p:      pool,
		logger: pgdbLogger,
	}

	if err = res.checkConnection(ctx); err != nil {
		res.Close()
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// Close closes all connections in the pool.
//
// It blocks until all connections are closed.
func (pgPool *Pool) Close() {
	pgPool.p.Close()
}

// setDefaultValue sets default query parameters.
//
// Keep it in sync with docs.
func setDefaultValues(values url.Values) {
	if !values.Has("pool_max_conns") {
		// the default is too low
		values.Set("pool_max_conns", "50")
	}

	values.Set("application_name", "FerretDB")

	// That only affects text protocol; pgx mostly uses a binary one.
	// See:
	//   - https://github.com/jackc/pgx/issues/520
	//   - https://github.com/jackc/pgx/issues/789
	//   - https://github.com/jackc/pgx/issues/863
	//
	// TODO https://github.com/FerretDB/FerretDB/issues/43
	values.Set("timezone", "UTC")
}

// simplifySetting simplifies PostgreSQL setting value for comparison.
func simplifySetting(v string) string {
	return strings.ToLower(strings.ReplaceAll(v, "-", ""))
}

// isSupportedEncoding checks `server_encoding` and `client_encoding` values.
func isSupportedEncoding(v string) bool {
	return simplifySetting(v) == simplifySetting(supportedEncoding)
}

// isSupportedLocale checks `lc_collate` and `lc_ctype` values.
func isSupportedLocale(v string) bool {
	v = simplifySetting(v)

	for _, s := range supportedLocales {
		if v == simplifySetting(s) {
			return true
		}
	}

	return false
}

// checkConnection checks PostgreSQL settings.
func (pgPool *Pool) checkConnection(ctx context.Context) error {
	rows, err := pgPool.p.Query(ctx, "SHOW ALL")
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		// handle variable number of columns as a workaround for https://github.com/cockroachdb/cockroach/issues/101715
		values, err := rows.Values()
		if err != nil {
			return lazyerrors.Error(err)
		}

		if len(values) < 2 {
			return lazyerrors.Errorf("invalid row: %#v", values)
		}
		name := values[0].(string)
		setting := values[1].(string)

		switch name {
		case "server_encoding", "client_encoding":
			if !isSupportedEncoding(setting) {
				return lazyerrors.Errorf("%q is %q; supported value is %q", name, setting, supportedEncoding)
			}
		case "lc_collate", "lc_ctype":
			if !isSupportedLocale(setting) {
				return lazyerrors.Errorf("%q is %q; supported values are %v", name, setting, supportedLocales)
			}
		case "standard_conforming_strings": // To sanitize safely: https://github.com/jackc/pgx/issues/868#issuecomment-725544647
			if setting != "on" {
				return lazyerrors.Errorf("%q is %q, want %q", name, setting, "on")
			}
		default:
			continue
		}

		if pgPool.logger != nil {
			pgPool.logger.Log(ctx, tracelog.LogLevelDebug, "PostgreSQL setting", map[string]any{
				"name":    name,
				"setting": setting,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
