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

package sqlitedb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/sqlite/sjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func InsertDocument(ctx context.Context, db, collection string, doc *types.Document) error {
	if err := doc.ValidateData(); err != nil {
		return err
	}

	var err error
	var dbConn *sql.DB

	dbConn, err = CreateCollectionIfNotExists(db, collection)
	if err != nil {
		return lazyerrors.Error(err)
	}
	defer dbConn.Close()

	p := &insertParams{
		schema: db,
		table:  collection,
		doc:    doc,
	}
	err = insert(ctx, dbConn, p)
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

func insert(ctx context.Context, db *sql.DB, p *insertParams) error {
	sql := fmt.Sprintf(`INSERT INTO %s (json) VALUES (?)`, p.table)

	marshalled := must.NotFail(sjson.Marshal(p.doc))

	_, err := db.ExecContext(ctx, sql, marshalled)
	if err == nil {
		return nil
	}

	return lazyerrors.Error(err)
}
