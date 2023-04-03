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

// metadata stores information about FerretDB collection and indexes.
type metadata struct {
	collection string // _id
	table      string
	indexes    []metadataIndex
}

// metadataIndex stores information about FerretDB index.
type metadataIndex struct {
	pgIndex string
	Index
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
func (ms *metadataStorage) store(ctx context.Context) (tableName string, created bool, err error) {
	var m *metadata
	m, err = ms.get(ctx, false)

	switch {
	case err == nil:
		// metadata already exist
		tableName = m.table
		return

	case errors.Is(err, ErrTableNotExist):
		// metadata don't exist, do nothing
	default:
		err = lazyerrors.Error(err)
		return
	}

	err = CreateDatabaseIfNotExists(ctx, ms.tx, ms.db)

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, ErrInvalidDatabaseName):
		return
	default:
		err = lazyerrors.Error(err)
		return
	}

	if err = createTableIfNotExists(ctx, ms.tx, ms.db, dbMetadataTableName); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	// Index to ensure that collection name is unique
	key := IndexKey{{Field: `_id`, Order: types.Ascending}}
	if err = createPgIndexIfNotExists(ctx, ms.tx, ms.db, dbMetadataTableName, dbMetadataIndexName, key, true); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	tableName = collectionNameToTableName(ms.collection)
	m = &metadata{
		collection: ms.collection,
		table:      tableName,
		indexes:    []metadataIndex{},
	}

	err = insert(ctx, ms.tx, &insertParams{
		schema: ms.db,
		table:  dbMetadataTableName,
		doc:    metadataToDocument(m),
	})

	switch {
	case err == nil:
		created = true
		return

	case errors.Is(err, ErrUniqueViolation):
		// If metadata were created by another transaction we consider it transaction conflict error
		// to mark that transaction should be retried.
		tableName = ""
		err = lazyerrors.Error(newTransactionConflictError(err))

		return

	default:
		tableName = ""
		err = lazyerrors.Error(err)

		return
	}
}

// getTableName returns PostgreSQL table name for the given FerretDB database and collection.
//
// If such metadata don't exist, it returns ErrTableNotExist.
func (ms *metadataStorage) getTableName(ctx context.Context) (string, error) {
	metadata, err := ms.get(ctx, false)
	if err != nil {
		return "", lazyerrors.Error(err)
	}

	return metadata.table, nil
}

// renameCollection renames metadataStorage.collection.
func (ms *metadataStorage) renameCollection(ctx context.Context, to string) error {
	metadata, err := ms.get(ctx, true)
	if err != nil {
		return lazyerrors.Error(err)
	}

	metadata.collection = to

	return ms.set(ctx, metadata)
}

// get returns metadata stored in the metadata table.
//
// If such metadata don't exist, it returns ErrTableNotExist.
func (ms *metadataStorage) get(ctx context.Context, forUpdate bool) (*metadata, error) {
	metadataExist, err := tableExists(ctx, ms.tx, ms.db, dbMetadataTableName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if !metadataExist {
		return nil, ErrTableNotExist
	}

	iterParams := &iteratorParams{
		schema:    ms.db,
		table:     dbMetadataTableName,
		filter:    must.NotFail(types.NewDocument("_id", ms.collection)),
		forUpdate: forUpdate,
	}

	iter, err := buildIterator(ctx, ms.tx, iterParams)
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

		idx, err := documentToMetadataIndex(idxDoc)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		indexes[i] = *idx
	}

	return &metadata{
		collection: must.NotFail(doc.Get("_id")).(string),
		table:      must.NotFail(doc.Get("table")).(string),
		indexes:    indexes,
	}, nil
}

// documentToMetadataIndex converts *types.Document to metadataIndex.
func documentToMetadataIndex(doc *types.Document) (*metadataIndex, error) {
	keyDoc := must.NotFail(doc.Get("key")).(*types.Document)
	key := make(IndexKey, keyDoc.Len())

	keyIter := keyDoc.Iterator()
	defer keyIter.Close()

	for i := 0; i < keyDoc.Len(); i++ {
		field, value, err := keyIter.Next()

		switch {
		case err == nil:
			key[i] = IndexKeyPair{
				Field: field,
				Order: types.SortType(value.(int32)),
			}
		default:
			return nil, lazyerrors.Error(err)
		}
	}

	return &metadataIndex{
		Index: Index{
			Name:   must.NotFail(doc.Get("name")).(string),
			Key:    key,
			Unique: must.NotFail(doc.Get("unique")).(bool),
		},
		pgIndex: must.NotFail(doc.Get("pgindex")).(string),
	}, nil
}

