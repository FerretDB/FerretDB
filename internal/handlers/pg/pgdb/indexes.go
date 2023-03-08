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

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// indexParams contains parameters for creating an index.
// TODO This type will become exported in https://github.com/FerretDB/FerretDB/issues/1509 (similar to QueryParams).
type indexParams struct {
	db         string   // FerretDB database name
	collection string   // FerretDB collection name
	index      string   // FerretDB index name
	key        indexKey // Index specification (pairs of field names and sort orders) // TODO
	unique     bool     // Whether the index is unique
}

// indexKey defines a type for index key - pairs of field names and sort orders.
type indexKey []indexKeyPair

// indexKeyPair consists of a field name and a sort order that are part of the index.
type indexKeyPair struct {
	field string
	order indexOrder
}

// indexOrder defines a type for index sort order.
type indexOrder int8

// indexOrder constants.
const (
	indexOrderAsc  indexOrder = 1
	indexOrderDesc indexOrder = -1
)

// createIndex creates a new index for the given params.
// TODO This method will become exported in https://github.com/FerretDB/FerretDB/issues/1509.
func createIndex(ctx context.Context, tx pgx.Tx, ip *indexParams) error {
	pgTable, pgIndex, err := newMetadata(tx, ip.db, ip.collection).setIndex(ctx, ip.index, ip.key, ip.unique)
	if err != nil {
		return err
	}

	if err := createPGIndexIfNotExists(ctx, tx, ip.db, pgTable, pgIndex, true); err != nil {
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
