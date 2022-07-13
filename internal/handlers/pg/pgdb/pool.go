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

// Package pgdb provides PostgreSQL connection utilities.
package pgdb

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// Supported encoding.
	encUTF8 = "UTF8"

	// Supported locales: (For more info see: https://www.gnu.org/software/libc/manual/html_node/Standard-Locales.html)
	localeC     = "C"
	localePOSIX = "POSIX"
)

// Regex validateCollectionNameRe validates collection names.
var validateCollectionNameRe = regexp.MustCompile("^[a-zA-Z_][a-zA-Z0-9_]{0,119}$")

// Errors are wrapped with lazyerrors.Error,
// so the caller needs to use errors.Is to check the error,
// for example, errors.Is(err, ErrSchemaNotExist).
var (
	// ErrTableNotExist indicates that there is no such table.
	ErrTableNotExist = fmt.Errorf("table does not exist")

	// ErrSchemaNotExist indicates that there is no such schema.
	ErrSchemaNotExist = fmt.Errorf("schema does not exist")

	// ErrAlreadyExist indicates that a schema or table already exists.
	ErrAlreadyExist = fmt.Errorf("schema or table already exist")

	// ErrInvalidTableName indicates that a schema or table didn't passed name checks.
	ErrInvalidTableName = fmt.Errorf("invalid table name")
)

// Pool represents PostgreSQL concurrency-safe connection pool.
type Pool struct {
	*pgxpool.Pool
	logger *zap.Logger // TODO remove, use Pool.Config().ConnConfig.Logger instead
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
	config.ConnConfig.Logger = zapadapter.NewLogger(logger.Named("pgdb.Pool"))

	p, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("pgdb.NewPool: %w", err)
	}

	res := &Pool{
		Pool:   p,
		logger: logger.Named("pgdb.Pool"),
	}

	if !lazy {
		err = res.checkConnection(ctx)
	}

	return res, err
}

// IsValidUTF8Locale Currently supported locale variants, compromised between https://www.postgresql.org/docs/9.3/multibyte.html
// and https://www.gnu.org/software/libc/manual/html_node/Locale-Names.html.
//
// Valid examples:
// * en_US.utf8,
// * en_US.utf-8
// * en_US.UTF8,
// * en_US.UTF-8.
func IsValidUTF8Locale(setting string) bool {
	lowered := strings.ToLower(setting)

	return lowered == "en_us.utf8" || lowered == "en_us.utf-8"
}