// set sets metadata for the given database and collection.
//
// To avoid data race, set should be called only after getMetadata with forUpdate = true is called,
// so that the metadata table is locked correctly.
func (ms *metadataStorage) set(ctx context.Context, metadata *metadata) error {
	doc := metadataToDocument(metadata)

	if _, err := setById(ctx, ms.tx, ms.db, dbMetadataTableName, "", ms.collection, doc); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// metadataToDocument converts metadata to *types.Document.
// Use this function to transform metadata to document to be stored in the database.
func metadataToDocument(metadata *metadata) *types.Document {
	indexesArr := types.MakeArray(len(metadata.indexes))

	for _, idx := range metadata.indexes {
		keyDoc := types.MakeDocument(len(idx.Key))
		for _, pair := range idx.Key {
			keyDoc.Set(pair.Field, int32(pair.Order)) // order is set as int32 to be pjson-marshaled correctly
		}

		indexesArr.Append(must.NotFail(types.NewDocument(
			"pgindex", idx.pgIndex,
			"name", idx.Name,
			"key", keyDoc,
			"unique", idx.Unique,
		)))
	}

	return must.NotFail(types.NewDocument(
		"_id", metadata.collection,
		"table", metadata.table,
		"indexes", indexesArr,
	))
}

// remove removes metadata.
//
// If such metadata don't exist, it doesn't return an error.
func (ms *metadataStorage) remove(ctx context.Context) error {
	_, err := deleteByIDs(ctx, ms.tx, execDeleteParams{
		schema: ms.db,
		table:  dbMetadataTableName,
	}, []any{ms.collection},
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
// If the given index already exists, it doesn't return an error.
//
// Indexes are stored in the `indexes` array of metadata entry.
//
// It returns a possibly wrapped error:
//   - ErrTableNotExist - if the metadata table doesn't exist.
//   - ErrIndexKeyAlreadyExist - if the given index key already exists.
//   - ErrIndexNameAlreadyExist - if the given index name already exists.
func (ms *metadataStorage) setIndex(ctx context.Context, index string, key IndexKey, unique bool) (pgTable string, pgIndex string, err error) { //nolint:lll // for readability
	metadata, err := ms.get(ctx, true)
	if err != nil {
		return
	}

	pgTable = metadata.table
	pgIndex = indexNameToPgIndexName(ms.collection, index)

	newIndex := metadataIndex{
		Index: Index{
			Name:   index,
			Key:    key,
			Unique: unique,
		},
		pgIndex: pgIndex,
	}

	// If index name or index key already exists, don't create it.
	// If existing name and key are equal to the given ones, don't return an error.
	// Otherwise, return an error.
	for _, existing := range metadata.indexes {
		var exists bool
		exists, err = checkExistingIndex(&existing, &newIndex)

		if err != nil {
			pgTable = ""
			pgIndex = ""

			return
		}

		if exists {
			return
		}
	}

	metadata.indexes = append(metadata.indexes, newIndex)

	if err = ms.set(ctx, metadata); err != nil {
		err = lazyerrors.Error(err)
		return
	}

	return
}

// checkExistingIndex checks if the given index already exists.
//
// It returns true and no error if existing and new index names and keys are equal.
// It returns false and no error if existing and new index names and keys are different.
// It returns false and an error if either name or key is equal (but not both).
// Possible errors:
//   - ErrIndexNameAlreadyExist - if index name already exists for a different key.
//   - ErrIndexKeyAlreadyExist - if index key already exists for a different name.
func checkExistingIndex(existing *metadataIndex, new *metadataIndex) (bool, error) {
	var indexNameMatch bool

	if existing.Name == new.Name {
		indexNameMatch = true
	}

	if len(existing.Key) != len(new.Key) {
		if indexNameMatch {
			return false, ErrIndexNameAlreadyExist
		}

		return false, nil
	}

	for i := range existing.Key {
		if existing.Key[i] != new.Key[i] {
			if indexNameMatch {
				return false, ErrIndexNameAlreadyExist
			}

			return false, nil
		}
	}

	// If we reached this line, the keys are equal.
	if indexNameMatch {
		return true, nil
	}

	return false, ErrIndexKeyAlreadyExist
}

// indexNameToPgIndexName returns index name in form <shortened_name>_<name_hash>_idx.
//
// Deprecated: this function usage is allowed for index metadata creation only.
func indexNameToPgIndexName(collection, index string) string {
	name := collection + "_" + index

	hash32 := fnv.New32a()
	must.NotFail(hash32.Write([]byte(name)))

	mangled := specialCharacters.ReplaceAllString(name, "_")

	nameSymbolsLeft := maxIndexNameLength - hash32.Size()*2 - 5 // 5 is for "_" delimiter and "_idx" suffix
	truncateTo := len(mangled)

	if truncateTo > nameSymbolsLeft {
		truncateTo = nameSymbolsLeft
	}

	return mangled[:truncateTo] + "_" + fmt.Sprintf("%08x", hash32.Sum(nil)) + "_idx"
}
