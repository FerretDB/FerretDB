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
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
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
	*pgxpool.Pool
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
//
// If lazy is true, then connectivity is not checked.
// Lazy connections are used by FerretDB when it starts earlier than backend.
// Non-lazy connections are used by tests.
func NewPool(ctx context.Context, uri string, logger *zap.Logger, lazy bool, p *state.Provider) (*Pool, error) {
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, fmt.Errorf("pgdb.NewPool: %w", err)
	}

	config.LazyConnect = lazy

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

	// try to log everything; logger's configuration will skip extra levels if needed
	config.ConnConfig.LogLevel = pgx.LogLevelTrace
	config.ConnConfig.Logger = zapadapter.NewLogger(logger.Named("pgdb"))

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("pgdb.NewPool: %w", err)
	}

	res := &Pool{
		Pool: pool,
	}

	if !lazy {
		if err = res.checkConnection(ctx); err != nil {
			res.Close()
			return nil, err
		}
	}

	return res, nil
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
	logger := pgPool.Config().ConnConfig.Logger

	rows, err := pgPool.Query(ctx, "SHOW ALL")
	if err != nil {
		return fmt.Errorf("pgdb.checkConnection: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, setting, description string
		if err := rows.Scan(&name, &setting, &description); err != nil {
			return fmt.Errorf("pgdb.checkConnection: %w", err)
		}

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
		default:
			continue
		}

		if logger != nil {
			logger.Log(ctx, pgx.LogLevelDebug, "PostgreSQL setting", map[string]any{
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

// SchemaStats is a stub to return a set of statistics for FerretDB server, database, collection
// - or, in terms of PostgreSQL, database, schema, table.
// Currently, it always returns an empty DBStats, this needs to be fixed later.
// TODO https://github.com/FerretDB/FerretDB/issues/1346
func (pgPool *Pool) SchemaStats(ctx context.Context, schema, collection string) (*DBStats, error) {
	return &DBStats{
		Name: schema,
	}, nil
}
