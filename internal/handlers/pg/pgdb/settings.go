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
	"hash/fnv"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

const (
	// Reserved prefix for database and collection names.
	reservedPrefix = "_ferretdb_"

	// Settings table name.
	settingsTableName = reservedPrefix + "settings"

	// PostgreSQL max table name length.
	maxTableNameLength = 63
)

// upsertSettings returns PostgreSQL table name for the given FerretDB database and collection names.
// If such settings don't exist, it creates them.
//
// It makes a document with _id and table fields and stores it in the settingsTableName table.
// The given FerretDB collection name is stored in the _id field,
// the corresponding PostgreSQL table name is stored in the table field.
// For _id field it creates unique index.
//
// If a PostgreSQL conflict occurs, it returns a possible wrapped transactionConflictError error
// which indicates that the caller could retry the transaction.
func upsertSettings(ctx context.Context, tx pgx.Tx, db, collection string) (string, error) {
	var tableName string

	if err := CreateDatabaseIfNotExists(ctx, tx, db); err != nil {
		return "", lazyerrors.Error(err)
	}

	err := createTableIfNotExists(ctx, tx, db, settingsTableName)
	if err != nil {
		return "", err
	}

	// Index to ensure that collection name is unique
	err = createIndexIfNotExists(ctx, tx, indexParams{
		schema:   db,
		table:    settingsTableName,
		isUnique: true,
	})
	if err != nil {
		return "", err
	}

	tableName = formatCollectionName(collection)
	settings := must.NotFail(types.NewDocument(
		"_id", collection,
		"table", tableName,
	))

	if err := insert(ctx, tx, insertParams{
		schema: db,
		table:  settingsTableName,
		doc:    settings,
		upsert: true,
	}); err != nil {
		return "", lazyerrors.Error(err)
	}

	if err != nil {
		return "", err
	}

	return tableName, nil
}

// getSettings returns PostgreSQL table name for the given database and collection.
// If such settings don't exist, it returns an error.
func getSettings(ctx context.Context, tx pgx.Tx, db, collection string) (string, error) {
	it, err := buildIterator(ctx, tx, iteratorParams{
		schema: db,
		table:  settingsTableName,
		filter: must.NotFail(types.NewDocument("_id", collection)),
	})
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	_, doc, err := it.Next()
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	it.Close()

	table, err := doc.Get("table")
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	return table.(string), nil
}

// removeSettings removes settings for the given database and collection.
// If such settings don't exist, it doesn't return an error.
func removeSettings(ctx context.Context, tx pgx.Tx, db, collection string) error {
	_, err := deleteByIds(ctx, tx, deleteParams{
		schema: db,
		table:  settingsTableName,
	}, []any{collection})

	return err
}

// formatCollectionName returns collection name in form <shortened_name>_<name_hash>.
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
