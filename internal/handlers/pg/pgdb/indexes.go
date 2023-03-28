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
	"strings"

	"github.com/jackc/pgx/v4"

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
// If the index exists, it doesn't return an error.
// If the collection doesn't exist, it will be created and then the index will be created.
func CreateIndexIfNotExists(ctx context.Context, tx pgx.Tx, db, collection string, i *Index) error {
	if err := CreateCollectionIfNotExists(ctx, tx, db, collection); err != nil {
		return err
	}

	pgTable, pgIndex, err := newMetadataStorage(tx, db, collection).setIndex(ctx, i.Name, i.Key, i.Unique)
	if err != nil {
		return err
	}

	if err := createPgIndexIfNotExists(ctx, tx, db, pgTable, pgIndex, i.Key, true); err != nil {
		return err
	}

	return nil
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

		// It's important to sanitize field.Field data here, as it's a user-provided value.
		fieldsDef[i] = fmt.Sprintf(`((_jsonb->%s)) %s`, quoteString(field.Field), order)
	}

	tx.Conn()

	sql := `CREATE` + unique + ` INDEX IF NOT EXISTS ` + pgx.Identifier{index}.Sanitize() +
		` ON ` + pgx.Identifier{schema, table}.Sanitize() + ` (` + strings.Join(fieldsDef, `, `) + `)`

	if _, err = tx.Exec(ctx, sql); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// quoteString returns a string that is safe to use in SQL queries.
//
// Warning! Avoid using this function unless there is no other way.
// Ideally, use a placeholder and pass the value as a parameter instead of calling this function.
//
// This approach is used in github.com/jackc/pgx/v4@v4.18.1/internal/sanitize/sanitize.go.
func quoteString(str string) string {
	return "'" + strings.ReplaceAll(str, "'", "''") + "'"
}
