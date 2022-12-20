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

	"github.com/FerretDB/FerretDB/internal/types"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// validateCollectionNameRe validates collection names.
var validateCollectionNameRe = regexp.MustCompile("^[a-zA-Z_-][a-zA-Z0-9_-]{0,119}$")

// Collections returns a sorted list of FerretDB collection names.
//
// It returns (possibly wrapped) ErrSchemaNotExist if FerretDB database / PostgreSQL schema does not exist.
func Collections(ctx context.Context, tx pgx.Tx, db string) ([]string, error) {
	it, err := buildIterator(ctx, tx, iteratorParams{
		schema: db,
		table:  settingsTableName,
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var collections []string
	defer slices.Sort(collections) // sort the result before returning

	defer it.Close()

	for {
		var doc *types.Document
		_, doc, err = it.Next()

		// if the context is canceled, we don't need to continue processing documents
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		switch {
		case err == nil:
			// do nothing
		case errors.Is(err, iterator.ErrIteratorDone):
			// no more documents
			return collections, nil
		default:
			return nil, err
		}

		collections = append(collections, must.NotFail(doc.Get("collection")).(string))
	}
}

// CollectionExists returns true if FerretDB collection exists.
func CollectionExists(ctx context.Context, tx pgx.Tx, db, collection string) (bool, error) {
	collections, err := Collections(ctx, tx, db)
	if err != nil {
		if errors.Is(err, ErrSchemaNotExist) {
			return false, nil
		}
		return false, err
	}

	return slices.Contains(collections, collection), nil
}

// CreateCollection creates a new FerretDB collection in existing database.
//
// It returns a possibly wrapped error:
//   - ErrInvalidTableName - if a FerretDB collection name doesn't conform to restrictions.
//   - ErrAlreadyExist - if a FerretDB collection with the given names already exists.
//   - ErrTableNotExist - is the required FerretDB database does not exist.
//
// Please use errors.Is to check the error.
func CreateCollection(ctx context.Context, tx pgx.Tx, db, collection string) error {
	if !validateCollectionNameRe.MatchString(collection) ||
		strings.HasPrefix(collection, reservedPrefix) {
		return ErrInvalidTableName
	}

	schemaExists, err := schemaExists(ctx, tx, db)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if !schemaExists {
		return ErrSchemaNotExist
	}

	table, err := addSettingsIfNotExists(ctx, tx, db, collection)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return createTableIfNotExists(ctx, tx, db, table)
}

// CreateCollectionIfNotExist ensures that given FerretDB database / PostgreSQL schema
// and FerretDB collection / PostgreSQL table exist.
// If needed, it creates both database and collection.
//
// True is returned if collection was created.
func CreateCollectionIfNotExist(ctx context.Context, pgPool *Pool, db, collection string) (bool, error) {
	var exists bool
	err := pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		var err error
		exists, err = CollectionExists(ctx, tx, db, collection)
		return err
	})
	if err != nil {
		return false, lazyerrors.Error(err)
	}

	if exists {
		return false, nil
	}

	// Collection (or even database) does not exist. Try to create them,
	// but keep in mind that it can be created in concurrent connection.

	err = pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		if err = CreateDatabaseIfNotExists(ctx, tx, db); err != nil {
			return err
		}
		return nil
	})
	if err != nil && !errors.Is(err, ErrAlreadyExist) {
		return false, lazyerrors.Error(err)
	}

	err = pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		err = CreateCollection(ctx, tx, db, collection)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil && !errors.Is(err, ErrAlreadyExist) {
		return false, lazyerrors.Error(err)
	}

	return true, nil
}

// DropCollection drops FerretDB collection.
//
// It returns (possibly wrapped) ErrTableNotExist if database or collection does not exist.
// Please use errors.Is to check the error.
func DropCollection(ctx context.Context, tx pgx.Tx, db, collection string) error {
	schemaExists, err := schemaExists(ctx, tx, db)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if !schemaExists {
		return ErrSchemaNotExist
	}

	table := formatCollectionName(collection)
	tables, err := tables(ctx, tx, db)
	if err != nil {
		return lazyerrors.Error(err)
	}
	if !slices.Contains(tables, table) {
		return ErrTableNotExist
	}

	err = removeSettings(ctx, tx, db, collection)
	if err != nil && !errors.Is(err, ErrTableNotExist) {
		return lazyerrors.Error(err)
	}
	if errors.Is(err, ErrTableNotExist) {
		return ErrTableNotExist
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/811
	sql := `DROP TABLE IF EXISTS ` + pgx.Identifier{db, table}.Sanitize() + ` CASCADE`
	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// createTableIfNotExists creates the given PostgreSQL table in the given schema if the table doesn't exist.
// If the table doesn't exist, it creates it.
// If the table already exists, it does nothing.
// If PostgreSQL can't create a table due to a concurrent connection (conflict), it returns errTransactionConflict.
// Otherwise, it returns some other possibly wrapped error.
func createTableIfNotExists(ctx context.Context, tx pgx.Tx, schema, table string) error {
	var err error

	sql := `CREATE TABLE IF NOT EXISTS ` + pgx.Identifier{schema, table}.Sanitize() + ` (_jsonb jsonb)`
	if _, err = tx.Exec(ctx, sql); err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return lazyerrors.Error(err)
	}

	switch pgErr.Code {
	case pgerrcode.UniqueViolation, pgerrcode.DuplicateObject, pgerrcode.DuplicateTable:
		// https://www.postgresql.org/message-id/CA+TgmoZAdYVtwBfp1FL2sMZbiHCWT4UPrzRLNnX1Nb30Ku3-gg@mail.gmail.com
		// Reproducible by integration tests.
		return newTransactionConflictError(err)
	default:
		return lazyerrors.Error(err)
	}
}
