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
	"time"

	"github.com/jackc/pgx/v4"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
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
	common.Ignored(document, h.l, ignoredFields...)

	params, err := prepareFindAndModifyParams(document)
	if err != nil {
		return nil, err
	}

	if params.maxTimeMS != 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(params.maxTimeMS)*time.Millisecond)
		defer cancel()

		ctx = ctxWithTimeout
	}

	// This is not very optimal as we need to fetch everything from the database to have a proper sort.
	// We might consider rewriting it later.
	resDocs := make([]*types.Document, 0, 16)
	err = h.pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		fetchedChan, err := h.pgPool.QueryDocuments(ctx, tx, params.sqlParam)
		if err != nil {
			return err
		}
		defer func() {
			// Drain the channel to prevent leaking goroutines.
			// TODO Offer a better design instead of channels: https://github.com/FerretDB/FerretDB/issues/898.
			for range fetchedChan {
			}
		}()

		var fetchedDocs []*types.Document
		for fetchedItem := range fetchedChan {
			if fetchedItem.Err != nil {
				return fetchedItem.Err
			}

			fetchedDocs = append(fetchedDocs, fetchedItem.Docs...)
		}

		err = common.SortDocuments(fetchedDocs, params.sort)
		if err != nil {
			return err
		}

		for _, doc := range fetchedDocs {
			matches, err := common.FilterDocument(doc, params.query)
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

	if params.update != nil { // we have update part
		var upsert *types.Document
		var upserted bool

		if params.upsert { //  we have upsert flag
			p := &upsertParams{
				hasUpdateOperators: params.hasUpdateOperators,
				query:              params.query,
				update:             params.update,
				sqlParam:           params.sqlParam,
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

			if params.hasUpdateOperators {
				upsert = resDocs[0].DeepCopy()
				_, err = common.UpdateDocument(upsert, params.update)
				if err != nil {
					return nil, err
				}

				_, err = h.update(ctx, &params.sqlParam, upsert)
				if err != nil {
					return nil, err
				}
			} else {
				upsert = params.update

				if !upsert.Has("_id") {
					must.NoError(upsert.Set("_id", must.NotFail(resDocs[0].Get("_id"))))
				}

				_, err = h.update(ctx, &params.sqlParam, upsert)
				if err != nil {
					return nil, err
				}
			}
		}

		var resultDoc *types.Document
		if params.returnNewDocument || len(resDocs) == 0 {
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

	if params.remove {
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

		_, err = h.delete(ctx, &params.sqlParam, resDocs)
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
	sqlParam           pgdb.SQLParam
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
				must.NoError(upsert.Set("_id", must.NotFail(params.query.Get("_id"))))
			} else {
				must.NoError(upsert.Set("_id", types.NewObjectID()))
			}
		}

		err := h.insert(ctx, params.sqlParam, upsert)
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
			must.NoError(upsert.Set(k, must.NotFail(params.update.Get(k))))
		}
	}

	_, err := h.update(ctx, &params.sqlParam, upsert)
	if err != nil {
		return nil, false, err
	}

	return upsert, false, nil
}

// findAndModifyParams represent all findAndModify requests' fields.
// It's filled by calling prepareFindAndModifyParams.
type findAndModifyParams struct {
	sqlParam                              pgdb.SQLParam
	query, sort, update                   *types.Document
	remove, upsert                        bool
	returnNewDocument, hasUpdateOperators bool
	maxTimeMS                             int32
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

	maxTimeMS, err := common.GetOptionalPositiveNumber(document, "maxTimeMS")
	if err != nil {
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

	if update != nil && remove {
		return nil, common.NewErrorMsg(common.ErrFailedToParse, "Cannot specify both an update and remove=true")
	}
	if upsert && remove {
		return nil, common.NewErrorMsg(common.ErrFailedToParse, "Cannot specify both upsert=true and remove=true")
	}
	if returnNewDocument && remove {
		return nil, common.NewErrorMsg(
			common.ErrFailedToParse,
			"Cannot specify both new=true and remove=true; 'remove' always returns the deleted document",
		)
	}

	var hasUpdateOperators bool
	for k := range update.Map() {
		if slices.Contains(
			[]string{
				"$currentDate",
				"$set",
				"$setOnInsert",
				"$unset",
				"$inc",
			},
			k,
		) {
			hasUpdateOperators = true
			break
		}
	}

	var comment string
	// get comment from a "comment" field
	if comment, err = common.GetOptionalParam(document, "comment", comment); err != nil {
		return nil, err
	}

	// get comment from query, e.g. db.collection.FindAndModify({"_id":"string", "$comment: "test"},{$set:{"v":"foo""}})
	if comment, err = common.GetOptionalParam(query, "$comment", comment); err != nil {
		return nil, err
	}

	return &findAndModifyParams{
		sqlParam: pgdb.SQLParam{
			DB:         db,
			Collection: collection,
			Comment:    comment,
		},
		query:              query,
		update:             update,
		sort:               sort,
		remove:             remove,
		upsert:             upsert,
		returnNewDocument:  returnNewDocument,
		hasUpdateOperators: hasUpdateOperators,
		maxTimeMS:          maxTimeMS,
	}, nil
}
