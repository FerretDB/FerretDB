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

package pg

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

const (
	// Supported encoding.
	encUTF8 = "UTF8"

	// Supported locales: (For more info see: https://www.gnu.org/software/libc/manual/html_node/Standard-Locales.html)
	localeC     = "C"
	localePOSIX = "POSIX"
)

var (
	ErrNotExist     = fmt.Errorf("schema or table does not exist")
	ErrAlreadyExist = fmt.Errorf("schema or table already exist")
)

// Pool data struct for *pgxpool.Pool.
type Pool struct {
	*pgxpool.Pool
}

// TableStats describes some statistics for a table.
type TableStats struct {
	Table       string
	TableType   string
	SizeTotal   int32
	SizeIndexes int32
	SizeTable   int32
	Rows        int32
}

// DBStats describes some statistics for a database.
type DBStats struct {
	Name         string
	CountTables  int32
	CountRows    int32
	SizeTotal    int64
	SizeIndexes  int64
	SizeSchema   int64
	CountIndexes int32
}

// NewPool returns a pgxpool, a concurrency-safe connection pool for pgx.
func NewPool(connString string, logger *zap.Logger, lazy bool) (*Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("pg.NewPool: %w", err)
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

	if logger.Core().Enabled(zap.DebugLevel) {
		config.ConnConfig.LogLevel = pgx.LogLevelTrace
		config.ConnConfig.Logger = zapadapter.NewLogger(logger.Named("pgconn.Pool"))
	}

	ctx := context.Background()

	p, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("pg.NewPool: %w", err)
	}

	res := &Pool{
		Pool: p,
	}

	if !lazy {
		err = res.checkConnection(ctx)
	}

	return res, err
}

// Currently supported locale variants, compromised between https://www.postgresql.org/docs/9.3/multibyte.html
// and https://www.gnu.org/software/libc/manual/html_node/Locale-Names.html.
//
// Valid examples:
// * en_US.utf8,
// * en_US.utf-8
// * en_US.UTF8,
// * en_US.UTF-8.
func validUtf8Locale(setting string) bool {
	lowered := strings.ToLower(setting)

	return lowered == "en_us.utf8" || lowered == "en_us.utf-8"
}

