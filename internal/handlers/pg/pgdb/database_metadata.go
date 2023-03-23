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
	"regexp"

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

// specialCharacters are potential problematic characters of pg table name
// that are replaced with `_`.
var specialCharacters = regexp.MustCompile("[^a-z][^a-z0-9_]*")

// metadataStorage offers methods to store and get metadata for the given database and collection.
type metadataStorage struct {
	tx         pgx.Tx
	db         string
	collection string
}

// metadata stores information about FerretDB collections and indexes.
type metadata struct {
	table   string          // Corresponding PostgreSQL table name
	indexes []metadataIndex // List of FerretDB indexes for the collection
}

// metadataIndex stores information about FerretDB indexes.
type metadataIndex struct {
	name    string   // FerretDB index name
	pgindex string   // Corresponding PostgreSQL index name
	key     IndexKey // Index specification (field name + sort order pairs)
	unique  bool     // Whether the index is unique
}

// newMetadataStorage returns a new instance of metadata for the given transaction, database and collection names.
func newMetadataStorage(tx pgx.Tx, db, collection string) *metadataStorage {
	return &metadataStorage{
		tx:         tx,
		db:         db,
		collection: collection,
	}
}

// store returns PostgreSQL table name for the given FerretDB database and collection names.
// It returns the table name, whether the metadata were created and an error if creation failed.
//
// If the metadata for the given settings don't exist, it creates them, including the creation
// of the PostgreSQL schema if needed.
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
func (m *metadataStorage) store(ctx context.Context) (string, bool, error) {
	var metadata *metadata
	metadata, err := m.get(ctx, false)

	switch {
	case err == nil:
		// metadata already exist
		return metadata.table, false, nil

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
		return "", false, err
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

	tableName := collectionNameToTableName(m.collection)
	doc := must.NotFail(types.NewDocument(
		"_id", m.collection,
		"table", tableName,
		"indexes", must.NotFail(types.NewArray()),
	))

	err = insert(ctx, m.tx, insertParams{
		schema: m.db,
		table:  dbMetadataTableName,
		doc:    doc,
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
func (m *metadataStorage) getTableName(ctx context.Context) (string, error) {
	metadata, err := m.get(ctx, false)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	return metadata.table, nil
}

// get returns metadata stored in the metadata table.
//
// If such metadata don't exist, it returns ErrTableNotExist.
func (m *metadataStorage) get(ctx context.Context, forUpdate bool) (*metadata, error) {
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
		return documentToMetadata(doc)

	case errors.Is(err, iterator.ErrIteratorDone):
		// no metadata found for the given collection name
		return nil, ErrTableNotExist

	default:
		return nil, lazyerrors.Error(err)
	}
}

// documentToMetadata converts *types.Document to metadata.
// Use this function if you want to transform document returned from the database to structured metadata.
func documentToMetadata(doc *types.Document) (*metadata, error) {
	var indexesArr *types.Array

	if val, err := doc.Get("indexes"); err != nil {
		// if there is no indexes field, consider it empty
		indexesArr = must.NotFail(types.NewArray())
	} else {
		indexesArr = val.(*types.Array)
	}

	indexes := make([]metadataIndex, indexesArr.Len())

	for i := 0; i < indexesArr.Len(); i++ {
		idxDoc := must.NotFail(indexesArr.Get(i)).(*types.Document)

		keyDoc := must.NotFail(idxDoc.Get("key")).(*types.Document)
		key := make(IndexKey, keyDoc.Len())

		keyIter := keyDoc.Iterator()
		defer keyIter.Close() // it's safe to defer here as we always read the whole iterator

		for j := 0; j < keyDoc.Len(); j++ {
			var field string
			var value any
			field, value, err := keyIter.Next()

			switch {
			case err == nil:
				key[j] = IndexKeyPair{
					Field: field,
					Order: IndexOrder(value.(int32)),
				}
			default:
				return nil, lazyerrors.Error(err)
			}
		}

		indexes[i] = metadataIndex{
			name:    must.NotFail(idxDoc.Get("name")).(string),
			pgindex: must.NotFail(idxDoc.Get("pgindex")).(string),
			key:     key,
			unique:  must.NotFail(idxDoc.Get("unique")).(bool),
		}
	}

	return &metadata{
		table:   must.NotFail(doc.Get("table")).(string),
		indexes: indexes,
	}, nil
}

// set sets metadata for the given database and collection.
//
// To avoid data race, set should be called only after getMetadata with forUpdate = true is called,
// so that the metadata table is locked correctly.
func (m *metadataStorage) set(ctx context.Context, metadata *metadata) error {
	doc := m.metadataToDocument(metadata)

	if _, err := setById(ctx, m.tx, m.db, dbMetadataTableName, "", m.collection, doc); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// metadataToDocument converts metadata to *types.Document.
// Use this function to transform metadata to document to be stored in the database.
func (m *metadataStorage) metadataToDocument(metadata *metadata) *types.Document {
	indexesArr := types.MakeArray(len(metadata.indexes))

	for _, idx := range metadata.indexes {
		keyDoc := types.MakeDocument(len(idx.key))
		for _, pair := range idx.key {
			keyDoc.Set(pair.Field, int32(pair.Order)) // order is set as int32 to be pjson-marshaled correctly
		}

		indexesArr.Append(must.NotFail(types.NewDocument(
			"pgindex", idx.pgindex,
			"name", idx.name,
			"key", keyDoc,
			"unique", idx.unique,
		)))
	}

	return must.NotFail(types.NewDocument(
		"_id", m.collection,
		"table", metadata.table,
		"indexes", indexesArr,
	))
}

// remove removes metadata.
//
// If such metadata don't exist, it doesn't return an error.
func (m *metadataStorage) remove(ctx context.Context) error {
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

// collectionNameToTableName returns name in form <shortened_name>_<name_hash>.
// It replaces special characters with `_`.
//
// Deprecated: this function usage is allowed for collection metadata creation only.
func collectionNameToTableName(name string) string {
	hash32 := fnv.New32a()
	must.NotFail(hash32.Write([]byte(name)))

	mangled := specialCharacters.ReplaceAllString(name, "_")

	nameSymbolsLeft := maxTableNameLength - hash32.Size()*2 - 1
	truncateTo := len(mangled)

	if truncateTo > nameSymbolsLeft {
		truncateTo = nameSymbolsLeft
	}

	return mangled[:truncateTo] + "_" + fmt.Sprintf("%08x", hash32.Sum(nil))
}

// setIndex sets the index info in the metadata table.
// It returns a PostgreSQL table name and index name that can be used to create index.
//
// Indexes are stored in the `indexes` array of metadata entry.
//
// It returns a possibly wrapped error:
//   - ErrTableNotExist - if the metadata table doesn't exist.
//   - ErrIndexAlreadyExist - if the given index already exists.
func (m *metadataStorage) setIndex(ctx context.Context, index string, key IndexKey, unique bool) (pgTable string, pgIndex string, err error) { //nolint:lll // for readability
	metadata, err := m.get(ctx, true)
	if err != nil {
		return "", "", err
	}

	pgTable = metadata.table
	pgIndex = formatIndexName(m.collection, index)

	if len(metadata.indexes) > 0 {
		for _, idx := range metadata.indexes {
			if idx.name == index {
				return "", "", ErrIndexAlreadyExist
			}
		}
	}

	metadata.indexes = append(metadata.indexes, metadataIndex{
		name:    index,
		pgindex: pgIndex,
		key:     key,
		unique:  unique,
	})

	if err = m.set(ctx, metadata); err != nil {
		return "", "", lazyerrors.Error(err)
	}

	return
}

// formatIndexName returns index name in form <shortened_name>_<name_hash>_idx.
//
// Deprecated: this function usage is allowed for index metadata creation only.
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
