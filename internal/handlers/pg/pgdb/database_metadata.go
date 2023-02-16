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

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// Reserved prefix for database and collection names.
	reservedPrefix = "_ferretdb_"

	// Database metadata table name.
	dbMetadataTableName = reservedPrefix + "database_metadata"

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
	tableName, err = getMetadata(ctx, tx, db, collection)

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
	if err := createIndexIfNotExists(ctx, tx, indexParams{
		schema:   db,
		table:    dbMetadataTableName,
		isUnique: true,
	}); err != nil {
		return "", false, lazyerrors.Error(err)
	}

	tableName = formatCollectionName(collection)
	metadata := must.NotFail(types.NewDocument(
		"_id", collection,
		"table", tableName,
	))

	if err := insert(ctx, tx, insertParams{
		schema: db,
		table:  dbMetadataTableName,
		doc:    metadata,
	}); err != nil {
		return "", false, lazyerrors.Error(err)
	}

	return tableName, true, nil
}

// getMetadata returns PostgreSQL table name for the given FerretDB database and collection.
//
// If such metadata don't exist, it returns ErrTableNotExist.
func getMetadata(ctx context.Context, tx pgx.Tx, db, collection string) (string, error) {
	metadataExist, err := tableExists(ctx, tx, db, dbMetadataTableName)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	if !metadataExist {
		return "", ErrTableNotExist
	}

	queryById := func() (*types.Document, error) {
		query := `SELECT _jsonb FROM ` + pgx.Identifier{db, dbMetadataTableName}.Sanitize()

		where, args := prepareWhereClause(must.NotFail(types.NewDocument("_id", collection)))
		query += where

		var b []byte
		err := tx.QueryRow(ctx, query, args...).Scan(&b)

		switch {
		case err == nil:
			// do nothing
		case errors.Is(err, pgx.ErrNoRows):
			return nil, nil
		default:
			return nil, lazyerrors.Error(err)
		}

		doc, err := pjson.Unmarshal(b)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil
	}

	doc, err := queryById()
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	if doc == nil {
		// no metadata found for the given collection name
		return "", ErrTableNotExist
	}

	table := must.NotFail(doc.Get("table"))

	return table.(string), nil
}

// removeMetadata removes metadata for the given database and collection.
//
// If such metadata don't exist, it doesn't return an error.
func removeMetadata(ctx context.Context, tx pgx.Tx, db, collection string) error {
	_, err := deleteByIDs(ctx, tx, deleteParams{
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
