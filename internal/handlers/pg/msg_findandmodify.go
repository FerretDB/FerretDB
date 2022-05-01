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

// MsgFindAndModify inserts, updates, or deletes, and returns a document matched by the query.
func (h *Handler) MsgFindAndModify(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/164

	unimplementedFields := []string{
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

	command := document.Command()

	var db, collection string
	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}
	if collection, err = common.GetRequiredParam[string](document, command); err != nil {
		return nil, err
	}

	if collection == "" {
		return nil, common.NewErrorMsg(
			common.ErrInvalidNamespace,
			fmt.Sprintf("Invalid namespace specified '%s.'", db),
		)
	}

	var query *types.Document
	var remove bool
	if query, err = common.GetOptionalParam(document, "query", query); err != nil {
		return nil, err
	}
	if remove, err = common.GetOptionalParam(document, "remove", remove); err != nil {
		return nil, err
	}

	var sort *types.Document
	var ok bool
	sortParam, err := document.Get("sort")
	if err == nil {
		sort, ok = sortParam.(*types.Document)
		if !ok {
			return nil, common.NewErrorMsg(
				common.ErrTypeMismatch,
				fmt.Sprintf("BSON field 'findAndModify.sort' is the wrong type '%T', expected type 'object'", sortParam),
			)
		}
	}

	var update *types.Document
	updateParam, err := document.Get("update")
	if err != nil && !remove {
		return nil, common.NewErrorMsg(common.ErrFailedToParse, "Either an update or remove=true must be specified")
	}
	if err == nil {
		switch updateParam := updateParam.(type) {
		case *types.Document:
			update = updateParam
		case *types.Array:
			return nil, common.NewErrorMsg(common.ErrNotImplemented, "Aggregation pipelines are not supported yet")
		default:
			return nil, common.NewErrorMsg(common.ErrBadValue, "Bad update value")
		}
	}

	fetchedDocs, err := h.fetch(ctx, db, collection)
	if err != nil {
		return nil, err
	}

	err = common.SortDocuments(fetchedDocs, sort)
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
		sql := fmt.Sprintf("DELETE FROM %s WHERE _jsonb->'_id' IN ($1)", pgx.Identifier{db, collection}.Sanitize())
		if _, err := h.pgPool.Exec(ctx, sql, id); err != nil {
			return nil, lazyerrors.Error(err)
		}

		var reply wire.OpMsg
		err = reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{types.MustNewDocument(
				"lastErrorObject", types.MustNewDocument("n", int32(1)),
				"value", types.MustConvertDocument(resDocs[0]),
				"ok", float64(1),
			)},
		})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		return &reply, nil
	}

	if len(resDocs) == 1 && update != nil {
		// TODO: process update
		if common.HasUpdateOperator(update) {
			err := common.UpdateDocument(resDocs[0], update)
			if err != nil {
				return nil, err
			}
		} else {
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
				pgx.Identifier{db, collection}.Sanitize(), strings.Join(placeholders, ", "),
			)
			_, err := h.pgPool.Exec(ctx, sql, ids...)
			if err != nil {
				// TODO check error code
				return nil, common.NewError(common.ErrNamespaceNotFound, fmt.Errorf("delete: ns not found: %w", err))
			}

			sql = fmt.Sprintf("INSERT INTO %s (_jsonb) VALUES ($1)", pgx.Identifier{db, collection}.Sanitize())
			b, err := fjson.Marshal(update)
			if err != nil {
				return nil, err
			}

			if _, err = h.pgPool.Exec(ctx, sql, b); err != nil {
				return nil, err
			}
		}

		var reply wire.OpMsg
		err = reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{types.MustNewDocument(
				"lastErrorObject", types.MustNewDocument("n", int32(1), "updateExisting", true),
				"value", types.MustConvertDocument(resDocs[0]),
				"ok", float64(1),
			)},
		})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return &reply, nil
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
