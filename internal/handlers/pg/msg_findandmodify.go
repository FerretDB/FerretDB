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

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgFindAndModify inserts, updates, or deletes, and returns a document matched by the query.
func (h *Handler) MsgFindAndModify(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/164

	unimplementedFields := []string{
		"sort",
		"update",
		"arrayFilters",
		"let",
	}
	if err := common.Unimplemented(document, unimplementedFields...); err != nil {
		return nil, err
	}

	for _, field := range []string{"new", "upsert"} {
		if err := common.UnimplementedNonDefault(document, field, func(v any) bool {
			b, ok := v.(bool)
			return ok && !b
		}); err != nil {
			return nil, err
		}
	}

	ignoredFields := []string{
		"fields",
		"bypassDocumentValidation",
		"writeConcern",
		"maxTimeMS",
		"collation",
		"hint",
		"comment",
	}
	common.Ignored(document, h.l, ignoredFields...)

	var query *types.Document
	var remove bool
	if query, err = common.GetOptionalParam(document, "query", query); err != nil {
		return nil, err
	}
	if remove, err = common.GetOptionalParam(document, "remove", remove); err != nil {
		return nil, err
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
		matches, err := common.FilterDocument(doc, query)
		if err != nil {
			return nil, err
		}

		if !matches {
			continue
		}

		resDocs = append(resDocs, doc)
	}

	// findAndModify always works with a single document
	if resDocs, err = common.LimitDocuments(resDocs, 1); err != nil {
		return nil, err
	}

	if len(resDocs) == 1 && remove {
		id := must.NotFail(fjson.Marshal(must.NotFail(resDocs[0].Get("_id"))))
		sql := fmt.Sprintf("DELETE FROM %s WHERE _jsonb->'_id' IN ($1)", pgx.Identifier{sp.db, sp.collection}.Sanitize())
		if _, err := h.pgPool.Exec(ctx, sql, id); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
