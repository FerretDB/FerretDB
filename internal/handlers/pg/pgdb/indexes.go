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

	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Index contains user-visible properties of FerretDB index.
type Index struct {
	Name   string
	Key    IndexKey
	Unique bool
}

// IndexKey is a list of field name + sort order pairs.
type IndexKey []IndexKeyPair

// IndexKeyPair consists of a field name and a sort order that are part of the index.
type IndexKeyPair struct {
	Field string
	Order types.SortType
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

	// TODO Add tests that indexes sorted correctly: https://github.com/FerretDB/FerretDB/issues/1509
	slices.SortFunc(res, func(a, b Index) bool { return a.Name < b.Name })

	return res, nil
}

// DropIndex drops index. If the index was not found, it returns error.
func DropIndex(ctx context.Context, tx pgx.Tx, db, collection string, index *Index) error {
	ms := newMetadataStorage(tx, db, collection)

	metadata, err := ms.get(ctx, true)
	if err != nil {
		return err
	}

	for i := len(metadata.indexes) - 1; i >= 0; i-- {
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
			return ErrIndexCannotDelete
		}

		if err = dropPgIndex(ctx, tx, db, current.pgIndex); err != nil {
			return lazyerrors.Error(err)
		}

		// todo check this removed the index we want
		metadata.indexes = append(metadata.indexes[:i], metadata.indexes[i+1:]...)

		return ms.set(ctx, metadata)
	}

	// Did not find the index to delete
	return ErrIndexNotExist
}

// DropAllIndexes deletes all indexes on the collection except _id index.
func DropAllIndexes(ctx context.Context, tx pgx.Tx, db, collection string) error {
	ms := newMetadataStorage(tx, db, collection)

	metadata, err := ms.get(ctx, true)
	if err != nil {
		return err
	}

	for i := len(metadata.indexes) - 1; i >= 0; i-- {
		if metadata.indexes[i].Name == "_id_" {
			continue
		}

		if err = dropPgIndex(ctx, tx, db, metadata.indexes[i].pgIndex); err != nil {
			return lazyerrors.Error(err)
		}

		// todo check this removed the index we want
		metadata.indexes = append(metadata.indexes[:i], metadata.indexes[i+1:]...)
	}

	return ms.set(ctx, metadata)
}

// createIndex creates a new index for the given params.
// TODO This method will become exported in https://github.com/FerretDB/FerretDB/issues/1509.
func createIndex(ctx context.Context, tx pgx.Tx, db, collection string, i *Index) error {
	pgTable, pgIndex, err := newMetadataStorage(tx, db, collection).setIndex(ctx, i.Name, i.Key, i.Unique)
	if err != nil {
		return err
	}

	if err := createPgIndexIfNotExists(ctx, tx, db, pgTable, pgIndex, true); err != nil {
		return err
	}

	return nil
}

// createPgIndexIfNotExists creates a new index for the given params if it does not exist.
func createPgIndexIfNotExists(ctx context.Context, tx pgx.Tx, schema, table, index string, isUnique bool) error {
	var err error

	unique := ""
	if isUnique {
		unique = " UNIQUE"
	}

	sql := `CREATE` + unique + ` INDEX IF NOT EXISTS ` + pgx.Identifier{index}.Sanitize() +
		` ON ` + pgx.Identifier{schema, table}.Sanitize() +
		` ((_jsonb->'_id'))` // TODO Provide ability to set fields https://github.com/FerretDB/FerretDB/issues/1509

	if _, err = tx.Exec(ctx, sql); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
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
