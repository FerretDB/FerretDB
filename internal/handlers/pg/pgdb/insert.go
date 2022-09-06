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
	"strconv"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// InsertDocument inserts a document into FerretDB database and collection.
// If database or collection does not exist, it will be created.
func InsertDocument(ctx context.Context, querier pgxtype.Querier, db, collection string, docs []*types.Document) error {
	exists, err := CollectionExists(ctx, querier, db, collection)
	if err != nil {
		return err
	}

	if !exists {
		if err := CreateDatabaseIfNotExists(ctx, querier, db); err != nil {
			return lazyerrors.Error(err)
		}

		if err := CreateCollection(ctx, querier, db, collection); err != nil && !errors.Is(err, ErrAlreadyExist) {
			return lazyerrors.Error(err)
		}
	}

	table, err := getTableName(ctx, querier, db, collection)
	if err != nil {
		return lazyerrors.Error(err)
	}

	sql := `INSERT INTO ` + pgx.Identifier{db, table}.Sanitize() +
		` (_jsonb) VALUES `
	args := make([]any, len(docs))
	for i, doc := range docs {
		if i > 0 {
			sql += ","
		}
		sql += "($" + strconv.Itoa(i+1) + ")"
		args[i] = must.NotFail(fjson.Marshal(doc))
	}
	sql += `;`

	if _, err = querier.Exec(ctx, sql, args...); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
