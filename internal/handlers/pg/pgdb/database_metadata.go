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
	"github.com/FerretDB/FerretDB/internal/util/iterator"
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

	// PostgreSQL max index name length.
	maxIndexNameLength = 63
)

// metadata is a type to structure methods that work with metadata storing and getting.
//
// Metadata consists of collections and indexes settings.
type metadata struct {
	tx         pgx.Tx
	db         string
	collection string
}

// newMetadata returns a new instance of metadata for the given transaction, database and collection names.
func newMetadata(tx pgx.Tx, db, collection string) *metadata {
	return &metadata{
		tx:         tx,
		db:         db,
		collection: collection,
	}
}

// ensure returns PostgreSQL table name for the given FerretDB database and collection names.
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
func (m *metadata) ensure(ctx context.Context) (tableName string, created bool, err error) {
	tableName, err = m.getTableName(ctx)

	switch {
	case err == nil:
		// metadata already exist
		return

	case errors.Is(err, ErrTableNotExist):
		// metadata don't exist, do nothing
	default:
		return "", false, lazyerrors.Error(err)
	}

	err = CreateDatabaseIfNotExists(ctx, m.tx, m.db)

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, ErrInvalidDatabaseName):
		return
	default:
		return "", false, lazyerrors.Error(err)
	}

	if err = createPGTableIfNotExists(ctx, m.tx, m.db, dbMetadataTableName); err != nil {
		return "", false, lazyerrors.Error(err)
	}

	// Index to ensure that collection name is unique
	if err = createPGIndexIfNotExists(ctx, m.tx, m.db, dbMetadataTableName, dbMetadataIndexName, true); err != nil {
		return "", false, lazyerrors.Error(err)
	}

	tableName = formatCollectionName(m.collection)
	metadata := must.NotFail(types.NewDocument(
		"_id", m.collection,
		"table", tableName,
		"indexes", must.NotFail(types.NewArray()),
	))

	err = insert(ctx, m.tx, insertParams{
		schema: m.db,
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

// getTableName returns PostgreSQL table name for the given FerretDB database and collection.
//
// If such metadata don't exist, it returns ErrTableNotExist.
func (m *metadata) getTableName(ctx context.Context) (string, error) {
	doc, err := m.get(ctx, false)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	table := must.NotFail(doc.Get("table"))

	return table.(string), nil
}

// get returns metadata stored in the metadata table.
//
// If such metadata don't exist, it returns ErrTableNotExist.
func (m *metadata) get(ctx context.Context, forUpdate bool) (*types.Document, error) {
	metadataExist, err := tableExists(ctx, m.tx, m.db, dbMetadataTableName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !metadataExist {
		return nil, ErrTableNotExist
	}

	iterParams := &iteratorParams{
		schema:    m.db,
		table:     dbMetadataTableName,
		filter:    must.NotFail(types.NewDocument("_id", m.collection)),
		forUpdate: forUpdate,
	}

	iter, err := buildIterator(ctx, m.tx, iterParams)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer iter.Close()

	// call iterator only once as only one document is expected.
	_, doc, err := iter.Next()

	switch {
	case err == nil:
		return doc, nil
	case errors.Is(err, iterator.ErrIteratorDone):
		// no metadata found for the given collection name
		return nil, ErrTableNotExist
	default:
		return nil, lazyerrors.Error(err)
	}
}

// set sets metadata for the given database and collection.
//
// To avoid data race, set should be called only after getMetadata with forUpdate = true is called,
// so that the metadata table is locked correctly.
func (m *metadata) set(ctx context.Context, doc *types.Document) error {
	if _, err := setById(ctx, m.tx, m.db, dbMetadataTableName, "", m.collection, doc); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// remove removes metadata.
//
// If such metadata don't exist, it doesn't return an error.
func (m *metadata) remove(ctx context.Context) error {
	_, err := deleteByIDs(ctx, m.tx, execDeleteParams{
		schema: m.db,
		table:  dbMetadataTableName,
	}, []any{m.collection},
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
	must.NotFail(hash32.Write([]byte(name)))

	nameSymbolsLeft := maxTableNameLength - hash32.Size()*2 - 1
	truncateTo := len(name)

	if truncateTo > nameSymbolsLeft {
		truncateTo = nameSymbolsLeft
	}

	return name[:truncateTo] + "_" + fmt.Sprintf("%x", hash32.Sum([]byte{}))
}

// setIndex sets the index info in the metadata table.
// It returns a PostgreSQL table name and index name that can be used to create index.
// If the given index already exists, it doesn't return an error.
//
// Indexes are stored in the `indexes` array of metadata entry.
//
// It returns a possibly wrapped error:
//   - ErrTableNotExist - if m.collection doesn't exist.
func (m *metadata) setIndex(ctx context.Context, index string, key IndexKey, unique bool) (pgTable string, pgIndex string, err error) { //nolint:lll // for readability
	metadata, err := m.get(ctx, true)
	if err != nil {
		return "", "", err
	}

	pgTable = must.NotFail(metadata.Get("table")).(string)
	pgIndex = formatIndexName(m.collection, index)

	indKey := types.MakeDocument(len(key))
	for _, pair := range key {
		indKey.Set(pair.Field, int32(pair.Order)) // order is set as int32 to be pjson-marshaled correctly
	}

	newIndex := must.NotFail(types.NewDocument(
		"pgindex", pgIndex,
		"name", index,
		"key", indKey,
		"unique", unique,
	))

	var indexes *types.Array
	if metadata.Has("indexes") {
		indexes = must.NotFail(metadata.Get("indexes")).(*types.Array)

		iter := indexes.Iterator()
		defer iter.Close()

		for {
			var idx any

			if _, idx, err = iter.Next(); err != nil {
				if errors.Is(err, iterator.ErrIteratorDone) {
					break
				}

				return "", "", lazyerrors.Error(err)
			}

			idxData := idx.(*types.Document)

			// if index name matches, return existing index
			idxName := must.NotFail(idxData.Get("name")).(string)

			if idxName == index {
				return pgTable, pgIndex, nil
			}

			// if index key matches, return existing index
			idxKey := must.NotFail(idxData.Get("key")).(*types.Document)

			if types.Compare(idxKey, indKey) == types.Equal {
				return pgTable, pgIndex, nil
			}
		}
	}

	indexes.Append(newIndex)
	metadata.Set("indexes", indexes)

	if err = m.set(ctx, metadata); err != nil {
		return "", "", lazyerrors.Error(err)
	}

	return
}

// formatIndexName returns index name in form <shortened_name>_<name_hash>_idx.
// Changing this logic will break compatibility with existing databases.
func formatIndexName(collection, index string) string {
	name := collection + "_" + index

	hash32 := fnv.New32a()
	must.NotFail(hash32.Write([]byte(name)))

	nameSymbolsLeft := maxIndexNameLength - hash32.Size()*2 - 5 // 5 is for "_" delimiter and "_idx" suffix
	truncateTo := len(name)

	if truncateTo > nameSymbolsLeft {
		truncateTo = nameSymbolsLeft
	}

	return name[:truncateTo] + "_" + fmt.Sprintf("%x", hash32.Sum([]byte{})) + "_idx"
}
