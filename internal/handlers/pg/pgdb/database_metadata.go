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
	"hash/fnv"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// Reserved prefix for database and collection names.
	reservedPrefix = "_ferretdb_"

	// Database metadata table name.
	dbMetadataTableName = reservedPrefix + "database_metadata"

	// Database metadata table unique _id index name.
	dbMetadataIndexName = dbMetadataTableName + "_id_idx"

	// PostgreSQL max table name length.
	maxTableNameLength = 63
)

// ensureMetadata returns PostgreSQL table name for the given FerretDB database and collection names.
// If such metadata don't exist, it creates them, including the creation of the PostgreSQL schema if needed.
// If metadata were created, it returns true as the second return value. If metadata already existed, it returns false.
//
// It makes a document with _id and table fields and stores it in the dbMetadataTableName table.
// The given FerretDB collection name is stored in the _id field,
// the corresponding PostgreSQL table name is stored in the table field.
// For _id field it creates unique index.
//
// It returns a possibly wrapped error:
//   - ErrInvalidDatabaseName - if the given database name doesn't conform to restrictions.
//   - *transactionConflictError - if a PostgreSQL conflict occurs (the caller could retry the transaction).
func ensureMetadata(ctx context.Context, tx pgx.Tx, db, collection string) (tableName string, created bool, err error) {
	tableName, err = getTableNameFromMetadata(ctx, tx, db, collection)

	switch {
	case err == nil:
		// metadata already exist
		return

	case errors.Is(err, ErrTableNotExist):
		// metadata don't exist, do nothing
	default:
		return "", false, lazyerrors.Error(err)
	}

	err = CreateDatabaseIfNotExists(ctx, tx, db)

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, ErrInvalidDatabaseName):
		return
	default:
		return "", false, lazyerrors.Error(err)
	}

	if err := createTableIfNotExists(ctx, tx, db, dbMetadataTableName); err != nil {
		return "", false, lazyerrors.Error(err)
	}

	// Index to ensure that collection name is unique
	if err = createIndexIfNotExists(ctx, tx, db, dbMetadataTableName, dbMetadataIndexName, true); err != nil {
		return "", false, lazyerrors.Error(err)
	}

	tableName = formatCollectionName(collection)
	metadata := must.NotFail(types.NewDocument(
		"_id", collection,
		"table", tableName,
		"indexes", must.NotFail(types.NewArray()),
	))

	err = insert(ctx, tx, insertParams{
		schema: db,
		table:  dbMetadataTableName,
		doc:    metadata,
	})

	switch {
	case err == nil:
		return tableName, true, nil
	case errors.Is(err, ErrUniqueViolation):
		// If metadata were created by another transaction we consider it transaction conflict error
		// to mark that transaction should be retried.
		return "", false, lazyerrors.Error(newTransactionConflictError(err))
	default:
		return "", false, lazyerrors.Error(err)
	}
}

// getTableNameFromMetadata returns PostgreSQL table name for the given FerretDB database and collection.
//
// If such metadata don't exist, it returns ErrTableNotExist.
func getTableNameFromMetadata(ctx context.Context, tx pgx.Tx, db, collection string) (string, error) {
	doc, err := getMetadata(ctx, tx, db, collection, false)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	table := must.NotFail(doc.Get("table"))

	return table.(string), nil
}

// getMetadata returns metadata for the given database and collection.
//
// If such metadata don't exist, it returns ErrTableNotExist.
func getMetadata(ctx context.Context, tx pgx.Tx, db, collection string, forUpdate bool) (*types.Document, error) {
	metadataExist, err := tableExists(ctx, tx, db, dbMetadataTableName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !metadataExist {
		return nil, ErrTableNotExist
	}

	doc, err := queryById(ctx, tx, db, dbMetadataTableName, collection, forUpdate)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if doc == nil {
		// no metadata found for the given collection name
		return nil, ErrTableNotExist
	}

	return doc, nil
}

// setMetadata sets metadata for the given database and collection.
//
// To avoid data race, setMetadata should be called only after getMetadata with forUpdate = true is called,
// so that the metadata table is locked correctly.
func setMetadata(ctx context.Context, tx pgx.Tx, db, collection string, metadata *types.Document) error {
	if _, err := setById(ctx, tx, db, dbMetadataTableName, "", collection, metadata); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// removeMetadata removes metadata for the given database and collection.
//
// If such metadata don't exist, it doesn't return an error.
func removeMetadata(ctx context.Context, tx pgx.Tx, db, collection string) error {
	_, err := deleteByIDs(ctx, tx, execDeleteParams{
		schema: db,
		table:  dbMetadataTableName,
	}, []any{collection},
	)

	if err == nil {
		return nil
	}

	return lazyerrors.Error(err)
}

// formatCollectionName returns collection name in form <shortened_name>_<name_hash>.
// Changing this logic will break compatibility with existing databases.
func formatCollectionName(name string) string {
	hash32 := fnv.New32a()
	_ = must.NotFail(hash32.Write([]byte(name)))

	nameSymbolsLeft := maxTableNameLength - hash32.Size()*2 - 1
	truncateTo := len(name)

	if truncateTo > nameSymbolsLeft {
		truncateTo = nameSymbolsLeft
	}

	return name[:truncateTo] + "_" + fmt.Sprintf("%x", hash32.Sum([]byte{}))
}
