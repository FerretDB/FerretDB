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
	"time"

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/ctxutil"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgFind implements `find` command.
func (h *Handler) MsgFind(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetFindParams(document, h.L)
	if err != nil {
		return nil, err
	}

	username := conninfo.Get(ctx).Username()

	db, err := h.b.Database(params.DB)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", params.DB, params.Collection)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "find")
		}

		return nil, lazyerrors.Error(err)
	}

	coll, err := db.Collection(params.Collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "find")
		}

		return nil, lazyerrors.Error(err)
	}

	var cList *backends.ListCollectionsResult

	if cList, err = db.ListCollections(ctx, nil); err != nil {
		return nil, err
	}

	var cInfo backends.CollectionInfo

	// TODO https://github.com/FerretDB/FerretDB/issues/3601
	if i, found := slices.BinarySearchFunc(cList.Collections, params.Collection, func(e backends.CollectionInfo, t string) int {
		return cmp.Compare(e.Name, t)
	}); found {
		cInfo = cList.Collections[i]
	}

	capped := cInfo.Capped()
	if params.Tailable {
		if !capped {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrBadValue,
				"tailable cursor requested on non capped collection",
				"tailable",
			)
		}
	}

	qp, err := h.makeFindQueryParams(params, &cInfo)
	if err != nil {
		return nil, err
	}

	cancel := func() {}
	if params.MaxTimeMS != 0 {
		ctx, cancel = context.WithCancel(ctx)

		go func() {
			ctxutil.Sleep(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
			//cancel()
		}()
		// TODO context leak?

	}

	var ctxKeepAlive bool
	defer func() {
		if ctxKeepAlive {
			return
		}

		cancel()
	}()

	// closer accumulates all things that should be closed / canceled.
	closer := iterator.NewMultiCloser(iterator.CloserFunc(cancel))

	queryRes, err := coll.Query(ctx, qp)
	if err != nil {
		closer.Close()
		return nil, lazyerrors.Error(err)
	}

	iter, err := h.makeFindIter(queryRes.Iter, closer, params)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	t := cursor.Normal

	if params.Tailable {
		t = cursor.Tailable
	}

	if params.AwaitData {
		t = cursor.TailableAwait
	}

	c := h.cursors.NewCursor(context.Background(), iter, &cursor.NewParams{
		Data: &findCursorData{
			coll:       coll,
			qp:         qp,
			findParams: params,
		},
		DB:           params.DB,
		Collection:   params.Collection,
		Username:     username,
		Type:         t,
		ShowRecordID: params.ShowRecordId,
	})

	cursorID := c.ID

	docs, err := iterator.ConsumeValuesN(c, int(params.BatchSize))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h.L.Debug(
		"Got first batch", zap.Int64("cursor_id", cursorID), zap.Stringer("type", c.Type),
		zap.Int("count", len(docs)), zap.Int64("batch_size", params.BatchSize),
		zap.Bool("single_batch", params.SingleBatch),
	)

	if params.SingleBatch || len(docs) < int(params.BatchSize) {
		c.Close()

		// It is not entirely clear if we should do that; more tests are needed.
		if c.Type != cursor.Normal {
			h.cursors.CloseAndRemove(c)
		}

		// let the client know that there are no more results
		cursorID = 0
	}

	if cursorID != 0 {
		ctxKeepAlive = true
	}

	firstBatch := types.MakeArray(len(docs))
	for _, doc := range docs {
		firstBatch.Append(doc)
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

type findCursorData struct {
	coll       backends.Collection
	qp         *backends.QueryParams
	findParams *common.FindParams
}

// makeFindQueryParams creates the backend's query parameters for the find command.
func (h *Handler) makeFindQueryParams(params *common.FindParams, cInfo *backends.CollectionInfo) (*backends.QueryParams, error) {
	qp := &backends.QueryParams{
		Comment: params.Comment,
	}

	var err error
	if params.Filter != nil {
		if qp.Comment, err = common.GetOptionalParam(params.Filter, "$comment", qp.Comment); err != nil {
			return nil, err
		}
	}

	if !h.DisablePushdown {
		qp.Filter = params.Filter
	}

	if params.Sort, err = common.ValidateSortDocument(params.Sort); err != nil {
		var pathErr *types.PathError
		if errors.As(err, &pathErr) && pathErr.Code() == types.ErrPathElementEmpty {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrPathContainsEmptyElement,
				"Empty field names in path are not allowed",
				"find",
			)
		}

		return nil, err
	}

	switch {
	case h.DisablePushdown:
		// Pushdown disabled
	case params.Sort.Len() == 0 && cInfo.Capped():
		// Pushdown default recordID sorting for capped collections
		qp.Sort = must.NotFail(types.NewDocument("$natural", int64(1)))
	case params.Sort.Len() == 1:
		if params.Sort.Keys()[0] != "$natural" {
			break
		}

		if !cInfo.Capped() {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrNotImplemented,
				"$natural sort for non-capped collection is not supported.",
				"find",
			)
		}

		qp.Sort = params.Sort
	}

	// Limit pushdown is not applied if:
	//  - pushdown is disabled;
	//  - `filter` is set, it must fetch all documents to filter them in memory;
	//  - `sort` is set, it must fetch all documents and sort them in memory;
	//  - `skip` is non-zero value, skip pushdown is not supported yet.
	if !h.DisablePushdown && params.Filter.Len() == 0 && params.Sort.Len() == 0 && params.Skip == 0 {
		qp.Limit = params.Limit
	}

	h.L.Sugar().Debugf("Converted %+v for %+v to %+v.", params, cInfo, qp)

	return qp, nil
}

// makeFindIter creates an iterator chain for the find command.
//
// Iter is passed from the backend's query.
// All iterators, including the initial one, are added to the passed closer,
// and the returned iterator is wrapped with it.
//
//nolint:lll // for readability
func (h *Handler) makeFindIter(iter types.DocumentsIterator, closer *iterator.MultiCloser, params *common.FindParams) (types.DocumentsIterator, error) {
	closer.Add(iter)

	iter = common.FilterIterator(iter, closer, params.Filter)

	iter, err := common.SortIterator(iter, closer, params.Sort)
	if err != nil {
		closer.Close()

		var pathErr *types.PathError
		if errors.As(err, &pathErr) && pathErr.Code() == types.ErrPathElementEmpty {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrPathContainsEmptyElement,
				"Empty field names in path are not allowed",
				"find",
			)
		}

		return nil, lazyerrors.Error(err)
	}

	iter = common.SkipIterator(iter, closer, params.Skip)

	iter = common.LimitIterator(iter, closer, params.Limit)

	if iter, err = common.ProjectionIterator(iter, closer, params.Projection, params.Filter); err != nil {
		closer.Close()
		return nil, lazyerrors.Error(err)
	}

	return iterator.WithClose(iter, closer.Close), nil
}
