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

package jsonb1

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDelete deletes document.
func (s *storage) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := common.Unimplemented(document, "let"); err != nil {
		return nil, err
	}
	common.Ignored(document, s.l, "ordered", "writeConcern")

	command := document.Command()

	var db, collection string
	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}
	if collection, err = common.GetRequiredParam[string](document, command); err != nil {
		return nil, err
	}

	var deletes *types.Array
	if deletes, err = common.GetOptionalParam(document, "deletes", deletes); err != nil {
		return nil, err
	}

	var deleted int32
	for i := 0; i < deletes.Len(); i++ {
		d, err := common.AssertType[*types.Document](must.NotFail(deletes.Get(i)))
		if err != nil {
			return nil, err
		}

		if err := common.Unimplemented(d, "collation", "hint", "comment"); err != nil {
			return nil, err
		}

		var filter *types.Document
		var limit int32
		if filter, err = common.GetOptionalParam(d, "q", filter); err != nil {
			return nil, err
		}
		if limit, err = common.GetOptionalParam(d, "limit", limit); err != nil {
			return nil, err
		}

		fetchedDocs, err := s.fetch(ctx, db, collection)
		if err != nil {
			return nil, err
		}

		resDocs := make([]*types.Document, 0, 16)
		for _, doc := range fetchedDocs {
			matches, err := common.FilterDocument(doc, filter)
			if err != nil {
				return nil, err
			}

			if !matches {
				continue
			}

			resDocs = append(resDocs, doc)
		}

		if resDocs, err = common.LimitDocuments(resDocs, limit); err != nil {
			return nil, err
		}

		if len(resDocs) == 0 {
			continue
		}

		var p pg.Placeholder
		placeholders := make([]string, len(resDocs))
		ids := make([]any, len(resDocs))
		for i, doc := range resDocs {
			placeholders[i] = p.Next()
			id := must.NotFail(doc.Get("_id"))
			ids[i] = must.NotFail(fjson.Marshal(id))
		}

		sql := fmt.Sprintf(
			"DELETE FROM %s WHERE _jsonb->'_id' IN (%s)",
			pgx.Identifier{db, collection}.Sanitize(), strings.Join(placeholders, ", "),
		)
		tag, err := s.pgPool.Exec(ctx, sql, ids...)
		if err != nil {
			// TODO check error code
			return nil, common.NewError(common.ErrNamespaceNotFound, fmt.Errorf("delete: ns not found: %w", err))
		}

		deleted += int32(tag.RowsAffected())
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"n", deleted,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
