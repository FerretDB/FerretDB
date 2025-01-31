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

package documentdb

import (
	"context"
	"log/slog"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"

	"github.com/FerretDB/FerretDB/v2/build/version"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/state"
)

// newPgxPool create a new pgx pool.
// No actual connections are established immediately.
// State's version fields will be set only after a connection is established
// by some query or ping.
func newPgxPool(uri string, l *slog.Logger, sp *state.Provider) (*pgxpool.Pool, error) {
	must.NotBeZero(sp)

	u, err := url.Parse(uri)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	q := u.Query()
	newPgxPoolSetDefaults(q)
	u.RawQuery = q.Encode()

	config, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// versions and parameters could change without FerretDB restart
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// see https://github.com/jackc/pgx/issues/1726#issuecomment-1711612138
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		return newPgxPoolCheckConn(ctx, conn, l, sp)
	}

	// port tracing, tweak logging
	// TODO https://github.com/FerretDB/FerretDB/issues/3554

	// try to log everything; logger's configuration will skip extra levels if needed
	config.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   logging.NewPgxLogger(l),
		LogLevel: tracelog.LogLevelTrace,
	}

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement

	p, err := pgxpool.NewWithConfig(todoCtx, config)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return p, nil
}

// newPgxPoolSetDefaults sets default PostgreSQL URI parameters.
//
// Keep it in sync with docs.
func newPgxPoolSetDefaults(values url.Values) {
	// the default is too low
	if !values.Has("pool_max_conns") {
		values.Set("pool_max_conns", "50")
	}

	// to avoid the need to close unused pools ourselves
	if !values.Has("pool_max_conn_idle_time") {
		values.Set("pool_max_conn_idle_time", "1m")
	}

	// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNECT-APPLICATION-NAME
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

// newPgxPoolCheckConn checks established PostgreSQL connection and that settings are what we expect.
func newPgxPoolCheckConn(ctx context.Context, conn *pgx.Conn, l *slog.Logger, sp *state.Provider) error {
	must.NotBeZero(sp)

	var postgresqlVersion, documentdbVersion string

	row := conn.QueryRow(ctx, `SELECT version(), documentdb_api.binary_extended_version()`)
	if err := row.Scan(&postgresqlVersion, &documentdbVersion); err != nil {
		return lazyerrors.Error(err)
	}

	if s := sp.Get(); s.PostgreSQLVersion != postgresqlVersion || s.DocumentDBVersion != documentdbVersion {
		err := sp.Update(func(s *state.State) {
			s.PostgreSQLVersion = postgresqlVersion
			s.DocumentDBVersion = documentdbVersion
		})
		if err != nil {
			l.ErrorContext(ctx, "newPgxPoolCheckConn: failed to update state", logging.Error(err))
		}

		if s.DocumentDBVersion != "" && s.DocumentDBVersion != version.DocumentDB {
			l.WarnContext(
				ctx, "newPgxPoolCheckConn: unexpected DocumentDB version",
				slog.String("expected", version.DocumentDB), slog.String("actual", s.DocumentDBVersion),
			)
		}
	}

	if _, err := conn.Exec(ctx, "SET documentdb.enableUserCrud TO true"); err != nil {
		return lazyerrors.Error(err)
	}

	if _, err := conn.Exec(ctx, "SET documentdb.maxUserLimit TO 100"); err != nil {
		return lazyerrors.Error(err)
	}

	rows, err := conn.Query(ctx, "SHOW ALL")
	if err != nil {
		return lazyerrors.Error(err)
	}

	var name, setting, description string
	scans := []any{&name, &setting, &description}

	_, err = pgx.ForEachRow(rows, scans, func() error {
		switch name {
		case "client_encoding", "server_encoding", "lc_collate", "lc_ctype", "standard_conforming_strings",
			"documentdb.enableUserCrud", "documentdb.maxUserLimit":
			l.DebugContext(ctx, "newPgxPoolCheckConn", slog.String(name, setting))
		}

		return nil
	})
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
