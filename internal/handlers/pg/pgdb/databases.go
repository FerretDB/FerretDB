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
	"regexp"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// validateDatabaseNameRe validates FerretDB database / PostgreSQL schema names.
var validateDatabaseNameRe = regexp.MustCompile("^[a-z_-][a-z0-9_-]{0,62}$")

// Databases returns a sorted list of FerretDB databases / PostgreSQL schemas.
func Databases(ctx context.Context, tx pgx.Tx) ([]string, error) {
	sql := "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name"
	rows, err := tx.Query(ctx, sql)
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

// CreateDatabaseIfNotExists creates a new FerretDB database (PostgreSQL schema).
//
// If a PostgreSQL conflict occurs it returns *transactionConflictError, and the caller could retry the transaction.
func CreateDatabaseIfNotExists(ctx context.Context, tx pgx.Tx, db string) error {
	if !validateDatabaseNameRe.MatchString(db) ||
		strings.HasPrefix(db, reservedPrefix) {
		return ErrInvalidDatabaseName
	}

	_, err := tx.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS `+pgx.Identifier{db}.Sanitize())
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return lazyerrors.Error(err)
	}

	switch pgErr.Code {
	case pgerrcode.DuplicateSchema, pgerrcode.UniqueViolation, pgerrcode.DuplicateObject:
		// https://www.postgresql.org/message-id/CA+TgmoZAdYVtwBfp1FL2sMZbiHCWT4UPrzRLNnX1Nb30Ku3-gg@mail.gmail.com
		// The same thing for schemas. Reproducible by integration tests.
		return newTransactionConflictError(err)
	default:
		return lazyerrors.Error(err)
	}
}

// DropDatabase drops FerretDB database.
//
// It returns ErrSchemaNotExist if schema does not exist.
func DropDatabase(ctx context.Context, tx pgx.Tx, db string) error {
	_, err := tx.Exec(ctx, `DROP SCHEMA `+pgx.Identifier{db}.Sanitize()+` CASCADE`)
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

// DatabaseSize returns the size of the current database in bytes.
func DatabaseSize(ctx context.Context, tx pgx.Tx) (int64, error) {
	var size int64

	err := tx.QueryRow(ctx, "SELECT pg_database_size(current_database())").Scan(&size)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// TablesSize returns the sum of sizes of all tables in the given database in bytes.
func (pgPool *Pool) TablesSize(ctx context.Context, tx pgx.Tx, db string) (int64, error) {
	tables, err := tablesFiltered(ctx, tx, db)
	if err != nil {
		return 0, err
	}

	// iterate over result to collect sizes
	var sizeOnDisk int64

	for _, name := range tables {
		var tableSize int64
		fullName := pgx.Identifier{db, name}.Sanitize()
		// If the table was deleted after we got the list of tables, pg_total_relation_size will return null.
		// We use COALESCE to scan this null value as 0 in this case.
		// Even though we run the query in a transaction, the current isolation level doesn't guarantee
		// that the table is not deleted (see https://www.postgresql.org/docs/14/transaction-iso.html).
		// PgPool (not a transaction) is used on purpose here. In this case, transaction doesn't lock
		// relations, and it's possible that the table/schema is deleted between the moment we get the list of tables
		// and the moment we get the size of the table. In this case, we might receive an error from the database,
		// and transaction will be interrupted. Such errors are not critical, we can just ignore them, and
		// we don't need to interrupt the whole transaction.
		err = pgPool.p.QueryRow(ctx, "SELECT COALESCE(pg_total_relation_size($1), 0)", fullName).Scan(&tableSize)
		if err == nil {
			sizeOnDisk += tableSize
			continue
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UndefinedTable, pgerrcode.InvalidSchemaName:
				// Table or schema was deleted after we got the list of tables, just ignore it
				continue
			}
		}

		return 0, err
	}

	return sizeOnDisk, nil
}
