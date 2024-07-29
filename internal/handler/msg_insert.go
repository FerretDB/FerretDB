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
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/FerretDB/wire"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// WriteErrorDocument returns a document representation of the write error.
//
// Find a better place for this function.
// TODO https://github.com/FerretDB/FerretDB/issues/3263
func WriteErrorDocument(we *mongo.WriteError) *types.Document {
	return must.NotFail(types.NewDocument(
		"index", int32(we.Index),
		"code", int32(we.Code),
		"errmsg", we.Message,
	))
}

// MsgInsert implements `insert` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgInsert(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := bson.OpMsgDocument(msg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetInsertParams(document, h.L)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	db, err := h.b.Database(params.DB)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", params.DB, params.Collection)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "insert")
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(params.Collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "insert")
		}

		return nil, lazyerrors.Error(err)
	}

	docsIter := params.Docs.Iterator()
	defer docsIter.Close()

	var inserted int32
	var writeErrors []*mongo.WriteError

	var done bool
	for !done {
		docs := make([]*types.Document, 0, h.BatchSize)
		docsIndexes := make([]int, 0, h.BatchSize)

		for j := 0; j < h.BatchSize; j++ {
			var i int
			var d any

			i, d, err = docsIter.Next()
			if errors.Is(err, iterator.ErrIteratorDone) {
				done = true
				break
			}

			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			doc := d.(*types.Document)

			if !doc.Has("_id") {
				doc.Set("_id", types.NewObjectID())
			}

			// TODO https://github.com/FerretDB/FerretDB/issues/3454
			if err = doc.ValidateData(); err == nil {
				docs = append(docs, doc)
				docsIndexes = append(docsIndexes, i)

				continue
			}

			var ve *types.ValidationError
			if !errors.As(err, &ve) {
				return nil, lazyerrors.Error(err)
			}

			var code handlererrors.ErrorCode
			switch ve.Code() {
			case types.ErrValidation, types.ErrIDNotFound:
				code = handlererrors.ErrBadValue
			case types.ErrWrongIDType:
				code = handlererrors.ErrInvalidID
			default:
				panic(fmt.Sprintf("Unknown error code: %v", ve.Code()))
			}

			writeErrors = append(writeErrors, &mongo.WriteError{
				Index:   i,
				Code:    int(code),
				Message: ve.Error(),
			})

			if params.Ordered {
				break
			}
		}

		if _, err = c.InsertAll(connCtx, &backends.InsertAllParams{Docs: docs}); err == nil {
			inserted += int32(len(docs))

			if params.Ordered && len(writeErrors) > 0 {
				break
			}

			continue
		}

		// insert doc one by one upon failing on batch insertion
		for j, doc := range docs {
			if _, err = c.InsertAll(connCtx, &backends.InsertAllParams{
				Docs: []*types.Document{doc},
			}); err == nil {
				inserted++

				continue
			}

			if !backends.ErrorCodeIs(err, backends.ErrorCodeInsertDuplicateID) {
				return nil, lazyerrors.Error(err)
			}

			writeErrors = append(writeErrors, &mongo.WriteError{
				Index:   docsIndexes[j],
				Code:    int(handlererrors.ErrDuplicateKeyInsert),
				Message: fmt.Sprintf(`E11000 duplicate key error collection: %s.%s`, params.DB, params.Collection),
			})

			if params.Ordered {
				break
			}
		}
	}

	res := must.NotFail(types.NewDocument(
		"n", inserted,
	))

	if len(writeErrors) > 0 {
		slices.SortFunc(writeErrors, func(a, b *mongo.WriteError) int {
			return cmp.Compare(a.Index, b.Index)
		})

		array := types.MakeArray(len(writeErrors))
		for _, we := range writeErrors {
			array.Append(WriteErrorDocument(we))
		}

		res.Set("writeErrors", array)
	}

	res.Set("ok", float64(1))

	return bson.NewOpMsg(
		res,
	)
}
