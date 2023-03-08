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

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/pg/pjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// InsertDocument inserts a document into FerretDB database and collection.
// If database or collection does not exist, it will be created.
//
// It returns possibly wrapped error:
//   - *types.ValidationError - if the document is not valid.
//   - ErrUniqueViolation - if pgerrcode.UniqueViolation error is caught (e.g. due to unique index constraint).
//   - ErrInvalidCollectionName - if the given collection name doesn't conform to restrictions.
//   - ErrInvalidDatabaseName - if the given database name doesn't conform to restrictions.
//   - *transactionConflictError - if a PostgreSQL conflict occurs (the caller could retry the transaction).
func InsertDocument(ctx context.Context, tx pgx.Tx, db, collection string, doc *types.Document) error {
	if err := doc.ValidateData(); err != nil {
		return err
	}

	var err error

	err = CreateCollectionIfNotExists(ctx, tx, db, collection)
	if err != nil {
		return lazyerrors.Error(err)
	}

	var table string
	table, err = newMetadata(tx, db, collection).getTableName(ctx)

	if err != nil {
		return lazyerrors.Error(err)
	}

	p := insertParams{
		schema: db,
		table:  table,
		doc:    doc,
	}
	err = insert(ctx, tx, p)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// insertParams describes the parameters for inserting a document into a table.
type insertParams struct {
	doc    *types.Document // document to insert
	schema string          // pg schema name
	table  string          // pg table name
}

// insert marshals and inserts a document with the given params.
//
// It returns possibly wrapped error:
//   - ErrUniqueViolation - if the pgerrcode.UniqueViolation error is caught (e.g. due to unique index constraint).
//   - *transactionConflictError - if a PostgreSQL conflict occurs (the caller could retry the transaction).
func insert(ctx context.Context, tx pgx.Tx, p insertParams) error {
	sql := `INSERT INTO ` + pgx.Identifier{p.schema, p.table}.Sanitize() +
		` (_jsonb) VALUES ($1)`

	_, err := tx.Exec(ctx, sql, must.NotFail(pjson.Marshal(p.doc)))
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return lazyerrors.Error(err)
	}

	switch pgErr.Code {
	case pgerrcode.UniqueViolation:
		// unique violation due to index constraint or database conflict
		return ErrUniqueViolation
	case pgerrcode.DeadlockDetected:
		return newTransactionConflictError(err)
	default:
		return lazyerrors.Error(err)
	}
}
