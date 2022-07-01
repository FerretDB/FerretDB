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

type fetched struct {
	docs []*types.Document
	err  error
}

// fetch fetches all documents from the given database and collection
// and sends them to the returned channel.
//
// The returned channel is always non-nil.
// The channel is closed when all documents are sent; the caller should always drain the channel.
// The returned error is the reason why the channel was closed:
// it is nil in case all documents were sent normally (or requested collection doesn't exist),
// error encountered during query initialization.
//
//
// If the collection doesn't exist, fetch returns a closed channel and no error.
func (h *Handler) fetch(ctx context.Context, param sqlParam) (<-chan fetched, error) {
	cdocs := make(chan fetched, 3)

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
		close(cdocs)
		return cdocs, nil
	}

	rows, err := h.pgPool.Query(ctx, sql)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	go func() {
		defer close(cdocs)
		defer rows.Close()

		for {
			select {
			case <-ctx.Done():
				h.l.Info("got a signal to stop fetching, fetch canceled",
					zap.String("schema", param.db), zap.String("table", param.collection),
				)
				return
			default:
				// fetch next batch of documents
			}

			res := make([]*types.Document, 2)
			for i := 0; i < len(res); i++ {
				doc, err := nextRow(rows)
				if err == io.EOF {
					break
				}
				if err != nil {
					cdocs <- fetched{err: err}
					return
				}
				res[i] = doc
			}

			cdocs <- fetched{docs: res}
		}
	}()

	return cdocs, nil
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
