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

	params, err := prepareFindAndModifyParams(document)
	if err != nil {
		return nil, err
	}

	fetchedDocs, err := h.fetch(ctx, params.sqlParam)
	if err != nil {
		return nil, err
	}

	err = common.SortDocuments(fetchedDocs, params.sort)
	if err != nil {
		return nil, err
	}

	resDocs := make([]*types.Document, 0, 16)
	for _, doc := range fetchedDocs {
		matches, err := common.FilterDocument(doc, params.query)
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

	if len(resDocs) == 0 {
		var reply wire.OpMsg
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"lastErrorObject", must.NotFail(types.NewDocument("n", int32(0), "updatedExisting", false)),
				"ok", float64(1),
			))},
		}))

		return &reply, nil
	}

	if len(resDocs) == 1 && params.remove {
		_, err = h.delete(ctx, params.sqlParam, resDocs)
		if err != nil {
			return nil, err
		}

		var reply wire.OpMsg
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"lastErrorObject", must.NotFail(types.NewDocument("n", int32(1))),
				"value", resDocs[0],
				"ok", float64(1),
			))},
		}))
		return &reply, nil
	}

	var hasUpdateOperators bool
	for k := range params.update.Map() {
		if _, ok := updateOperators[k]; ok {
			hasUpdateOperators = true
		}
	}

	if hasUpdateOperators { //nolint:nestif // no way to make it simple now
		upsert := resDocs[0].DeepCopy()
		err = common.UpdateDocument(upsert, params.update)
		if err != nil {
			return nil, err
		}

		_, err = h.update(ctx, params.sqlParam, upsert)
		if err != nil {
			return nil, err
		}
	} else {
		_, err = h.delete(ctx, params.sqlParam, resDocs)
		if err != nil {
			return nil, err
		}

		err = h.insert(ctx, params.sqlParam, params.update)
		if err != nil {
			return nil, err
		}
	}

	var reply wire.OpMsg
	resultDoc := resDocs[0]
	if params.returnNewDocument {
		resultDoc = params.update
	}
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"lastErrorObject", must.NotFail(types.NewDocument("n", int32(1), "updatedExisting", true)),
			"value", resultDoc,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// findAndModifyParams represent all findAndModify requests' fields.
// It's filled by calling prepareFindAndModifyParams.
type findAndModifyParams struct {
	sqlParam                          sqlParam
	query, sort, update               *types.Document
	remove, upsert, returnNewDocument bool
}

// prepareFindAndModifyParams prepares findAndModify request fields.
func prepareFindAndModifyParams(document *types.Document) (*findAndModifyParams, error) {
	var err error

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
		sqlParam: sqlParam{
			db:         db,
			collection: collection,
		},
		query:             query,
		update:            update,
		sort:              sort,
		remove:            remove,
		upsert:            upsert,
		returnNewDocument: returnNewDocument,
	}, nil
}

var updateOperators = map[string]struct{}{}

func init() {
	for _, o := range []string{"$currentDate", "$inc", "$min", "$max", "$mul", "$rename", "$set", "$setOnInsert", "$unset"} {
		updateOperators[o] = struct{}{}
	}
}
