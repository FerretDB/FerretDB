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
	"fmt"
	"strings"
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
			var upsert *types.Document
			var upserted bool

			if params.Upsert { //  we have upsert flag
				p := &upsertParams{
					hasUpdateOperators: params.HasUpdateOperators,
					query:              params.Query,
					update:             params.Update,
					queryParams:        &queryParams,
				}
				upsert, upserted, err = upsertDocuments(ctx, dbPool, tx, resDocs, p)
				if err != nil {
					return err
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

				if params.HasUpdateOperators {
					upsert = resDocs[0].DeepCopy()
					_, err = common.UpdateDocument(upsert, params.Update)
					if err != nil {
						return err
					}

					if _, err = updateDocument(ctx, tx, &queryParams, upsert); err != nil {
						return err
					}
				} else {
					upsert = params.Update

					if !upsert.Has("_id") {
						upsert.Set("_id", must.NotFail(resDocs[0].Get("_id")))
					}

					if _, err = updateDocument(ctx, tx, &queryParams, upsert); err != nil {
						return err
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

			var value any
			value = resultDoc
			if upserted && hasFilterOperator(params.Query) {
				value = types.Null
			}

			must.NoError(reply.SetSections(wire.OpMsgSection{
				Documents: []*types.Document{must.NotFail(types.NewDocument(
					"lastErrorObject", lastErrorObject,
					"value", value,
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

// upsertParams represent parameters for Handler.upsert method.
type upsertParams struct {
	hasUpdateOperators bool
	query, update      *types.Document
	queryParams        *pgdb.QueryParams
}

// upsertDocuments inserts new document if no documents in query result or updates given document.
// When inserting new document we must check that `_id` is present, so we must extract `_id` from query or generate a new one.
func upsertDocuments(ctx context.Context, dbPool *pgdb.Pool, tx pgx.Tx, docs []*types.Document, params *upsertParams) (*types.Document, bool, error) { //nolint:lll // argument list is too long
	// TODO split that block into own function since insert and update are very different
	// (and one uses dbPool, while other uses tx)
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
			upsert.Set("_id", getUpsertID(params.query))
		}

		if err := insertDocument(ctx, dbPool, params.queryParams, upsert); err != nil {
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
			v := must.NotFail(params.update.Get(k))
			if k == "_id" {
				return nil, false, commonerrors.NewCommandError(
					commonerrors.ErrImmutableField,
					fmt.Errorf(
						"Plan executor error during findAndModify :: caused by :: After applying the update, "+
							"the (immutable) field '_id' was found to have been altered to _id: \"%s\"",
						v,
					),
				)
			}
			upsert.Set(k, v)
		}
	}

	_, err := updateDocument(ctx, tx, params.queryParams, upsert)
	if err != nil {
		return nil, false, err
	}

	return upsert, false, nil
}

// getUpsertID gets id for upsert document
func getUpsertID(query *types.Document) any {
	id, err := query.Get("_id")
	if err != nil {
		return types.NewObjectID()
	}

	filter, ok := id.(*types.Document)
	if !ok {
		return id
	}

	if filter.Has("$exists") {
		return types.NewObjectID()
	}

	return id
}

// hasFilterOperator returns true if query contains any operator
func hasFilterOperator(query *types.Document) bool {
	iter := query.Iterator()
	defer iter.Close()

	for {
		k, v, err := iter.Next()
		if err != nil {
			return false
		}

		if strings.HasPrefix(k, "$") {
			return true
		}

		doc, ok := v.(*types.Document)
		if !ok {
			continue
		}

		if hasFilterOperator(doc) {
			return true
		}
	}
}
