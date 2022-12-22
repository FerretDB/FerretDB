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
func InsertDocument(ctx context.Context, tx pgx.Tx, db, collection string, doc *types.Document) error {
	if err := doc.ValidateData(); err != nil {
		return err
	}

	var err error

	err = CreateCollectionIfNotExist(ctx, tx, db, collection)
	if err != nil && !errors.Is(err, ErrAlreadyExist) {
		return err
	}

	var table string
	table, err = getSettings(ctx, tx, db, collection)
	if err != nil {
		return lazyerrors.Error(err)
	}

	p := insertParams{
		schema:         db,
		table:          table,
		doc:            doc,
		ignoreConflict: false,
	}
	err = insert(ctx, tx, p)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// insertParams describes the parameters for inserting a document into a table.
type insertParams struct {
	schema         string          // pg schema name
	table          string          // pg table name
	doc            *types.Document // document to insert
	ignoreConflict bool            // on conflict do nothing
}

// insert marshals and inserts a document with the given params.
func insert(ctx context.Context, tx pgx.Tx, p insertParams) error {
	sql := `INSERT INTO ` + pgx.Identifier{p.schema, p.table}.Sanitize() +
		` (_jsonb) VALUES ($1)`

	if p.ignoreConflict {
		sql += ` ON CONFLICT DO NOTHING`
	}

	_, err := tx.Exec(ctx, sql, must.NotFail(pjson.Marshal(p.doc)))

	return err
}
