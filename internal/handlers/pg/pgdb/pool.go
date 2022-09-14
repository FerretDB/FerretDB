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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
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
func NewPool(ctx context.Context, connString string, logger *zap.Logger, lazy bool) (*Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("pgdb.NewPool: %w", err)
	}

	config.LazyConnect = lazy

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

	p, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("pgdb.NewPool: %w", err)
	}

	res := &Pool{
		Pool: p,
	}

	if !lazy {
		err = res.checkConnection(ctx)
	}

	return res, err
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

// SchemaStats returns a set of statistics for FerretDB server, database, collection - or, in terms of PostgreSQL,
// database, schema, table.
// If schema is empty, it calculates statistics across the whole PostgreSQL database.
// If collection is empty it calculates statistics across the whole PostgreSQL schema.
func (pgPool *Pool) SchemaStats(ctx context.Context, schema, collection string) (*DBStats, error) {
	var res DBStats

	sql := `
    SELECT COUNT(distinct t.table_name)                                                             AS CountTables,
           COALESCE(SUM(s.n_live_tup), 0)                                                           AS CountRows,
           COALESCE(SUM(pg_total_relation_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0)  AS SizeTotal,
           COALESCE(SUM(pg_indexes_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0)         AS SizeIndexes,
           COALESCE(SUM(pg_relation_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0)        AS SizeRelation,
           COUNT(distinct i.indexname)                                                              AS CountIndexes
      FROM information_schema.tables AS t
      LEFT OUTER
      JOIN pg_stat_user_tables       AS s ON s.schemaname = t.table_schema
                                         AND s.relname = t.table_name
      LEFT OUTER
      JOIN pg_indexes                AS i ON i.schemaname = t.table_schema
                                         AND i.tablename = t.table_name`

	// TODO Exclude service schemas from the query above https://github.com/FerretDB/FerretDB/issues/1068

	args := []any{}

	if schema != "" {
		sql += " WHERE t.table_schema = $1"
		args = append(args, schema)

		if collection != "" {
			sql += " AND t.table_name = $2"
			args = append(args, collection)
		}
	}

	res.Name = schema
	err := pgPool.QueryRow(ctx, sql, args...).
		Scan(&res.CountTables, &res.CountRows, &res.SizeTotal, &res.SizeIndexes, &res.SizeRelation, &res.CountIndexes)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return &res, nil
}

// SetDocumentByID sets a document by its ID.
//
// Deprecated: use function instead.
func (pgPool *Pool) SetDocumentByID(ctx context.Context, sp *SQLParam, id any, doc *types.Document) (int64, error) {
	var n int64
	err := pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		var err error
		n, err = SetDocumentByID(ctx, tx, sp, id, doc)
		return err
	})

	return n, err
}

// DeleteDocumentsByID deletes documents by given IDs.
//
// Deprecated: use function instead.
func (pgPool *Pool) DeleteDocumentsByID(ctx context.Context, sp *SQLParam, ids []any) (int64, error) {
	var n int64
	err := pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		var err error
		n, err = DeleteDocumentsByID(ctx, tx, sp, ids)
		return err
	})

	return n, err
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
