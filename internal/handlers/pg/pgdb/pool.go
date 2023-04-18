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
	"fmt"
	"net/url"
	"strings"

	zapadapter "github.com/jackc/pgx-zap"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/state"
)

const (
	// Supported encoding.
	encUTF8 = "UTF8"

	// Supported locales: (For more info see: https://www.gnu.org/software/libc/manual/html_node/Standard-Locales.html)
	localeC     = "C"
	localePOSIX = "POSIX"
)

// Pool represents PostgreSQL concurrency-safe connection pool.
type Pool struct {
	p      *pgxpool.Pool
	logger *zapadapter.Logger
}

// DBStats describes statistics for a database.
type DBStats struct {
	Name         string
	CountTables  int32
	CountRows    int32
	SizeTotal    int64
	SizeIndexes  int64
	SizeRelation int64
	CountIndexes int32
}

// NewPool returns a new concurrency-safe connection pool.
//
// Passed context is used only by the first checking connection.
// Canceling it after that function returns does nothing.
func NewPool(ctx context.Context, uri string, logger *zap.Logger, p *state.Provider) (*Pool, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("pgdb.NewPool: %w", err)
	}

	// pgx 'defaultMaxConns' is 4, which is not enough for us.
	// Set it to 20 by default if no query parameter is defined.
	// See: https://github.com/FerretDB/FerretDB/issues/1844
	values := u.Query()
	if !values.Has("pool_max_conns") {
		values.Set("pool_max_conns", "20")
	}

	u.RawQuery = values.Encode()

	config, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("pgdb.NewPool: %w", err)
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		var v string
		if err := conn.QueryRow(ctx, `SHOW server_version`).Scan(&v); err != nil {
			return err
		}

		if err := p.Update(func(s *state.State) { s.HandlerVersion = v }); err != nil {
			logger.Error("pgdb.Pool.AfterConnect: failed to update state", zap.Error(err))
		}

		return nil
	}

	// That only affects text protocol; pgx mostly uses a binary one.
	// See:
	// * https://github.com/jackc/pgx/issues/520
	// * https://github.com/jackc/pgx/issues/789
	// * https://github.com/jackc/pgx/issues/863
	// * https://github.com/FerretDB/FerretDB/issues/43
	config.ConnConfig.RuntimeParams["timezone"] = "UTC"

	config.ConnConfig.RuntimeParams["application_name"] = "FerretDB"
	config.ConnConfig.RuntimeParams["search_path"] = ""

	pgdbLogger := zapadapter.NewLogger(logger.Named("pgdb"))

	// try to log everything; logger's configuration will skip extra levels if needed
	config.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   pgdbLogger,
		LogLevel: tracelog.LogLevelTrace,
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("pgdb.NewPool: %w", err)
	}

	res := &Pool{
		p:      pool,
		logger: pgdbLogger,
	}

	if err = res.checkConnection(ctx); err != nil {
		res.Close()
		return nil, err
	}

	return res, nil
}

// Close closes all connections in the pool.
//
// It blocks until all connections are closed.
func (pgPool *Pool) Close() {
	pgPool.p.Close()
}

// isValidUTF8Locale Currently supported locale variants, compromised between https://www.postgresql.org/docs/9.3/multibyte.html
// and https://www.gnu.org/software/libc/manual/html_node/Locale-Names.html.
//
// Valid examples:
// * en_US.utf8,
// * en_US.utf-8
// * en_US.UTF8,
// * en_US.UTF-8.
func isValidUTF8Locale(setting string) bool {
	lowered := strings.ToLower(setting)

	return lowered == "en_us.utf8" || lowered == "en_us.utf-8"
}

// checkConnection checks PostgreSQL settings.
func (pgPool *Pool) checkConnection(ctx context.Context) error {
	rows, err := pgPool.p.Query(ctx, "SHOW ALL")
	if err != nil {
		return fmt.Errorf("pgdb.checkConnection: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		// handle variable number of columns as a workaround for https://github.com/cockroachdb/cockroach/issues/101715
		values, err := rows.Values()
		if err != nil {
			return fmt.Errorf("pgdb.checkConnection: %w", err)
		}

		if len(values) < 2 {
			return fmt.Errorf("pgdb.checkConnection: invalid row: %#v", values)
		}
		name := values[0].(string)
		setting := values[1].(string)

		switch name {
		case "server_encoding":
			if setting != encUTF8 {
				return fmt.Errorf("pgdb.checkConnection: %q is %q, want %q", name, setting, encUTF8)
			}
		case "client_encoding":
			if setting != encUTF8 {
				return fmt.Errorf("pgdb.checkConnection: %q is %q, want %q", name, setting, encUTF8)
			}
		case "lc_collate":
			if setting != localeC && setting != localePOSIX && !isValidUTF8Locale(setting) {
				return fmt.Errorf("pgdb.checkConnection: %q is %q", name, setting)
			}
		case "lc_ctype":
			if setting != localeC && setting != localePOSIX && !isValidUTF8Locale(setting) {
				return fmt.Errorf("pgdb.checkConnection: %q is %q", name, setting)
			}
		case "standard_conforming_strings": // To sanitize safely: https://github.com/jackc/pgx/issues/868#issuecomment-725544647
			if setting != "on" {
				return fmt.Errorf("pgdb.checkConnection: %q is %q, want %q", name, setting, "on")
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
		return fmt.Errorf("pgdb.checkConnection: %w", err)
	}

	return nil
}

// Stats returns a set of statistics for FerretDB server, database, collection
// - or, in terms of PostgreSQL, database, schema, table.
//
// It returns ErrTableNotExist is the given collection does not exist, and ignores other errors.
func (pgPool *Pool) Stats(ctx context.Context, db, collection string) (*DBStats, error) {
	res := &DBStats{
		Name: db,
	}

	err := pgPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		sql := `
	SELECT COUNT(distinct t.table_name)                                                         AS CountTables,
		COALESCE(SUM(s.n_live_tup), 0)                                                          AS CountRows,
		COALESCE(SUM(pg_total_relation_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0) AS SizeTotal,
		COALESCE(SUM(pg_indexes_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0)        AS SizeIndexes,
		COALESCE(SUM(pg_relation_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0)       AS SizeRelation,
		COUNT(distinct i.indexname)                                                             AS CountIndexes
	FROM information_schema.tables AS t
		LEFT OUTER JOIN pg_stat_user_tables AS s ON s.schemaname = t.table_schema AND s.relname = t.table_name
		LEFT OUTER JOIN pg_indexes          AS i ON i.schemaname = t.table_schema AND i.tablename = t.table_name`

		// TODO Exclude service schemas from the query above https://github.com/FerretDB/FerretDB/issues/1068

		var args []any

		if db != "" {
			sql += " WHERE t.table_schema = $1"
			args = append(args, db)

			if collection != "" {
				metadata, err := newMetadataStorage(tx, db, collection).get(ctx, false)
				if err != nil {
					return err
				}

				sql += " AND t.table_name = $2"
				args = append(args, metadata.table)
			}
		}

		row := tx.QueryRow(ctx, sql, args...)

		return row.Scan(&res.CountTables, &res.CountRows, &res.SizeTotal, &res.SizeIndexes, &res.SizeRelation, &res.CountIndexes)
	})

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, ErrTableNotExist):
		// return this error as is because it can be handled by the caller
		return nil, err
	default:
		// just log it for now
		// TODO https://github.com/FerretDB/FerretDB/issues/1346
		pgPool.logger.Log(ctx, tracelog.LogLevelError, "pgdb.Stats: failed to get stats", map[string]any{"err": err})
	}

	return res, nil
}
