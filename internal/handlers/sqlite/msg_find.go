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
	"time"

	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgFind implements HandlerInterface.
func (h *Handler) MsgFind(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetFindParams(document, h.L)
	if err != nil {
		return nil, err
	}

	if params.BatchSize == 0 {
		var reply wire.OpMsg
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"cursor", must.NotFail(types.NewDocument(
					"firstBatch", types.MakeArray(0),
					"id", int64(0),
					"ns", params.DB+"."+params.Collection,
				)),
				"ok", float64(1),
			))},
		}))

		return &reply, nil
	}

	db := h.b.Database(params.DB)
	defer db.Close()

	cancel := func() {}
	if params.MaxTimeMS != 0 {
		// It is not if maxTimeMS affects only find, or both find and getMore (as the current code does).
		// TODO https://github.com/FerretDB/FerretDB/issues/1808
		ctx, cancel = context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
	}

	queryRes, err := db.Collection(params.Collection).Query(ctx, nil)
	if err != nil {
		cancel()
		return nil, lazyerrors.Error(err)
	}

	iter := queryRes.Iter

	closer := iterator.NewMultiCloser(iter, iterator.CloserFunc(cancel))

	iter = common.FilterIterator(iter, closer, params.Filter)

	iter, err = common.SortIterator(iter, closer, params.Sort)
	if err != nil {
		closer.Close()

		var pathErr *types.DocumentPathError
		if errors.As(err, &pathErr) && pathErr.Code() == types.ErrDocumentPathEmptyKey {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrPathContainsEmptyElement,
				"Empty field names in path are not allowed",
				document.Command(),
			)
		}

		return nil, lazyerrors.Error(err)
	}

	iter = common.SkipIterator(iter, closer, params.Skip)

	iter = common.LimitIterator(iter, closer, params.Limit)

	iter, err = common.ProjectionIterator(iter, closer, params.Projection, params.Filter)
	if err != nil {
		closer.Close()
		return nil, lazyerrors.Error(err)
	}

	cIter := iterator.WithClose(iterator.Interface[struct{}, *types.Document](iter), closer.Close)

	var cursorID int64
	var docs []*types.Document

	if h.EnableCursors {
		cursor := h.cursors.NewCursor(ctx, &cursor.NewParams{
			Iter:       cIter,
			DB:         params.DB,
			Collection: params.Collection,
		})

		cursorID = cursor.ID

		cIter = cursor

		docs, err = iterator.ConsumeValuesN(cIter, int(params.BatchSize))
	} else {
		docs, err = iterator.ConsumeValues(cIter)
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	firstBatch := types.MakeArray(len(docs))
	for _, doc := range docs {
		firstBatch.Append(doc)
	}

	if firstBatch.Len() < int(params.BatchSize) {
		// Cursor ID 0 lets the client know that there are no more results.
		// Cursor is already closed and removed from the registry by this point.
		cursorID = 0
	}

	if params.SingleBatch {
		cIter.Close()
		cursorID = 0
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"firstBatch", firstBatch,
				"id", cursorID,
				"ns", params.DB+"."+params.Collection,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
