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

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
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

	// This is not very optimal as we need to fetch everything from the database to have a proper sort.
	// We might consider rewriting it later.
	resDocs := h.fetch(ctx, params.SQLParam)


		var fetchedDocs []*types.Document
		for fetchedItem := range fetchedChan {
			if fetchedItem.Err != nil {
				return fetchedItem.Err
			}

			fetchedDocs = append(fetchedDocs, fetchedItem.Docs...)
		}

		err = common.SortDocuments(fetchedDocs, params.Sort)
		if err != nil {
			return err
		}

		for _, doc := range fetchedDocs {
			matches, err := common.FilterDocument(doc, params.Query)
			if err != nil {
				return err
			}

			if !matches {
				continue
			}

			resDocs = append(resDocs, doc)
		}

		return nil
	})

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
				sqlParam:           params.SQLParam,
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

				_, err = h.update(ctx, &params.SQLParam, upsert)
				if err != nil {
					return nil, err
				}
			} else {
				upsert = params.Update

				if !upsert.Has("_id") {
					must.NoError(upsert.Set("_id", must.NotFail(resDocs[0].Get("_id"))))
				}

				_, err = h.update(ctx, &params.SQLParam, upsert)
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
			must.NoError(lastErrorObject.Set("upserted", must.NotFail(resultDoc.Get("_id"))))
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

		_, err = h.delete(ctx, &params.SQLParam, resDocs)
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
