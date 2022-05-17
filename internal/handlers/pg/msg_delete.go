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
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDelete deletes documents matched by the query.
func (h *Handler) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := common.Unimplemented(document, "let"); err != nil {
		return nil, err
	}
	common.Ignored(document, h.l, "ordered", "writeConcern")

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
		if filter, err = common.GetOptionalParam(d, "q", filter); err != nil {
			return nil, err
		}

		var limit int64
		if l, _ := d.Get("limit"); l != nil {
			if limit, err = common.GetWholeNumberParam(l); err != nil {
				return nil, err
			}
		}

		var sp sqlParam
		if sp.db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
			return nil, err
		}
		collectionParam, err := document.Get(document.Command())
		if err != nil {
			return nil, err
		}
		var ok bool
		if sp.collection, ok = collectionParam.(string); !ok {
			return nil, common.NewErrorMsg(
				common.ErrBadValue,
				fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			)
		}

		fetchedDocs, err := h.fetch(ctx, sp)
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

		var p pgdb.Placeholder
		placeholders := make([]string, len(resDocs))
		ids := make([]any, len(resDocs))
		for i, doc := range resDocs {
			placeholders[i] = p.Next()
			id := must.NotFail(doc.Get("_id"))
			ids[i] = must.NotFail(fjson.Marshal(id))
		}

		sql := fmt.Sprintf(
			"DELETE FROM %s WHERE _jsonb->'_id' IN (%s)",
			pgx.Identifier{sp.db, sp.collection}.Sanitize(), strings.Join(placeholders, ", "),
		)
		tag, err := h.pgPool.Exec(ctx, sql, ids...)
		if err != nil {
			// TODO check error code
			return nil, common.NewError(common.ErrNamespaceNotFound, fmt.Errorf("delete: ns not found: %w", err))
		}

		deleted += int32(tag.RowsAffected())
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"n", deleted,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
