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
	"fmt"
	"time"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
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

	username, _ := conninfo.Get(ctx).Auth()

	db, err := h.b.Database(params.DB)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", params.DB, params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "find")
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(params.Collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			msg := fmt.Sprintf("Invalid collection name: %s", params.Collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, "find")
		}

		return nil, lazyerrors.Error(err)
	}

	if params.Tailable {
		// TODO https://github.com/FerretDB/FerretDB/issues/3601
		res, err := db.ListCollections(ctx, &backends.ListCollectionsParams{})
		if err != nil {
			return nil, err
		}

		var collInfo *backends.CollectionInfo

		for _, coll := range res.Collections {
			if coll.Name == params.Collection {
				collInfo = &coll
			}
		}

		if collInfo == nil {
			// coll doesnt exist
		}

		if !collInfo.Capped() {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrBadValue,
				fmt.Sprintf(
					"error processing query: ns=%s.%sTree:"+
						"  $eq null\nSort: {}\nProj: {}\n tailable cursor requested on non capped collection",
					params.DB,
					params.Collection,
				),
				"find",
			)
		}

		return nil, common.Unimplemented(document, "tailable")
	}

	qp := &backends.QueryParams{
		Comment: params.Comment,
	}

	if params.Filter != nil {
		if qp.Comment, err = common.GetOptionalParam(params.Filter, "$comment", qp.Comment); err != nil {
			return nil, err
		}
	}

	if !h.DisableFilterPushdown {
		qp.Filter = params.Filter
	}

	// Skip sorting if there are more than one sort parameters
	if h.EnableUnsafeSortPushdown && params.Sort.Len() == 1 {
		var order types.SortType

		k := params.Sort.Keys()[0]
		v := params.Sort.Values()[0]

		order, err = common.GetSortType(k, v)
		if err != nil {
			return nil, err
		}

		qp.Sort = &backends.SortField{
			Key:        k,
			Descending: order == types.Descending,
		}
	}

	// Limit pushdown is not applied if:
	//  - `filter` is set, it must fetch all documents to filter them in memory;
	//  - `sort` is set but `UnsafeSortPushdown` is not set, it must fetch all documents
	//  and sort them in memory;
	//  - `skip` is non-zero value, skip pushdown is not supported yet.
	if params.Filter.Len() == 0 && (params.Sort.Len() == 0 || h.EnableUnsafeSortPushdown) && params.Skip == 0 {
		qp.Limit = params.Limit
	}

	cancel := func() {}
	if params.MaxTimeMS != 0 {
		// It is not clear if maxTimeMS affects only find, or both find and getMore (as the current code does).
		// TODO https://github.com/FerretDB/FerretDB/issues/2984
		ctx, cancel = context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
	}

	// closer accumulates all things that should be closed / canceled.
	closer := iterator.NewMultiCloser(iterator.CloserFunc(cancel))

	queryRes, err := c.Query(ctx, qp)
	if err != nil {
		closer.Close()
		return nil, lazyerrors.Error(err)
	}

	closer.Add(queryRes.Iter)

	iter := common.FilterIterator(queryRes.Iter, closer, params.Filter)

	iter, err = common.SortIterator(iter, closer, params.Sort)
	if err != nil {
		closer.Close()

		var pathErr *types.PathError
		if errors.As(err, &pathErr) && pathErr.Code() == types.ErrPathElementEmpty {
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

	// Combine iterators chain and closer into a cursor to pass around.
	// The context will be canceled when client disconnects or after maxTimeMS.
	cursor := h.cursors.NewCursor(ctx, &cursor.NewParams{
		Iter:         iterator.WithClose(iterator.Interface[struct{}, *types.Document](iter), closer.Close),
		DB:           params.DB,
		Collection:   params.Collection,
		Username:     username,
		ShowRecordID: params.ShowRecordId,
	})

	cursorID := cursor.ID

	firstBatchDocs, err := iterator.ConsumeValuesN(iterator.Interface[struct{}, *types.Document](cursor), int(params.BatchSize))
	if err != nil {
		cursor.Close()
		return nil, lazyerrors.Error(err)
	}

	firstBatch := types.MakeArray(len(firstBatchDocs))
	for _, doc := range firstBatchDocs {
		if params.ShowRecordId {
			doc.Set("$recordId", doc.RecordID())
		}

		firstBatch.Append(doc)
	}

	if params.SingleBatch || firstBatch.Len() < int(params.BatchSize) {
		// support tailable cursors
		// TODO https://github.com/FerretDB/FerretDB/issues/2283

		// let the client know that there are no more results
		cursorID = 0

		cursor.Close()
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