func (p *Pool) checkConnection(ctx context.Context) error {
	logger := p.Config().ConnConfig.Logger

	rows, err := p.Query(ctx, "SHOW ALL")
	if err != nil {
		return fmt.Errorf("pg.Pool.checkConnection: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, setting, description string
		if err := rows.Scan(&name, &setting, &description); err != nil {
			return fmt.Errorf("pg.Pool.checkConnection: %w", err)
		}

		switch name {
		case "server_encoding":
			if setting != encUTF8 {
				return fmt.Errorf("pg.Pool.checkConnection: %q is %q, want %q", name, setting, encUTF8)
			}
		case "client_encoding":
			if setting != encUTF8 {
				return fmt.Errorf("pg.Pool.checkConnection: %q is %q, want %q", name, setting, encUTF8)
			}
		case "lc_collate":
			if setting != localeC && setting != localePOSIX && !validUtf8Locale(setting) {
				return fmt.Errorf("pg.Pool.checkConnection: %q is %q", name, setting)
			}
		case "lc_ctype":
			if setting != localeC && setting != localePOSIX && !validUtf8Locale(setting) {
				return fmt.Errorf("pg.Pool.checkConnection: %q is %q", name, setting)
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
		return fmt.Errorf("pg.Pool.checkConnection: %w", err)
	}

	return nil
}

// Schemas returns a sorted list of FerretDB database / PostgreSQL schema names.
func (pgPool *Pool) Schemas(ctx context.Context) ([]string, error) {
	sql := "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name"
	rows, err := pgPool.Query(ctx, sql)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	res := make([]string, 0, 2)
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if strings.HasPrefix(name, "pg_") || name == "information_schema" {
			continue
		}

		res = append(res, name)
	}
	if err = rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// Tables returns a sorted list of FerretDB collection / PostgreSQL table names.
func (pgPool *Pool) Tables(ctx context.Context, db string) ([]string, error) {
	sql := "SELECT table_name FROM information_schema.tables WHERE table_schema = $1 ORDER BY table_name"
	rows, err := pgPool.Query(ctx, sql, db)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	res := make([]string, 0, 2)
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, lazyerrors.Error(err)
		}

		res = append(res, name)
	}
	if err = rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// CreateSchema creates a new FerretDB database / PostgreSQL schema.
//
// It returns ErrAlreadyExist if schema already exist.
func (pgPool *Pool) CreateSchema(ctx context.Context, db string) error {
	sql := `CREATE SCHEMA ` + pgx.Identifier{db}.Sanitize()
	_, err := pgPool.Exec(ctx, sql)

	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.DuplicateSchema {
		return ErrAlreadyExist
	}

	return err
}

// DropSchema drops FerretDB database / PostgreSQL schema.
//
// It returns ErrNotExist if schema does not exist.
func (pgPool *Pool) DropSchema(ctx context.Context, db string) error {
	sql := `DROP SCHEMA ` + pgx.Identifier{db}.Sanitize() + ` CASCADE`
	_, err := pgPool.Exec(ctx, sql)

	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.InvalidSchemaName {
		return ErrNotExist
	}

	return err
}

// CreateTable creates a new FerretDB collection / PostgreSQL jsonb table.
//
// It returns ErrAlreadyExist if table already exist.
func (pgPool *Pool) CreateTable(ctx context.Context, db, collection string) error {
	sql := `CREATE TABLE ` + pgx.Identifier{db, collection}.Sanitize() + ` (_jsonb jsonb)`
	_, err := pgPool.Exec(ctx, sql)

	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.DuplicateTable {
		return ErrAlreadyExist
	}

	return err
}

// DropTable drops FerretDB collection / PostgreSQL table.
//
// It returns ErrNotExist is table does not exist.
func (pgPool *Pool) DropTable(ctx context.Context, db, collection string) error {
	// TODO probably not CASCADE
	sql := `DROP TABLE ` + pgx.Identifier{db, collection}.Sanitize() + `CASCADE`
	_, err := pgPool.Exec(ctx, sql)

	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UndefinedTable {
		return ErrNotExist
	}

	return err
}

// TableStats returns a set of statistics for a table.
func (pgPool *Pool) TableStats(ctx context.Context, db, table string) (*TableStats, error) {
	res := new(TableStats)
	sql := `
    SELECT table_name, table_type,
           pg_total_relation_size('"'||t.table_schema||'"."'||t.table_name||'"'),
           pg_indexes_size('"'||t.table_schema||'"."'||t.table_name||'"'),
           pg_relation_size('"'||t.table_schema||'"."'||t.table_name||'"'),
           COALESCE(s.n_live_tup, 0)
      FROM information_schema.tables AS t
      LEFT OUTER
      JOIN pg_stat_user_tables AS s ON s.schemaname = t.table_schema
                                      and s.relname = t.table_name
     WHERE t.table_schema = $1
       AND t.table_name = $2`

	err := pgPool.QueryRow(ctx, sql, db, table).
		Scan(&res.Table, &res.TableType, &res.SizeTotal, &res.SizeIndexes, &res.SizeTable, &res.Rows)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// DBStats returns a set of statistics for a database.
func (pgPool *Pool) DBStats(ctx context.Context, db string) (*DBStats, error) {
	res := new(DBStats)
	sql := `
    SELECT COUNT(distinct t.table_name)                                                             AS CountTables,
           COALESCE(SUM(s.n_live_tup), 0)                                                           AS CountRows,
           COALESCE(SUM(pg_total_relation_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0)  AS SizeTotal,
           COALESCE(SUM(pg_indexes_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0)         AS SizeIndexes,
           COALESCE(SUM(pg_relation_size('"'||t.table_schema||'"."'||t.table_name||'"')), 0)        AS SizeSchema,
           COUNT(distinct i.indexname)                                                              AS CountIndexes
      FROM information_schema.tables AS t
      LEFT OUTER
      JOIN pg_stat_user_tables       AS s ON s.schemaname = t.table_schema
                                         AND s.relname = t.table_name
      LEFT OUTER
      JOIN pg_indexes                AS i ON i.schemaname = t.table_schema
                                         AND i.tablename = t.table_name
     WHERE t.table_schema = $1`

	res.Name = db
	err := pgPool.QueryRow(ctx, sql, db).
		Scan(&res.CountTables, &res.CountRows, &res.SizeTotal, &res.SizeIndexes, &res.SizeSchema, &res.CountIndexes)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}
