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

package sqlite

import (
	"context"
	"errors"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgUpdate implements HandlerInterface.
func (h *Handler) MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetUpdateParams(document, h.L)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	matched, modified, upserted, err := h.updateDocument(ctx, params)
	if err != nil {
		return nil, lazyerrors.Error(err)
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

// updateDocument iterate through all documents in collection and update them.
func (h *Handler) updateDocument(ctx context.Context, params *common.UpdatesParams) (int32, int32, *types.Array, error) {
	var matched, modified int32
	var upserted types.Array

	db := h.b.Database(params.DB)
	defer db.Close()

	err := db.CreateCollection(ctx, &backends.CreateCollectionParams{Name: params.Collection})
	if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionAlreadyExists) {
		err = nil
	}

	if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
		//TODO: check if its needed
		return 0, 0, nil, err
	}

	if err != nil {
		return 0, 0, nil, lazyerrors.Error(err)
	}

	for _, u := range params.Updates {
		res, err := db.Collection(params.Collection).Query(ctx, nil)
		if err != nil {
			return 0, 0, nil, lazyerrors.Error(err)
		}

		var resDocs []*types.Document

		defer res.Iter.Close()

		for {
			var doc *types.Document

			_, doc, err = res.Iter.Next()
			if err != nil {
				if errors.Is(err, iterator.ErrIteratorDone) {
					break
				}

				return 0, 0, nil, lazyerrors.Error(err)
			}

			var matches bool

			matches, err = common.FilterDocument(doc, u.Filter)
			if err != nil {
				return 0, 0, nil, lazyerrors.Error(err)
			}

			if !matches {
				continue
			}

			resDocs = append(resDocs, doc)
		}

		res.Iter.Close()

		if len(resDocs) == 0 {
			if !u.Upsert {
				// nothing to do, continue to the next update operation
				continue
			}

			doc := u.Filter.DeepCopy()
			if _, err = common.UpdateDocument("update", doc, u.Update); err != nil {
				return 0, 0, nil, lazyerrors.Error(err)
			}

			if !doc.Has("_id") {
				doc.Set("_id", types.NewObjectID())
			}

			upserted.Append(must.NotFail(types.NewDocument(
				"index", int32(0), // TODO
				"_id", must.NotFail(doc.Get("_id")),
			)))

			// TODO https://github.com/FerretDB/FerretDB/issues/2612

			_, err := db.Collection(params.Collection).Insert(ctx, &backends.InsertParams{
				Iter: must.NotFail(types.NewArray(doc)).Iterator(),
			})
			if err != nil {
				return 0, 0, nil, err
			}

			matched++

			continue
		}

		if len(resDocs) > 1 && !u.Multi {
			resDocs = resDocs[:1]
		}

		matched += int32(len(resDocs))

		for _, doc := range resDocs {
			changed, err := common.UpdateDocument("update", doc, u.Update)
			if err != nil {
				return 0, 0, nil, lazyerrors.Error(err)
			}

			if !changed {
				continue
			}

			updateRes, err := db.
				Collection(params.Collection).
				Update(ctx, &backends.UpdateParams{Docs: must.NotFail(types.NewArray(doc))})
			if err != nil {
				return 0, 0, nil, lazyerrors.Error(err)
			}

			modified += int32(updateRes.Updated)
		}
	}

	return matched, modified, &upserted, nil
}
