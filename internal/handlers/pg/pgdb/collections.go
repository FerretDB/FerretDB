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
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/AlekSi/pointer"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// validateCollectionNameRe validates collection names.
// Empty collection name, names with `$` and `\x00`,
// or exceeding the 235 bytes limit are not allowed.
// Collection names that start with `.` are also not allowed.
var validateCollectionNameRe = regexp.MustCompile("^[^\\.$\x00][^$\x00]{0,234}$")

// Collections returns a sorted list of FerretDB collection names.
//
// It returns (possibly wrapped) ErrSchemaNotExist if FerretDB database / PostgreSQL schema does not exist.
func Collections(ctx context.Context, tx pgx.Tx, db string) ([]string, error) {
	metadataExist, err := tableExists(ctx, tx, db, dbMetadataTableName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// if metadata table doesn't exist, there are no collections in the database
	if !metadataExist {
		return []string{}, nil
	}

	iter, _, err := buildIterator(ctx, tx, &iteratorParams{
		schema: db,
		table:  dbMetadataTableName,
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var collections []string

	defer iter.Close()

	for {
		var doc *types.Document
		_, doc, err = iter.Next()

		// if the context is canceled, we don't need to continue processing documents
		if ctx.Err() != nil {
			return nil, context.Cause(ctx)
		}

		switch {
		case err == nil:
			// do nothing
		case errors.Is(err, iterator.ErrIteratorDone):
			// no more documents
			slices.Sort(collections)
			return collections, nil
		default:
			return nil, lazyerrors.Error(err)
		}

		collections = append(collections, must.NotFail(doc.Get("_id")).(string))
	}
}

// CollectionExists returns true if FerretDB collection exists.
func CollectionExists(ctx context.Context, tx pgx.Tx, db, collection string) (bool, error) {
	_, err := newMetadataStorage(tx, db, collection).getTableName(ctx)

	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, ErrTableNotExist):
		return false, nil
	default:
		return false, lazyerrors.Error(err)
	}
}

// CreateCollection creates a new FerretDB collection with the given name in the given database.
// If the database does not exist, it will be created.
// It also creates a unique index on the _id field.
//
// It returns possibly wrapped error:
//   - ErrInvalidDatabaseName - if the given database name doesn't conform to restrictions.
//   - ErrInvalidCollectionName - if the given collection name doesn't conform to restrictions.
//   - ErrCollectionStartsWithDot - if the given collection name starts with dot.
//   - ErrAlreadyExist - if a FerretDB collection with the given name already exists.
//   - *transactionConflictError - if a PostgreSQL conflict occurs (the caller could retry the transaction).
func CreateCollection(ctx context.Context, tx pgx.Tx, db, collection string) error {
	if strings.HasPrefix(collection, ".") {
		return ErrCollectionStartsWithDot
	}

	if !validateCollectionNameRe.MatchString(collection) ||
		strings.HasPrefix(collection, reservedPrefix) ||
		!utf8.ValidString(collection) {
		return ErrInvalidCollectionName
	}

	table, created, err := newMetadataStorage(tx, db, collection).store(ctx)
	if err != nil {
		return lazyerrors.Error(err)
	}

	if !created {
		return ErrAlreadyExist
	}

	if err = createTableIfNotExists(ctx, tx, db, table); err != nil {
		return lazyerrors.Error(err)
	}

	// Create default index on _id field.
	indexParams := &Index{
		Name:   "_id_",
		Key:    IndexKey{{Field: "_id", Order: types.Ascending}},
		Unique: pointer.ToBool(true),
	}

	if _, err := CreateIndexIfNotExists(ctx, tx, db, collection, indexParams); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// CreateCollectionIfNotExists ensures that given FerretDB database and collection exist.
// If the database does not exist, it will be created.
//
// It returns true if the collection was created, false if it already existed or an error occurred.
//
// It returns possibly wrapped error:
//   - ErrInvalidDatabaseName - if the given database name doesn't conform to restrictions.
//   - ErrInvalidCollectionName - if the given collection name doesn't conform to restrictions.
//   - *transactionConflictError - if a PostgreSQL conflict occurs (the caller could retry the transaction).
func CreateCollectionIfNotExists(ctx context.Context, tx pgx.Tx, db, collection string) (bool, error) {
	err := CreateCollection(ctx, tx, db, collection)

	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, ErrAlreadyExist):
		return false, nil
	default:
		return false, lazyerrors.Error(err)
	}
}

// DropCollection drops FerretDB collection.
//
// It returns (possibly wrapped) ErrTableNotExist if database or collection does not exist.
// Please use errors.Is to check the error.
//
// Test correctness for concurrent cases.
// TODO https://github.com/FerretDB/FerretDB/issues/1684
func DropCollection(ctx context.Context, tx pgx.Tx, db, collection string) error {
	ms := newMetadataStorage(tx, db, collection)
	tableName, err := ms.getTableName(ctx)
	if err != nil {
		return err
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/811
	sql := `DROP TABLE IF EXISTS ` + pgx.Identifier{db, tableName}.Sanitize() + ` CASCADE`
	if _, err = tx.Exec(ctx, sql); err != nil {
		return lazyerrors.Error(err)
	}

	if err = ms.remove(ctx); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// RenameCollection changes the name of an existing collection.
//
// It returns ErrTableNotExist if either source database or collection does not exist.
// It returns ErrAlreadyExist if the target database or collection already exists.
// It returns ErrInvalidCollectionName if collection name is not valid.
func RenameCollection(ctx context.Context, tx pgx.Tx, db, collectionFrom, collectionTo string) error {
	if !validateCollectionNameRe.MatchString(collectionTo) ||
		strings.HasPrefix(collectionTo, reservedPrefix) ||
		!utf8.ValidString(collectionTo) ||
		len(collectionTo) > maxTableNameLength {
		return ErrInvalidCollectionName
	}

	return newMetadataStorage(tx, db, collectionFrom).renameCollection(ctx, collectionTo)
}

// createTableIfNotExists creates the given PostgreSQL table in the given schema if the table doesn't exist.
// If the table already exists, it does nothing.
//
// If a PostgreSQL conflict occurs it returns errTransactionConflict, and the caller could retry the transaction.
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
