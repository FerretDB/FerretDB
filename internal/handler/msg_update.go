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

package handler

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgUpdate implements `update` command.
func (h *Handler) MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetUpdateParams(document, h.L)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/2612
	_ = params.Ordered

	var we *mongo.WriteError

	matched, modified, upserted, err := h.updateDocument(ctx, params)
	if err != nil {
		we, err = handleUpdateError(params.DB, params.Collection, err)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	res := must.NotFail(types.NewDocument(
		"n", matched,
	))

	if we != nil {
		res.Set("writeErrors", must.NotFail(types.NewArray(WriteErrorDocument(we))))
	}

	if upserted.Len() != 0 {
		res.Set("upserted", upserted)
	}

	res.Set("nModified", modified)
	res.Set("ok", float64(1))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.MakeOpMsgSection(
		res,
	)))

	return &reply, nil
}

// updateDocument iterate through all documents in collection and update them.
func (h *Handler) updateDocument(ctx context.Context, params *common.UpdateParams) (int32, int32, *types.Array, error) {
	var matched, modified int32
	var upserted types.Array

	db, err := h.b.Database(params.DB)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", params.DB, params.Collection)
			return 0, 0, nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "update")
		}

		return 0, 0, nil, lazyerrors.Error(err)
	}

	err = db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: params.Collection})

	switch {
	case err == nil:
		// nothing
	case backends.ErrorCodeIs(err, backends.ErrorCodeCollectionAlreadyExists):
		// nothing
	case backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid):
		msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
		return 0, 0, nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "insert")
	default:
		return 0, 0, nil, lazyerrors.Error(err)
	}

	for _, u := range params.Updates {
		c, err := db.Collection(params.Collection)
		if err != nil {
			if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
				msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
				return 0, 0, nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "insert")
			}

			return 0, 0, nil, lazyerrors.Error(err)
		}

		var qp backends.QueryParams
		if !h.DisablePushdown {
			qp.Filter = u.Filter
		}

		res, err := c.Query(ctx, &qp)
		if err != nil {
			return 0, 0, nil, lazyerrors.Error(err)
		}

		closer := iterator.NewMultiCloser()
		defer closer.Close()

		closer.Add(res.Iter)

		iter := common.FilterIterator(res.Iter, closer, u.Filter)

		if !u.Multi {
			iter = common.LimitIterator(iter, closer, 1)
		}

		result, err := common.UpdateDocument(ctx, c, "update", iter, &u)
		if err != nil {
			return 0, 0, nil, lazyerrors.Error(err)
		}

		matched += result.Matched.Count
		modified += result.Modified.Count

		if result.Upserted.Doc != nil {
			doc := result.Upserted.Doc
			upserted.Append(must.NotFail(types.NewDocument(
				"index", int32(upserted.Len()),
				"_id", must.NotFail(doc.Get("_id")),
			)))

			// in case of upsert, MongoDB sets the matched count to 1
			matched++
		}
	}

	return matched, modified, &upserted, nil
}
