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
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Index contains user-visible properties of FerretDB index.
type Index struct {
	Name   string
	Key    IndexKey
	Unique *bool // we have to use pointer to determine whether the field was set or not
}

// IndexKey is a list of field name + sort order pairs.
type IndexKey []IndexKeyPair

// IndexKeyPair consists of a field name and a sort order that are part of the index.
type IndexKeyPair struct {
	Field string
	Order types.SortType
}

// Indexes returns a list of indexes for the given database and collection.
//
// If the given collection does not exist, it returns ErrTableNotExist.
func Indexes(ctx context.Context, tx pgx.Tx, db, collection string) ([]Index, error) {
	metadata, err := newMetadataStorage(tx, db, collection).get(ctx, false)
	if err != nil {
		return nil, err
	}

	res := make([]Index, len(metadata.indexes))

	for i, idx := range metadata.indexes {
		res[i] = idx.Index
	}

	return res, nil
}

// CreateIndexIfNotExists creates a new index for the given params if such an index doesn't exist.
//
// If index creation also caused the collection to be created, it returns true as the first return value.
//
// If the index exists, it doesn't return an error.
func CreateIndexIfNotExists(ctx context.Context, tx pgx.Tx, db, collection string, i *Index) (bool, error) {
	var collCreated bool
	var err error

	if collCreated, err = CreateCollectionIfNotExists(ctx, tx, db, collection); err != nil {
		return false, err
	}

	pgTable, pgIndex, err := newMetadataStorage(tx, db, collection).setIndex(ctx, i.Name, i.Key, i.Unique)
	if err != nil {
		return false, err
	}

	var unique bool
	if i.Unique != nil {
		unique = *i.Unique
	}

	if err := createPgIndexIfNotExists(ctx, tx, db, pgTable, pgIndex, i.Key, unique); err != nil {
		return false, err
	}

	return collCreated, nil
}

// Equal returns true if the given index key is equal to the current one.
func (k IndexKey) Equal(v IndexKey) bool {
	if len(k) != len(v) {
		return false
	}

	for i := range k {
		if k[i] != v[i] {
			return false
		}
	}

	return true
}

// DropIndex drops index. If the index was not found, it returns error.
func DropIndex(ctx context.Context, tx pgx.Tx, db, collection string, index *Index) (int32, error) {
	ms := newMetadataStorage(tx, db, collection)

	metadata, err := ms.get(ctx, true)
	if err != nil {
		return 0, err
	}

	nIndexesWas := int32(len(metadata.indexes))

	for i := nIndexesWas - 1; i >= 0; i-- {
		current := metadata.indexes[i]

		var deleteCurrentIndex bool

		if index.Name != "" {
			// delete by name
			deleteCurrentIndex = current.Name == index.Name
		} else {
			// delete by key
			deleteCurrentIndex = current.Key.Equal(index.Key)
		}

		if !deleteCurrentIndex {
			continue
		}

		if current.Name == "_id_" {
			// cannot delete _id index
			return 0, ErrIndexCannotDelete
		}

		if err = dropPgIndex(ctx, tx, db, current.pgIndex); err != nil {
			return 0, lazyerrors.Error(err)
		}

		// remove i-th element from the slice
		metadata.indexes = append(metadata.indexes[:i], metadata.indexes[i+1:]...)

		if err := ms.set(ctx, metadata); err != nil {
			return 0, lazyerrors.Error(err)
		}

		return nIndexesWas, nil
	}

	// Did not find the index to delete
	return 0, ErrIndexNotExist
}

// DropAllIndexes deletes all indexes on the collection except _id index.
func DropAllIndexes(ctx context.Context, tx pgx.Tx, db, collection string) (int32, error) {
	ms := newMetadataStorage(tx, db, collection)

	metadata, err := ms.get(ctx, true)
	if err != nil {
		return 0, lazyerrors.Error(err)
	}

	nIndexesWas := int32(len(metadata.indexes))

	for i := nIndexesWas - 1; i >= 0; i-- {
		if metadata.indexes[i].Name == "_id_" {
			continue
		}

		if err = dropPgIndex(ctx, tx, db, metadata.indexes[i].pgIndex); err != nil {
			return 0, lazyerrors.Error(err)
		}

		// remove i-th element from the slice
		metadata.indexes = append(metadata.indexes[:i], metadata.indexes[i+1:]...)
	}

	if err := ms.set(ctx, metadata); err != nil {
		return 0, lazyerrors.Error(err)
	}

	return nIndexesWas, nil
}

// createPgIndexIfNotExists creates a new index for the given params if it does not exist.
func createPgIndexIfNotExists(ctx context.Context, tx pgx.Tx, schema, table, index string, fields IndexKey, isUnique bool) error {
	if len(fields) == 0 {
		return lazyerrors.Errorf("no fields for index")
	}

	var err error

	unique := ""
	if isUnique {
		unique = " UNIQUE"
	}

	fieldsDef := make([]string, len(fields))

	for i, field := range fields {
		var order string

		switch field.Order {
		case types.Ascending:
			order = "ASC"
		case types.Descending:
			order = "DESC"
		default:
			return lazyerrors.Errorf("unknown sort order: %d", field.Order)
		}

		// if the key is foo.bar, then need to modify it to foo -> bar
		fs := strings.Split(field.Field, ".")
		transformedParts := make([]string, len(fs))

		for j, f := range fs {
			// It's important to sanitize field.Field data here, as it's a user-provided value.
			transformedParts[j] = quoteString(f)
		}
		fieldsDef[i] = fmt.Sprintf(`((_jsonb->%s)) %s`, strings.Join(transformedParts, " -> "), order)
	}

	sql := `CREATE` + unique + ` INDEX IF NOT EXISTS ` + pgx.Identifier{index}.Sanitize() +
		` ON ` + pgx.Identifier{schema, table}.Sanitize() + ` (` + strings.Join(fieldsDef, `, `) + `)`

	_, err = tx.Exec(ctx, sql)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return lazyerrors.Error(err)
	}

	switch pgErr.Code {
	case pgerrcode.UniqueViolation:
		return ErrUniqueViolation
	default:
		return lazyerrors.Error(err)
	}
}

// dropPgIndex drops the given index.
func dropPgIndex(ctx context.Context, tx pgx.Tx, schema, index string) error {
	var err error

	sql := `DROP INDEX ` + pgx.Identifier{schema, index}.Sanitize()

	if _, err = tx.Exec(ctx, sql); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// quoteString returns a string that is safe to use in SQL queries.
//
// Deprecated: Warning! Avoid using this function unless there is no other way.
// Ideally, use a placeholder and pass the value as a parameter instead of calling this function.
//
// This approach is used in github.com/jackc/pgx/v4@v4.18.1/internal/sanitize/sanitize.go.
func quoteString(str string) string {
	// We need "standard_conforming_strings=on" and "client_encoding=UTF8" (checked in checkConnection),
	// otherwise we can't sanitize safely: https://github.com/jackc/pgx/issues/868#issuecomment-725544647
	return "'" + strings.ReplaceAll(str, "'", "''") + "'"
}