// checkConnection checks PostgreSQL settings.
func (pgPool *Pool) checkConnection(ctx context.Context) error {
	logger := pgPool.Config().ConnConfig.Logger

	rows, err := pgPool.Query(ctx, "SHOW ALL")
	if err != nil {
		return fmt.Errorf("pgdb.Pool.checkConnection: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, setting, description string
		if err := rows.Scan(&name, &setting, &description); err != nil {
			return fmt.Errorf("pgdb.Pool.checkConnection: %w", err)
		}

		switch name {
		case "server_encoding":
			if setting != encUTF8 {
				return fmt.Errorf("pgdb.Pool.checkConnection: %q is %q, want %q", name, setting, encUTF8)
			}
		case "client_encoding":
			if setting != encUTF8 {
				return fmt.Errorf("pgdb.Pool.checkConnection: %q is %q, want %q", name, setting, encUTF8)
			}
		case "lc_collate":
			if setting != localeC && setting != localePOSIX && !IsValidUTF8Locale(setting) {
				return fmt.Errorf("pgdb.Pool.checkConnection: %q is %q", name, setting)
			}
		case "lc_ctype":
			if setting != localeC && setting != localePOSIX && !IsValidUTF8Locale(setting) {
				return fmt.Errorf("pgdb.Pool.checkConnection: %q is %q", name, setting)
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
		return fmt.Errorf("pgdb.Pool.checkConnection: %w", err)
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

// Collections returns a sorted list of FerretDB collection names.
func (pgPool *Pool) Collections(ctx context.Context, db string) ([]string, error) {
	schemaExists, err := pgPool.schemaExists(ctx, pgPool, db)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !schemaExists {
		return nil, ErrSchemaNotExist
	}

	var settings *types.Document
	var collections *types.Document

	err = pgPool.inTransaction(ctx, func(tx pgx.Tx) error {
		var serr error
		settings, serr = pgPool.getSettingsTable(ctx, tx, db)
		return serr
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	collectionsDoc := must.NotFail(settings.Get("collections"))

	var ok bool
	collections, ok = collectionsDoc.(*types.Document)
	if !ok {
		return nil, lazyerrors.Errorf("invalid settings document: %v", collectionsDoc)
	}

	return collections.Keys(), nil
}

// Tables returns a sorted list of PostgreSQL table names.
// Returns empty slice if schema does not exist.
// Tables with prefix "_ferretdb_" are filtered out.
func (pgPool *Pool) Tables(ctx context.Context, schema string) ([]string, error) {
	// TODO query settings table instead: https://github.com/FerretDB/FerretDB/issues/125

	var tables []string

	err := pgPool.inTransaction(ctx, func(tx pgx.Tx) error {
		var err error
		tables, err = pgPool.tables(ctx, tx, schema)
		if err != nil {
			return lazyerrors.Error(err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(tables))
	for _, table := range tables {
		if strings.HasPrefix(table, reservedCollectionPrefix) {
			continue
		}

		filtered = append(filtered, table)
	}

	return filtered, nil
}

// CreateDatabase creates a new FerretDB database (PostgreSQL schema).
//
// It returns (possibly wrapped) ErrAlreadyExist if schema already exist,
// use errors.Is to check the error.
func (pgPool *Pool) CreateDatabase(ctx context.Context, db string) error {
	err := pgPool.inTransaction(ctx, func(tx pgx.Tx) error {
		sql := `CREATE SCHEMA ` + pgx.Identifier{db}.Sanitize()
		_, err := tx.Exec(ctx, sql)

		if err == nil {
			err = pgPool.createSettingsTable(ctx, tx, db)
		}
		return err
	})

	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return lazyerrors.Error(err)
	}

	switch pgErr.Code {
	case pgerrcode.DuplicateSchema:
		return ErrAlreadyExist
	case pgerrcode.UniqueViolation, pgerrcode.DuplicateObject:
		// https://www.postgresql.org/message-id/CA+TgmoZAdYVtwBfp1FL2sMZbiHCWT4UPrzRLNnX1Nb30Ku3-gg@mail.gmail.com
		// The same thing for schemas. Reproducible by integration tests.
		return ErrAlreadyExist
	default:
		return lazyerrors.Error(err)
	}
}

// DropDatabase drops FerretDB database.
//
// It returns ErrTableNotExist if schema does not exist.
func (pgPool *Pool) DropDatabase(ctx context.Context, db string) error {
	sql := `DROP SCHEMA ` + pgx.Identifier{db}.Sanitize() + ` CASCADE`
	_, err := pgPool.Exec(ctx, sql)
	if err == nil {
		return nil
	}

	pgErr, ok := err.(*pgconn.PgError)
	if !ok {
		return lazyerrors.Error(err)
	}

	switch pgErr.Code {
	case pgerrcode.InvalidSchemaName:
		return ErrSchemaNotExist
	default:
		return lazyerrors.Error(err)
	}
}

// CreateCollection creates a new FerretDB collection in existing schema.
//
// It returns a possibly wrapped error:
//  * ErrInvalidTableName - if a FerretDB collection name doesn't conform to restrictions.
//  * ErrAlreadyExist - if a FerretDB collection with the given names already exists.
//  * ErrTableNotExist - is the required FerretDB database does not exist.
// Please use errors.Is to check the error.
func (pgPool *Pool) CreateCollection(ctx context.Context, querier pgxtype.Querier, db, collection string) error {
	if !validateCollectionNameRe.MatchString(collection) {
		return ErrInvalidTableName
	}

	if strings.HasPrefix(collection, reservedCollectionPrefix) {
		return ErrInvalidTableName
	}

	schemaExists, err := pgPool.schemaExists(ctx, querier, db)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if !schemaExists {
		return ErrSchemaNotExist
	}

	table := formatCollectionName(collection)
	tables, err := pgPool.tables(ctx, querier, db)
	if err != nil {
		return err
	}
	if slices.Contains(tables, table) {
		return ErrAlreadyExist
	}

	settings, err := pgPool.getSettingsTable(ctx, querier, db)
	if err != nil {
		return lazyerrors.Error(err)
	}

	collectionsDoc := must.NotFail(settings.Get("collections"))
	collections, ok := collectionsDoc.(*types.Document)
	if !ok {
		return lazyerrors.Errorf("expected document but got %[1]T: %[1]v", collectionsDoc)
	}

	if collections.Has(collection) {
		return nil
	}

	must.NoError(collections.Set(collection, table))
	must.NoError(settings.Set("collections", collections))

	err = pgPool.updateSettingsTable(ctx, querier, db, settings)
	if err != nil {
		return lazyerrors.Error(err)
	}

	sql := `CREATE TABLE IF NOT EXISTS ` + pgx.Identifier{db, table}.Sanitize() + ` (_jsonb jsonb)`
	_, err = querier.Exec(ctx, sql)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return lazyerrors.Error(err)
	}

	switch pgErr.Code {
	case pgerrcode.UniqueViolation, pgerrcode.DuplicateObject:
		// https://www.postgresql.org/message-id/CA+TgmoZAdYVtwBfp1FL2sMZbiHCWT4UPrzRLNnX1Nb30Ku3-gg@mail.gmail.com
		// Reproducible by integration tests.
		return ErrAlreadyExist
	default:
		return lazyerrors.Error(err)
	}
}

// DropCollection drops FerretDB collection.
//
// It returns (possibly wrapped) ErrTableNotExist if schema or table does not exist.
//  Please use errors.Is to check the error.
func (pgPool *Pool) DropCollection(ctx context.Context, schema, collection string) error {
	schemaExists, err := pgPool.schemaExists(ctx, pgPool, schema)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if !schemaExists {
		return ErrSchemaNotExist
	}

	table := formatCollectionName(collection)
	err = pgPool.inTransaction(ctx, func(tx pgx.Tx) error {
		tables, err := pgPool.tables(ctx, tx, schema)
		if err != nil {
			return lazyerrors.Error(err)
		}
		if !slices.Contains(tables, table) {
			return ErrTableNotExist
		}

		err = pgPool.removeTableFromSettings(ctx, tx, schema, collection)
		if err != nil && !errors.Is(err, ErrTableNotExist) {
			return lazyerrors.Error(err)
		}
		if errors.Is(err, ErrTableNotExist) {
			return ErrTableNotExist
		}

		// TODO https://github.com/FerretDB/FerretDB/issues/811
		sql := `DROP TABLE IF EXISTS` + pgx.Identifier{schema, table}.Sanitize() + `CASCADE`
		_, err = tx.Exec(ctx, sql)
		if err != nil {
			return lazyerrors.Error(err)
		}

		return nil
	})

	return err
}

// CreateTableIfNotExist ensures that given FerretDB database / PostgreSQL schema
// and FerretDB collection / PostgreSQL table exist.
// If needed, it creates both schema and table.
//
// True is returned if table was created.
func (pgPool *Pool) CreateTableIfNotExist(ctx context.Context, db, collection string) (bool, error) {
	exists, err := pgPool.CollectionExists(ctx, db, collection)
	if err != nil {
		return false, lazyerrors.Error(err)
	}
	if exists {
		return false, nil
	}

	// Table (or even schema) does not exist. Try to create it,
	// but keep in mind that it can be created in concurrent connection.

	if err := pgPool.CreateDatabase(ctx, db); err != nil && !errors.Is(err, ErrAlreadyExist) {
		return false, lazyerrors.Error(err)
	}

	// TODO use a transaction instead of pgPool: https://github.com/FerretDB/FerretDB/issues/866
	if err := pgPool.CreateCollection(ctx, pgPool, db, collection); err != nil {
		if errors.Is(err, ErrAlreadyExist) {
			return false, nil
		}
		return false, lazyerrors.Error(err)
	}

	return true, nil
}

// CollectionExists returns true if FerretDB collection exists.
func (pgPool *Pool) CollectionExists(ctx context.Context, db, collection string) (bool, error) {
	collections, err := pgPool.Collections(ctx, db)
	if err != nil {
		if errors.Is(err, ErrSchemaNotExist) {
			return false, nil
		}
		return false, err
	}

	return slices.Contains(collections, collection), nil
}

// SchemaStats returns a set of statistics for FerretDB database / PostgreSQL schema and table.
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
                                         AND i.tablename = t.table_name
     WHERE t.table_schema = $1`

	args := []any{schema}
	if collection != "" {
		sql = sql + " AND t.table_name = $2"
		args = append(args, collection)
	}

	res.Name = schema
	err := pgPool.QueryRow(ctx, sql, args...).
		Scan(&res.CountTables, &res.CountRows, &res.SizeTotal, &res.SizeIndexes, &res.SizeRelation, &res.CountIndexes)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return &res, nil
}

// QueryDocuments returns a list of documents for given FerretDB database and collection.
func (pgPool *Pool) QueryDocuments(ctx context.Context, db, collection, comment string) ([]*types.Document, error) {
	var res []*types.Document
	err := pgPool.inTransaction(ctx, func(tx pgx.Tx) error {
		table, err := pgPool.getTableName(ctx, tx, db, collection)
		if err != nil {
			return err
		}

		sql := `SELECT _jsonb `
		if comment != "" {
			comment = strings.ReplaceAll(comment, "/*", "/ *")
			comment = strings.ReplaceAll(comment, "*/", "* /")

			sql += `/* ` + comment + ` */ `
		}

		sql += `FROM ` + pgx.Identifier{db, table}.Sanitize()

		rows, err := tx.Query(ctx, sql)
		if err != nil {
			return lazyerrors.Error(err)
		}
		defer rows.Close()

		for rows.Next() {
			var b []byte
			if err := rows.Scan(&b); err != nil {
				return lazyerrors.Error(err)
			}

			doc, err := fjson.Unmarshal(b)
			if err != nil {
				return lazyerrors.Error(err)
			}

			res = append(res, doc.(*types.Document))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// SetDocumentByID sets a document by its ID.
func (pgPool *Pool) SetDocumentByID(ctx context.Context, db, collection string, id any, doc *types.Document) (int64, error) {
	var tag pgconn.CommandTag
	err := pgPool.inTransaction(ctx, func(tx pgx.Tx) error {
		table, err := pgPool.getTableName(ctx, tx, db, collection)
		if err != nil {
			return err
		}

		sql := "UPDATE " + pgx.Identifier{db, table}.Sanitize() +
			" SET _jsonb = $1 WHERE _jsonb->'_id' = $2"

		tag, err = tx.Exec(ctx, sql, must.NotFail(fjson.Marshal(doc)), must.NotFail(fjson.Marshal(id)))
		return err
	})
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

// DeleteDocumentsByID deletes documents by given IDs.
func (pgPool *Pool) DeleteDocumentsByID(ctx context.Context, db, collection string, ids []any) (int64, error) {
	var tag pgconn.CommandTag
	err := pgPool.inTransaction(ctx, func(tx pgx.Tx) error {
		table, err := pgPool.getTableName(ctx, tx, db, collection)
		if err != nil {
			return err
		}

		var p Placeholder
		idsMarshalled := make([]any, len(ids))
		placeholders := make([]string, len(ids))
		for i, id := range ids {
			placeholders[i] = p.Next()
			idsMarshalled[i] = must.NotFail(fjson.Marshal(id))
		}

		sql := `DELETE FROM ` + pgx.Identifier{db, table}.Sanitize() +
			` WHERE _jsonb->'_id' IN (` +
			strings.Join(placeholders, ", ") +
			`)`

		tag, err = tx.Exec(ctx, sql, idsMarshalled...)
		return err
	})
	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}

// InsertDocument inserts a document into FerretDB database and collection.
// If database or collection does not exist, it will be created.
func (pgPool *Pool) InsertDocument(ctx context.Context, db, collection string, doc *types.Document) error {
	exists, err := pgPool.CollectionExists(ctx, db, collection)
	if err != nil {
		return err
	}

	if !exists {
		if err := pgPool.CreateDatabase(ctx, db); err != nil && !errors.Is(err, ErrAlreadyExist) {
			return lazyerrors.Error(err)
		}

		// TODO use a transaction instead of pgPool: https://github.com/FerretDB/FerretDB/issues/866
		if err := pgPool.CreateCollection(ctx, pgPool, db, collection); err != nil {
			if errors.Is(err, ErrAlreadyExist) {
				return nil
			}
			return lazyerrors.Error(err)
		}
	}

	err = pgPool.inTransaction(ctx, func(tx pgx.Tx) error {
		table, err := pgPool.getTableName(ctx, tx, db, collection)
		if err != nil {
			return err
		}

		sql := `INSERT INTO ` + pgx.Identifier{db, table}.Sanitize() +
			` (_jsonb) VALUES ($1)`

		_, err = tx.Exec(ctx, sql, must.NotFail(fjson.Marshal(doc)))
		return err
	})

	return err
}

// tables returns a list of PostgreSQL table names.
func (pgPool *Pool) tables(ctx context.Context, querier pgxtype.Querier, schema string) ([]string, error) {
	sql := `SELECT table_name ` +
		`FROM information_schema.columns ` +
		`WHERE table_schema = $1 ` +
		`GROUP BY table_name ` +
		`ORDER BY table_name`
	rows, err := querier.Query(ctx, sql, schema)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	tables := make([]string, 0, 2)
	var name string
	for rows.Next() {
		if err = rows.Scan(&name); err != nil {
			return nil, lazyerrors.Error(err)
		}

		tables = append(tables, name)
	}
	if err = rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return tables, nil
}

// schemaExists returns true if given schema exists.
func (pgPool *Pool) schemaExists(ctx context.Context, querier pgxtype.Querier, db string) (bool, error) {
	sql := `SELECT nspname FROM pg_catalog.pg_namespace WHERE nspname = $1`
	rows, err := querier.Query(ctx, sql, db)
	if err != nil {
		return false, lazyerrors.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		must.NoError(rows.Scan(&name))
		if name == db {
			return true, nil
		}
	}

	return false, nil
}

// inTransaction wraps the given function f in a transaction.
// If f returns an error, the transaction is rolled back.
// Errors are wrapped with lazyerrors.Error,
// so the caller needs to use errors.Is to check the error,
// for example, errors.Is(err, ErrSchemaNotExist).
func (pgPool *Pool) inTransaction(ctx context.Context, f func(pgx.Tx) error) (err error) {
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
			pgPool.logger.Error("failed to perform rollback", zap.Error(rerr))
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
