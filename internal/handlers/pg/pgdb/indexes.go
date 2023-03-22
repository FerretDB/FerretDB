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

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Index contains user-visible properties of FerretDB index.
type Index struct {
	Name   string   // FerretDB index name
	Key    IndexKey // Index specification (field name + sort order pairs)
	Unique bool     // Whether the index is unique
}

// IndexKey is a list of field name + sort order pairs.
type IndexKey []IndexKeyPair

// IndexKeyPair consists of a field name and a sort order that are part of the index.
type IndexKeyPair struct {
	Field string
	Order IndexOrder
}

// IndexOrder defines a type for index sort order.
type IndexOrder int8

// IndexOrder constants.
const (
	IndexOrderAsc  IndexOrder = 1
	IndexOrderDesc IndexOrder = -1
)

// Indexes returns a list of indexes for the given database and collection.
//
// If the given collection does not exist, it returns ErrTableNotExist.
func Indexes(ctx context.Context, tx pgx.Tx, db, collection string) ([]Index, error) {
	metadata, err := newMetadata(tx, db, collection).get(ctx, false)
	if err != nil {
		return nil, err
	}

	if !metadata.Has("indexes") {
		return []Index{}, nil
	}

	indexes := must.NotFail(metadata.Get("indexes")).(*types.Array)

	res := make([]Index, indexes.Len())
	iter := indexes.Iterator()

	defer iter.Close()

	for {
		i, idx, err := iter.Next()

		switch {
		case err == nil:
			idx := idx.(*types.Document)
			key := must.NotFail(idx.Get("key")).(*types.Document)

			res[i] = Index{
				Name:   must.NotFail(idx.Get("name")).(string),
				Unique: must.NotFail(idx.Get("unique")).(bool),
				Key:    make([]IndexKeyPair, 0, key.Len()),
			}

			keyIter := key.Iterator()
			defer keyIter.Close() // it's safe to defer here as we always read the whole iterator

			for i := 0; i < key.Len(); i++ {
				var field string
				var value any
				field, value, err = keyIter.Next()

				switch {
				case err == nil:
					res[i].Key = append(res[i].Key, IndexKeyPair{
						Field: field,
						Order: IndexOrder(value.(int32)),
					})
				default:
					return nil, lazyerrors.Error(err)
				}
			}

		case errors.Is(err, iterator.ErrIteratorDone):
			// no more indexes
			// TODO Check indexes order when more than one index exist https://github.com/FerretDB/FerretDB/issues/1509
			// slices.Sort(res)
			return res, nil
		default:
			return nil, lazyerrors.Error(err)
		}
	}
}

// CreateIndexIfNotExists creates a new index for the given params if such an index doesn't exist.
// If the index exists, it doesn't return an error.
//
// It returns a possibly wrapped error:
//   - ErrTableNotExist - if the given collection doesn't exist.
func CreateIndexIfNotExists(ctx context.Context, tx pgx.Tx, db, collection string, i *Index) error {
	pgTable, pgIndex, err := newMetadata(tx, db, collection).setIndex(ctx, i.Name, i.Key, i.Unique)
	if err != nil {
		return err
	}

	if err := createPGIndexIfNotExists(ctx, tx, db, pgTable, pgIndex, true); err != nil {
		return err
	}

	return nil
}

// createPGIndexIfNotExists creates a new index for the given params if it does not exist.
func createPGIndexIfNotExists(ctx context.Context, tx pgx.Tx, schema, table, index string, isUnique bool) error {
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
