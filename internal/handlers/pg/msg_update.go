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

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgUpdate updates documents that are matched by the query.
//
//nolint:nestif // TODO refactor to simplify
func (h *Handler) MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := common.Unimplemented(document, "let"); err != nil {
		return nil, err
	}
	common.Ignored(document, h.l, "ordered", "writeConcern", "bypassDocumentValidation", "comment")

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

	var updates *types.Array
	if updates, err = common.GetOptionalParam(document, "updates", updates); err != nil {
		return nil, err
	}

	created, err := h.pgPool.EnsureTableExist(ctx, sp.db, sp.collection)
	if err != nil {
		return nil, err
	}
	if created {
		h.l.Info("Created table.", zap.String("schema", sp.db), zap.String("table", sp.collection))
	}

	var matched, modified int32
	var upserted types.Array
	for i := 0; i < updates.Len(); i++ {
		update, err := common.AssertType[*types.Document](must.NotFail(updates.Get(i)))
		if err != nil {
			return nil, err
		}

		unimplementedFields := []string{
			"c",
			"multi",
			"collation",
			"arrayFilters",
			"hint",
		}
		if err := common.Unimplemented(update, unimplementedFields...); err != nil {
			return nil, err
		}

		var q, u *types.Document
		var upsert bool
		if q, err = common.GetOptionalParam(update, "q", q); err != nil {
			return nil, err
		}
		if u, err = common.GetOptionalParam(update, "u", u); err != nil {
			return nil, err
		}
		if upsert, err = common.GetOptionalParam(update, "upsert", upsert); err != nil {
			return nil, err
		}

		fetchedDocs, err := h.fetch(ctx, sp)
		if err != nil {
			return nil, err
		}

		resDocs := make([]*types.Document, 0, 16)
		for _, doc := range fetchedDocs {
			matches, err := common.FilterDocument(doc, q)
			if err != nil {
				return nil, err
			}

			if !matches {
				continue
			}

			resDocs = append(resDocs, doc)
		}

		if len(resDocs) == 0 {
			if !upsert {
				// nothing to do, continue to the next update operation
				continue
			}

			doc := q.DeepCopy()
			if err = common.UpdateDocument(doc, u); err != nil {
				return nil, lazyerrors.Error(err)
			}
			if !doc.Has("_id") {
				must.NoError(doc.Set("_id", types.NewObjectID()))
			}

			must.NoError(upserted.Append(must.NotFail(types.NewDocument(
				"index", int32(0), // TODO
				"_id", must.NotFail(doc.Get("_id")),
			))))

			sql := fmt.Sprintf("INSERT INTO %s (_jsonb) VALUES ($1)", pgx.Identifier{sp.db, sp.collection}.Sanitize())
			b, err := fjson.Marshal(doc)
			if err != nil {
				return nil, err
			}

			if _, err := h.pgPool.Exec(ctx, sql, b); err != nil {
				return nil, err
			}

			matched += 1

			continue
		}

		matched += int32(len(resDocs))

		for _, doc := range resDocs {
			if err = common.UpdateDocument(doc, u); err != nil {
				return nil, lazyerrors.Error(err)
			}

			sql := "UPDATE " + pgx.Identifier{sp.db, sp.collection}.Sanitize() +
				" SET _jsonb = $1 WHERE _jsonb->'_id' = $2"
			id := must.NotFail(doc.Get("_id"))
			tag, err := h.pgPool.Exec(ctx, sql, must.NotFail(fjson.Marshal(doc)), must.NotFail(fjson.Marshal(id)))
			if err != nil {
				return nil, err
			}

			modified += int32(tag.RowsAffected())
		}
	}

	res := must.NotFail(types.NewDocument(
		"n", matched,
	))
	if upserted.Len() != 0 {
		must.NoError(res.Set("upserted", &upserted))
	}
	must.NoError(res.Set("nModified", modified))
	must.NoError(res.Set("ok", float64(1)))

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{res},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
