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

package pg

import (
	"context"
	"io"
	"strings"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// sqlParam represents options/parameters used for sql query.
type sqlParam struct {
	db         string
	collection string
	comment    string
}

// fetch fetches all documents from the given database and collection.
// If collection doesn't exist it returns an empty slice and no error.
//
// TODO https://github.com/FerretDB/FerretDB/issues/372
func (h *Handler) fetch(ctx context.Context, param sqlParam) ([]*types.Document, error) {
	sql := `SELECT `
	if param.comment != "" {
		param.comment = strings.ReplaceAll(param.comment, "/*", "/ *")
		param.comment = strings.ReplaceAll(param.comment, "*/", "* /")

		sql += `/* ` + param.comment + ` */ `
	}
	sql += `_jsonb FROM ` + pgx.Identifier{param.db, param.collection}.Sanitize()

	// Special case: check if collection exists at all
	collectionExists, err := h.pgPool.TableExists(ctx, param.db, param.collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	if !collectionExists {
		h.l.Info(
			"Table doesn't exist, handling a case to deal with a non-existing collection.",
			zap.String("schema", param.db), zap.String("table", param.collection),
		)
		return []*types.Document{}, nil
	}

	rows, err := h.pgPool.Query(ctx, sql)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	var res []*types.Document
	for {
		doc, err := nextRow(rows)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		res = append(res, doc)
	}

	return res, nil
}

// nextRow returns the next document from the given rows.
func nextRow(rows pgx.Rows) (*types.Document, error) {
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, lazyerrors.Error(err)
		}
		return nil, io.EOF
	}

	var b []byte
	if err := rows.Scan(&b); err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := fjson.Unmarshal(b)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc.(*types.Document), nil
}
