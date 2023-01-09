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

package tigris

import (
	"context"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgFindAndModify implements HandlerInterface.
func (h *Handler) MsgFindAndModify(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	unimplementedFields := []string{
		"arrayFilters",
		"let",
		"fields",
	}
	if err := common.Unimplemented(document, unimplementedFields...); err != nil {
		return nil, err
	}

	ignoredFields := []string{
		"bypassDocumentValidation",
		"writeConcern",
		"collation",
		"hint",
	}
	common.Ignored(document, h.L, ignoredFields...)

	params, err := common.PrepareFindAndModifyParams(document)
	if err != nil {
		return nil, err
	}

	if params.MaxTimeMS != 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
		defer cancel()

		ctx = ctxWithTimeout
	}

	fp := tigrisdb.FetchParam{
		DB:         params.DB,
		Collection: params.Collection,
		Filter:     params.Query,
	}

	resDocs, err := h.fetchAndFilterDocs(ctx, &fp)
	if err != nil {
		return nil, err
	}

	err = common.SortDocuments(resDocs, params.Sort)
	if err != nil {
		return nil, err
	}

	// findAndModify always works with a single document
	if resDocs, err = common.LimitDocuments(resDocs, 1); err != nil {
		return nil, err
	}

	if params.Update != nil { // we have update part
		var upsert *types.Document
		var upserted bool

		if params.Upsert { //  we have upsert flag
			p := &upsertParams{
				hasUpdateOperators: params.HasUpdateOperators,
				query:              params.Query,
				update:             params.Update,
				fetchParam:         &fp,
			}

			upsert, upserted, err = h.upsert(ctx, resDocs, p)
			if err != nil {
				return nil, err
			}
		} else { // process update as usual
			if len(resDocs) == 0 {
				var reply wire.OpMsg
				must.NoError(reply.SetSections(wire.OpMsgSection{
					Documents: []*types.Document{must.NotFail(types.NewDocument(
						"lastErrorObject", must.NotFail(types.NewDocument("n", int32(0), "updatedExisting", false)),
						"value", types.Null,
						"ok", float64(1),
					))},
				}))

				return &reply, nil
			}

			if params.HasUpdateOperators {
				upsert = resDocs[0].DeepCopy()
				_, err = common.UpdateDocument(upsert, params.Update)
				if err != nil {
					return nil, err
				}

				_, err = h.update(ctx, &fp, upsert)
				if err != nil {
					return nil, err
				}
			} else {
				upsert = params.Update

				if !upsert.Has("_id") {
					upsert.Set("_id", must.NotFail(resDocs[0].Get("_id")))
				}

				_, err = h.update(ctx, &fp, upsert)
				if err != nil {
					return nil, err
				}
			}
		}

		var resultDoc *types.Document
		if params.ReturnNewDocument || len(resDocs) == 0 {
			resultDoc = upsert
		} else {
			resultDoc = resDocs[0]
		}

		lastErrorObject := must.NotFail(types.NewDocument(
			"n", int32(1),
			"updatedExisting", len(resDocs) > 0,
		))

		if upserted {
			lastErrorObject.Set("upserted", must.NotFail(resultDoc.Get("_id")))
		}

		var reply wire.OpMsg
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"lastErrorObject", lastErrorObject,
				"value", resultDoc,
				"ok", float64(1),
			))},
		}))

		return &reply, nil
	}

	if params.Remove {
		if len(resDocs) == 0 {
			var reply wire.OpMsg
			must.NoError(reply.SetSections(wire.OpMsgSection{
				Documents: []*types.Document{must.NotFail(types.NewDocument(
					"lastErrorObject", must.NotFail(types.NewDocument("n", int32(0))),
					"ok", float64(1),
				))},
			}))

			return &reply, nil
		}

		_, err = h.delete(ctx, &fp, resDocs)
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

	return nil, lazyerrors.New("bad flags combination")
}

// upsertParams represent parameters for Handler.upsert method.
type upsertParams struct {
	hasUpdateOperators bool
	query, update      *types.Document
	fetchParam         *tigrisdb.FetchParam
}

// upsert inserts new document if no documents in query result or updates given document.
// When inserting new document we must check that `_id` is present, so we must extract `_id` from query or generate a new one.
func (h *Handler) upsert(ctx context.Context, docs []*types.Document, params *upsertParams) (*types.Document, bool, error) {
	if len(docs) == 0 {
		upsert := must.NotFail(types.NewDocument())

		if params.hasUpdateOperators {
			_, err := common.UpdateDocument(upsert, params.update)
			if err != nil {
				return nil, false, err
			}
		} else {
			upsert = params.update
		}

		if !upsert.Has("_id") {
			if params.query.Has("_id") {
				upsert.Set("_id", must.NotFail(params.query.Get("_id")))
			} else {
				upsert.Set("_id", types.NewObjectID())
			}
		}

		err := h.insert(ctx, params.fetchParam, upsert)
		if err != nil {
			return nil, false, err
		}

		return upsert, true, nil
	}

	upsert := docs[0].DeepCopy()

	if params.hasUpdateOperators {
		_, err := common.UpdateDocument(upsert, params.update)
		if err != nil {
			return nil, false, err
		}
	} else {
		for _, k := range params.update.Keys() {
			upsert.Set(k, must.NotFail(params.update.Get(k)))
		}
	}

	_, err := h.update(ctx, params.fetchParam, upsert)
	if err != nil {
		return nil, false, err
	}

	return upsert, false, nil
}
