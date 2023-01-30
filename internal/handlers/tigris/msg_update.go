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

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
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

	if err := common.Unimplemented(document, "let"); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "ordered", "writeConcern", "bypassDocumentValidation", "comment")

	var fp tigrisdb.FetchParam

	if fp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if fp.Collection, ok = collectionParam.(string); !ok {
		return nil, common.NewCommandErrorMsgWithArgument(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	var updates *types.Array
	if updates, err = common.GetOptionalParam(document, "updates", updates); err != nil {
		return nil, err
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
			"collation",
			"arrayFilters",
			"hint",
		}
		if err := common.Unimplemented(update, unimplementedFields...); err != nil {
			return nil, err
		}

		var q, u *types.Document
		var upsert bool
		var multi bool
		if q, err = common.GetOptionalParam(update, "q", q); err != nil {
			return nil, err
		}

		fp.Filter = q

		if u, err = common.GetOptionalParam(update, "u", u); err != nil {
			// TODO check if u is an array of aggregation pipeline stages
			return nil, err
		}
		if u != nil {
			if err = common.ValidateUpdateOperators(u); err != nil {
				return nil, err
			}
		}

		if upsert, err = common.GetOptionalParam(update, "upsert", upsert); err != nil {
			return nil, err
		}

		if multi, err = common.GetOptionalParam(update, "multi", multi); err != nil {
			return nil, err
		}

		resDocs, err := fetchAndFilterDocs(ctx, dbPool, &fp)
		if err != nil {
			return nil, err
		}

		if len(resDocs) == 0 {
			if !upsert {
				// nothing to do, continue to the next update operation
				continue
			}

			doc := q.DeepCopy()
			if _, err = common.UpdateDocument(doc, u); err != nil {
				return nil, err
			}
			if !doc.Has("_id") {
				doc.Set("_id", types.NewObjectID())
			}

			upserted.Append(must.NotFail(types.NewDocument(
				"index", int32(0), // TODO
				"_id", must.NotFail(doc.Get("_id")),
			)))

			if err = insertDocument(ctx, dbPool, &fp, doc); err != nil {
				return nil, err
			}

			matched++
			continue
		}

		if len(resDocs) > 1 && !multi {
			resDocs = resDocs[:1]
		}

		matched += int32(len(resDocs))

		for _, doc := range resDocs {
			changed, err := common.UpdateDocument(doc, u)
			if err != nil {
				return nil, err
			}

			if !changed {
				continue
			}

			res, err := updateDocument(ctx, dbPool, &fp, doc)
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
func updateDocument(ctx context.Context, dbPool *tigrisdb.TigrisDB, fp *tigrisdb.FetchParam, doc *types.Document) (int, error) {
	err := dbPool.ReplaceDocument(ctx, fp.DB, fp.Collection, doc)

	var valErr *types.ValidationError
	var driverErr *driver.Error

	switch {
	case err == nil:
		return 1, nil
	case errors.As(err, &valErr):
		return 0, common.NewCommandErrorMsg(common.ErrBadValue, err.Error())
	case errors.As(err, &driverErr):
		if tigrisdb.IsInvalidArgument(err) {
			return 0, common.NewCommandErrorMsg(common.ErrDocumentValidationFailure, err.Error())
		}

		return 0, lazyerrors.Error(err)
	default:
		return 0, lazyerrors.Error(err)
	}
}
