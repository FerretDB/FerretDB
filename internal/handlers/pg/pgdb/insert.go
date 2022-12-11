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

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// InsertDocument inserts a document into FerretDB database and collection.
// If database or collection does not exist, it will be created.
// If the document is not valid, it returns *types.ValidationError.
func InsertDocument(ctx context.Context, pgPool *Pool, db, collection string, doc *types.Document) error {
	if err := doc.ValidateData(); err != nil {
		return err
	}

	var exists bool
	err := pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		var err error
		exists, err = CollectionExists(ctx, tx, db, collection)
		return err
	})
	if err != nil {
		return err
	}

	if !exists {
		err = pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
			if err = CreateDatabaseIfNotExists(ctx, tx, db); err != nil {
				return lazyerrors.Error(err)
			}
			return nil
		})
		if err != nil && !errors.Is(err, ErrAlreadyExist) {
			return err
		}

		err = pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
			if err = CreateCollection(ctx, tx, db, collection); err != nil {
				return lazyerrors.Error(err)
			}
			return nil
		})
		if err != nil && !errors.Is(err, ErrAlreadyExist) {
			return err
		}
	}

	var table string
	err = pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		table, err = getTableName(ctx, tx, db, collection)
		return err
	})
	if err != nil {
		return lazyerrors.Error(err)
	}

	sql := `INSERT INTO ` + pgx.Identifier{db, table}.Sanitize() +
		` (_jsonb) VALUES ($1)`

	err = pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		_, err = tx.Exec(ctx, sql, must.NotFail(pjson.MarshalWithSchema(doc)))
		return err
	})
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
