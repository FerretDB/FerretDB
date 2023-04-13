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
	"errors"
	"fmt"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgFindAndModify implements HandlerInterface.
func (h *Handler) MsgFindAndModify(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetFindAndModifyParams(document, h.L)
	if err != nil {
		return nil, err
	}

	if params.MaxTimeMS != 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
		defer cancel()

		ctx = ctxWithTimeout
	}

	qp := tigrisdb.QueryParams{
		DB:         params.DB,
		Collection: params.Collection,
		Filter:     params.Query,
	}

	resDocs, err := fetchAndFilterDocs(ctx, &fetchParams{dbPool, &qp, h.DisablePushdown})
	if err != nil {
		return nil, err
	}

	if err = common.SortDocuments(resDocs, params.Sort); err != nil {
		var pathErr *types.DocumentPathError
		if errors.As(err, &pathErr) && pathErr.Code() == types.ErrDocumentPathEmptyKey {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrPathContainsEmptyElement,
				"FieldPath field names may not be empty strings.",
				document.Command(),
			)
		}

		return nil, lazyerrors.Error(err)
	}

	// findAndModify always works with a single document
	if resDocs, err = common.LimitDocuments(resDocs, 1); err != nil {
		return nil, err
	}

	if params.Update != nil { // we have update part
		var resValue any
		var insertedID any

		if params.Upsert { //  we have upsert flag
			var upsertParams *common.UpsertParams
			upsertParams, err = upsertDocuments(ctx, dbPool, resDocs, &qp, params)
			if err != nil {
				return nil, err
			}

			resValue = upsertParams.ReturnValue

			if upsertParams.Operation == common.UpsertOperationInsert {
				insertedID = must.NotFail(upsertParams.Upsert.Get("_id"))
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

			var upsert *types.Document

			if params.HasUpdateOperators {
				upsert = resDocs[0].DeepCopy()
				if _, err = common.UpdateDocument(upsert, params.Update); err != nil {
					return nil, err
				}

				if _, err = updateDocument(ctx, dbPool, &qp, upsert); err != nil {
					return nil, err
				}
			} else {
				upsert = params.Update

				if !upsert.Has("_id") {
					upsert.Set("_id", must.NotFail(resDocs[0].Get("_id")))
				}

				if _, err = updateDocument(ctx, dbPool, &qp, upsert); err != nil {
					return nil, err
				}
			}

			resValue = resDocs[0]
			if params.ReturnNewDocument {
				resValue = upsert
			}
		}

		lastErrorObject := must.NotFail(types.NewDocument(
			"n", int32(1),
			"updatedExisting", len(resDocs) > 0,
		))

		if insertedID != nil {
			lastErrorObject.Set("upserted", insertedID)
		}

		var reply wire.OpMsg
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"lastErrorObject", lastErrorObject,
				"value", resValue,
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
					"value", types.Null,
					"ok", float64(1),
				))},
			}))

			return &reply, nil
		}

		if _, err = deleteDocuments(ctx, dbPool, &qp, resDocs); err != nil {
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

// upsertDocuments inserts new document for insert operation,
// or updates the document.
func upsertDocuments(ctx context.Context, dbPool *tigrisdb.TigrisDB, docs []*types.Document, query *tigrisdb.QueryParams, params *common.FindAndModifyParams) (*common.UpsertParams, error) { //nolint:lll // argument list is too long
	res, err := common.PrepareDocumentForUpsert(docs, params)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	switch res.Operation {
	case common.UpsertOperationInsert:
		if err = insertDocument(ctx, dbPool, query, res.Upsert); err != nil {
			return nil, err
		}

		return res, nil
	case common.UpsertOperationUpdate:
		_, err = updateDocument(ctx, dbPool, query, res.Upsert)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return res, nil
	default:
		panic(fmt.Sprintf("unsupported upsert operation %s", res.Operation.String()))
	}
}
