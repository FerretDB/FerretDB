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

	unimplementedFields := []string{
		"arrayFilters",
		"let",
	}
	if err := common.Unimplemented(document, unimplementedFields...); err != nil {
		return nil, err
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

	p, err := prepareFindAndModifyParams(document, command)
	if err != nil {
		return nil, err
	}

	fetchedDocs, err := h.fetch(ctx, p.db, p.collection)
	if err != nil {
		return nil, err
	}

	err = common.SortDocuments(fetchedDocs, p.sort)
	if err != nil {
		return nil, err
	}

	resDocs := make([]*types.Document, 0, 16)
	for _, doc := range fetchedDocs {
		matches, err := common.FilterDocument(doc, p.query)
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

	if len(resDocs) == 1 && p.remove {
		_, err = h.delete(ctx, resDocs, p.db, p.collection)
		if err != nil {
			return nil, err
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

	if len(resDocs) == 1 && p.update != nil {
		return h.update(ctx, p.update, resDocs, p.db, p.collection, p.returnNewDocument)
	}
	if p.update != nil && p.upsert {
		return h.upsert(ctx, p.update, err, p.db, p.collection)
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"lastErrorObject", types.MustNewDocument("n", int32(0), "updatedExisting", false),
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

type findAndModifyParams struct {
	db, collection                    string
	query, sort, update               *types.Document
	remove, upsert, returnNewDocument bool
}

func prepareFindAndModifyParams(document *types.Document, command string) (*findAndModifyParams, error) {
	var err error
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

	var remove bool
	if remove, err = common.GetBoolOptionalParam(document, "remove"); err != nil {
		return nil, err
	}
	var returnNewDocument bool
	if returnNewDocument, err = common.GetBoolOptionalParam(document, "new"); err != nil {
		return nil, err
	}
	var upsert bool
	if upsert, err = common.GetBoolOptionalParam(document, "upsert"); err != nil {
		return nil, err
	}

	var query *types.Document
	if query, err = common.GetOptionalParam(document, "query", query); err != nil {
		return nil, err
	}

	var sort *types.Document
	if sort, err = common.GetOptionalParam(document, "sort", sort); err != nil {
		return nil, err
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
			return nil, common.NewErrorMsg(common.ErrFailedToParse, "Update argument must be either an object or an array")
		}
	}

	return &findAndModifyParams{
		db:                db,
		collection:        collection,
		query:             query,
		update:            update,
		sort:              sort,
		remove:            remove,
		upsert:            upsert,
		returnNewDocument: returnNewDocument,
	}, nil
}

func (h *Handler) upsert(ctx context.Context, update *types.Document, err error, db string, collection string) (*wire.OpMsg, error) {
	if common.HasUpdateOperator(update) {
		// TODO: skip upsert with update operators for now
		return nil, common.NewErrorMsg(common.ErrNotImplemented, "upsert with update operators not implemented")
	} else {
		err = h.insert(ctx, update, db, collection)
		if err != nil {
			return nil, err
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"lastErrorObject", types.MustNewDocument(
				"n", int32(1),
				"updatedExisting", false,
				"upserted", must.NotFail(update.Get("_id")),
			),
			"value", types.MustConvertDocument(update),
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

func (h *Handler) update(ctx context.Context, update *types.Document, resDocs []*types.Document, db string, collection string, returnNewDocument bool) (*wire.OpMsg, error) {
	if common.HasUpdateOperator(update) {
		err := common.UpdateDocument(resDocs[0], update)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := h.delete(ctx, resDocs, db, collection)
		if err != nil {
			return nil, err
		}

		err = h.insert(ctx, update, db, collection)
		if err != nil {
			return nil, err
		}
	}

	var reply wire.OpMsg
	resultDoc := resDocs[0]
	if returnNewDocument {
		resultDoc = update
	}
	err := reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"lastErrorObject", types.MustNewDocument("n", int32(1), "updatedExisting", true),
			"value", types.MustConvertDocument(resultDoc),
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
