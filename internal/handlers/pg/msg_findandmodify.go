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
	"errors"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
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

	queryParams := pgdb.QueryParams{
		DB:         params.DB,
		Collection: params.Collection,
		Comment:    params.Comment,
		Filter:     params.Query,
	}

	// This is not very optimal as we need to fetch everything from the database to have a proper sort.
	// We might consider rewriting it later.
	var reply wire.OpMsg
	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		var resDocs []*types.Document
		resDocs, err = fetchAndFilterDocs(ctx, &fetchParams{tx, &queryParams, h.DisablePushdown})
		if err != nil {
			return err
		}

		if err = common.SortDocuments(resDocs, params.Sort); err != nil {
			var pathErr *types.DocumentPathError
			if errors.As(err, &pathErr) && pathErr.Code() == types.ErrDocumentPathEmptyKey {
				return commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrPathContainsEmptyElement,
					"FieldPath field names may not be empty strings.",
					document.Command(),
				)
			}

			return lazyerrors.Error(err)
		}

		// findAndModify always works with a single document
		if resDocs, err = common.LimitDocuments(resDocs, 1); err != nil {
			return err
		}

		if params.Update != nil { // we have update part
			var resValue any
			var insertedID any

			if params.Upsert { //  we have upsert flag
				var res *common.UpsertResult
				res, err = upsertDocuments(ctx, dbPool, tx, resDocs, &queryParams, params)
				if err != nil {
					return err
				}

				resValue = res.ReturnValue

				if res.Insert != nil {
					insertedID = must.NotFail(res.Insert.Get("_id"))
				}
			} else { // process update as usual
				if len(resDocs) == 0 {
					must.NoError(reply.SetSections(wire.OpMsgSection{
						Documents: []*types.Document{must.NotFail(types.NewDocument(
							"lastErrorObject", must.NotFail(types.NewDocument("n", int32(0), "updatedExisting", false)),
							"value", types.Null,
							"ok", float64(1),
						))},
					}))

					return nil
				}

				var upsert *types.Document

				if params.HasUpdateOperators {
					upsert = resDocs[0].DeepCopy()
					_, err = common.UpdateDocument(upsert, params.Update)
					if err != nil {
						return err
					}
				} else {
					upsert = params.Update

					if !upsert.Has("_id") {
						upsert.Set("_id", must.NotFail(resDocs[0].Get("_id")))
					}
				}

				if _, err = updateDocument(ctx, tx, &queryParams, upsert); err != nil {
					return err
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

			must.NoError(reply.SetSections(wire.OpMsgSection{
				Documents: []*types.Document{must.NotFail(types.NewDocument(
					"lastErrorObject", lastErrorObject,
					"value", resValue,
					"ok", float64(1),
				))},
			}))

			return nil
		}

		if params.Remove {
			if len(resDocs) == 0 {
				must.NoError(reply.SetSections(wire.OpMsgSection{
					Documents: []*types.Document{must.NotFail(types.NewDocument(
						"lastErrorObject", must.NotFail(types.NewDocument("n", int32(0))),
						"value", types.Null,
						"ok", float64(1),
					))},
				}))

				return nil
			}

			if _, err = deleteDocuments(ctx, dbPool, &queryParams, resDocs); err != nil {
				return err
			}

			must.NoError(reply.SetSections(wire.OpMsgSection{
				Documents: []*types.Document{must.NotFail(types.NewDocument(
					"lastErrorObject", must.NotFail(types.NewDocument("n", int32(1))),
					"value", resDocs[0],
					"ok", float64(1),
				))},
			}))
			return nil
		}

		return lazyerrors.New("bad flags combination")
	})

	if err != nil {
		return nil, err
	}

	return &reply, nil
}

// upsertDocuments inserts new document if no documents in query result or updates given document.
func upsertDocuments(ctx context.Context, dbPool *pgdb.Pool, tx pgx.Tx, docs []*types.Document, query *pgdb.QueryParams, params *common.FindAndModifyParams) (*common.UpsertResult, error) { //nolint:lll // argument list is too long
	res, err := common.UpsertDocument(docs, params)
	if err != nil {
		return nil, err
	}

	if res.Insert != nil {
		if err = insertDocument(ctx, dbPool, query, res.Insert); err != nil {
			return nil, err
		}

		return res, nil
	}

	_, err = updateDocument(ctx, tx, query, res.Update)
	if err != nil {
		return nil, err
	}

	return res, nil
}
