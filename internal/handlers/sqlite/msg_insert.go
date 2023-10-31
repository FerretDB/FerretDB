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
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// writeError represents a single write error details.
//
// Find a better place for this struct.
// TODO https://github.com/FerretDB/FerretDB/issues/3263
type writeError struct {
	// the order of fields is weird to make the struct smaller due to alignment

	errmsg string
	index  int32
	code   commonerrors.ErrorCode
}

// Document returns a document representation of the write error.
func (we *writeError) Document() *types.Document {
	return must.NotFail(types.NewDocument(
		"index", we.index,
		"code", int32(we.code),
		"errmsg", we.errmsg,
	))
}

// MsgInsert implements HandlerInterface.
func (h *Handler) MsgInsert(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
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
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "insert")
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(params.Collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "insert")
		}

		return nil, lazyerrors.Error(err)
	}

	docsIter := params.Docs.Iterator()
	defer docsIter.Close()

	var inserted int32
	var writeErrors []*writeError

	var done bool
	for !done {
		const batchSize = 1000
		docs := make([]*types.Document, 0, batchSize)
		docsIndexes := make([]int32, 0, batchSize)

		for j := 0; j < batchSize; j++ {
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
				docsIndexes = append(docsIndexes, int32(i))

				continue
			}

			var ve *types.ValidationError
			if !errors.As(err, &ve) {
				return nil, lazyerrors.Error(err)
			}

			var code commonerrors.ErrorCode
			switch ve.Code() {
			case types.ErrValidation, types.ErrIDNotFound:
				code = commonerrors.ErrBadValue
			case types.ErrWrongIDType:
				code = commonerrors.ErrInvalidID
			default:
				panic(fmt.Sprintf("Unknown error code: %v", ve.Code()))
			}

			writeErrors = append(writeErrors, &writeError{
				index:  int32(i),
				code:   code,
				errmsg: ve.Error(),
			})

			if params.Ordered {
				break
			}
		}

		if _, err = c.InsertAll(ctx, &backends.InsertAllParams{Docs: docs}); err == nil {
			inserted += int32(len(docs))

			if params.Ordered && len(writeErrors) > 0 {
				break
			}

			continue
		}

		// insert doc one by one upon failing on batch insertion
		for j, doc := range docs {
			if _, err = c.InsertAll(ctx, &backends.InsertAllParams{
				Docs: []*types.Document{doc},
			}); err == nil {
				inserted++

				continue
			}

			if !backends.ErrorCodeIs(err, backends.ErrorCodeInsertDuplicateID) {
				return nil, lazyerrors.Error(err)
			}

			writeErrors = append(writeErrors, &writeError{
				index:  docsIndexes[j],
				code:   commonerrors.ErrDuplicateKeyInsert,
				errmsg: fmt.Sprintf(`E11000 duplicate key error collection: %s.%s`, params.DB, params.Collection),
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
		slices.SortFunc(writeErrors, func(a, b *writeError) int {
			return cmp.Compare(a.index, b.index)
		})

		array := types.MakeArray(len(writeErrors))
		for _, we := range writeErrors {
			array.Append(we.Document())
		}

		res.Set("writeErrors", array)
	}

	res.Set("ok", float64(1))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{res},
	}))

	return &reply, nil
}
