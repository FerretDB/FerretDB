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

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgUpdate implements HandlerInterface.
func (h *Handler) MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetUpdateParams(document, h.L)
	if err != nil {
		return nil, err
	}

	var matched, modified int32
	var upserted types.Array

	for _, u := range params.Updates {
		qp := tigrisdb.QueryParams{
			DB:         params.DB,
			Collection: params.Collection,
			Filter:     u.Filter,
		}

		resDocs, err := fetchAndFilterDocs(ctx, &fetchParams{dbPool, &qp, h.DisableFilterPushdown})
		if err != nil {
			return nil, err
		}

		if len(resDocs) == 0 {
			if !u.Upsert {
				// nothing to do, continue to the next update operation
				continue
			}

			doc := u.Filter.DeepCopy()
			if _, err = common.UpdateDocument(doc, u.Update); err != nil {
				return nil, err
			}
			if !doc.Has("_id") {
				doc.Set("_id", types.NewObjectID())
			}

			upserted.Append(must.NotFail(types.NewDocument(
				"index", int32(0), // TODO
				"_id", must.NotFail(doc.Get("_id")),
			)))

			if err = insertDocument(ctx, dbPool, &qp, doc); err != nil {
				return nil, err
			}

			matched++
			continue
		}

		if len(resDocs) > 1 && !u.Multi {
			resDocs = resDocs[:1]
		}

		matched += int32(len(resDocs))

		for _, doc := range resDocs {
			changed, err := common.UpdateDocument(doc, u.Update)
			if err != nil {
				return nil, err
			}

			if !changed {
				continue
			}

			res, err := updateDocument(ctx, dbPool, &qp, doc)
			if err != nil {
				return nil, err
			}
			modified += int32(res)
		}
	}

	res := must.NotFail(types.NewDocument(
		"n", matched,
	))

	if upserted.Len() != 0 {
		res.Set("upserted", &upserted)
	}

	res.Set("nModified", modified)
	res.Set("ok", float64(1))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{res},
	}))

	return &reply, nil
}

// updateDocument replaces given document.
func updateDocument(ctx context.Context, dbPool *tigrisdb.TigrisDB, qp *tigrisdb.QueryParams, doc *types.Document) (int, error) {
	err := dbPool.ReplaceDocument(ctx, qp.DB, qp.Collection, doc)

	var valErr *types.ValidationError
	var driverErr *driver.Error

	switch {
	case err == nil:
		return 1, nil
	case errors.As(err, &valErr):
		return 0, commonerrors.NewCommandErrorMsg(commonerrors.ErrBadValue, err.Error())
	case errors.As(err, &driverErr):
		if tigrisdb.IsInvalidArgument(err) {
			return 0, commonerrors.NewCommandErrorMsg(commonerrors.ErrDocumentValidationFailure, err.Error())
		}

		return 0, lazyerrors.Error(err)
	default:
		return 0, lazyerrors.Error(err)
	}
}
