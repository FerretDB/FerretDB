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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDelete implements HandlerInterface.
func (h *Handler) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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

	common.Ignored(document, h.L, "writeConcern")

	var deletes *types.Array
	if deletes, err = common.GetOptionalParam(document, "deletes", deletes); err != nil {
		return nil, err
	}

	ordered := true
	if ordered, err = common.GetOptionalParam(document, "ordered", ordered); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "comment")

	var qp tigrisdb.QueryParams

	if qp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if qp.Collection, ok = collectionParam.(string); !ok {
		return nil, common.NewCommandErrorMsgWithArgument(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	var deleted int32
	var delErrors common.WriteErrors

	// process every delete filter
	for i := 0; i < deletes.Len(); i++ {
		// get document with filter
		deleteDoc, err := common.AssertType[*types.Document](must.NotFail(deletes.Get(i)))
		if err != nil {
			return nil, err
		}

		filter, limit, err := h.prepareDeleteParams(deleteDoc)
		if err != nil {
			return nil, err
		}

		del, err := h.execDelete(ctx, dbPool, &qp, filter, limit)
		if err == nil {
			deleted += del
			continue
		}

		delErrors.Append(err, int32(i))

		if ordered {
			break
		}
	}

	replyDoc := must.NotFail(types.NewDocument(
		"ok", float64(1),
	))

	if delErrors.Len() > 0 {
		replyDoc = delErrors.Document()
	}

	replyDoc.Set("n", deleted)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{replyDoc},
	}))

	return &reply, nil
}

// prepareDeleteParams extracts query filter and limit from delete document.
func (h *Handler) prepareDeleteParams(deleteDoc *types.Document) (*types.Document, int64, error) {
	var err error

	if err = common.Unimplemented(deleteDoc, "collation", "hint"); err != nil {
		return nil, 0, err
	}

	// get filter from document
	var filter *types.Document
	if filter, err = common.GetOptionalParam(deleteDoc, "q", filter); err != nil {
		return nil, 0, err
	}

	common.Ignored(filter, h.L, "$comment")

	l, err := deleteDoc.Get("limit")
	if err != nil {
		return nil, 0, common.NewCommandErrorMsgWithArgument(
			common.ErrMissingField,
			"BSON field 'delete.deletes.limit' is missing but a required field",
			"limit",
		)
	}

	var limit int64
	if limit, err = common.GetWholeNumberParam(l); err != nil || limit < 0 || limit > 1 {
		return nil, 0, common.NewCommandErrorMsgWithArgument(
			common.ErrFailedToParse,
			fmt.Sprintf("The limit field in delete objects must be 0 or 1. Got %v", l),
			"limit",
		)
	}

	return filter, limit, nil
}

// execDelete fetches documents, filter them out and limiting with the given limit value.
// It returns the number of deleted documents or an error.
func (h *Handler) execDelete(ctx context.Context, dbPool *tigrisdb.TigrisDB, qp *tigrisdb.QueryParams, filter *types.Document, limit int64) (int32, error) { //nolint:lll // argument list is too long
	var err error

	resDocs := make([]*types.Document, 0, 16)

	var deleted int32

	// fetch current items from collection
	iter, err := dbPool.QueryDocuments(ctx, qp)
	if err != nil {
		return 0, err
	}

	defer iter.Close()

	// iterate through every document and delete matching ones
	for {
		var doc *types.Document

		_, doc, err = iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return 0, lazyerrors.Error(err)
		}

		// fetch current items from collection
		matches, err := common.FilterDocument(doc, filter)
		if err != nil {
			return 0, err
		}

		if !matches {
			continue
		}

		resDocs = append(resDocs, doc)
	}

	if resDocs, err = common.LimitDocuments(resDocs, limit); err != nil {
		return 0, err
	}

	// if no field is matched in a row, go to the next one
	if len(resDocs) == 0 {
		return 0, nil
	}

	res, err := deleteDocuments(ctx, dbPool, qp, resDocs)
	if err != nil {
		return 0, err
	}

	deleted += int32(res)

	return deleted, nil
}

// deleteDocuments deletes documents by _id.
func deleteDocuments(ctx context.Context, dbPool *tigrisdb.TigrisDB, qp *tigrisdb.QueryParams, docs []*types.Document) (int, error) { //nolint:lll // argument list is too long
	ids := make([]map[string]any, len(docs))
	for i, doc := range docs {
		id := must.NotFail(tjson.Marshal(must.NotFail(doc.Get("_id"))))
		ids[i] = map[string]any{"_id": json.RawMessage(id)}
	}

	var f driver.Filter
	switch len(ids) {
	case 0:
		f = driver.Filter(`{}`)
	case 1:
		f = must.NotFail(json.Marshal(ids[0]))
	default:
		f = must.NotFail(json.Marshal(map[string]any{"$or": ids}))
	}

	_, err := dbPool.Driver.UseDatabase(qp.DB).Delete(ctx, qp.Collection, f)
	if err != nil {
		return 0, lazyerrors.Error(err)
	}

	return len(ids), nil
}
