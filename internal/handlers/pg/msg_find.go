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
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgFind implements HandlerInterface.
func (h *Handler) MsgFind(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetFindParams(document, h.L)
	if err != nil {
		return nil, err
	}

	username, _ := conninfo.Get(ctx).Auth()

	qp := &pgdb.QueryParams{
		DB:         params.DB,
		Collection: params.Collection,
		Comment:    params.Comment,
	}

	// get comment from query, e.g. db.collection.find({$comment: "test"})
	if params.Filter != nil {
		if qp.Comment, err = common.GetOptionalParam(params.Filter, "$comment", qp.Comment); err != nil {
			return nil, err
		}
	}

	if !h.DisableFilterPushdown {
		qp.Filter = params.Filter
	}

	if h.EnableSortPushdown && params.Projection == nil {
		qp.Sort = params.Sort
	}

	// Limit pushdown is not applied if:
	//  - `filter` is set, it must fetch all documents to filter them in memory;
	//  - `sort` is set but `EnableSortPushdown` is not set, it must fetch all documents
	//  and sort them in memory;
	//  - `skip` is non-zero value, skip pushdown is not supported yet.
	if params.Filter.Len() == 0 && (params.Sort.Len() == 0 || h.EnableSortPushdown) && params.Skip == 0 {
		qp.Limit = params.Limit
	}

	cancel := func() {}
	if params.MaxTimeMS != 0 {
		// It is not clear if maxTimeMS affects only find, or both find and getMore (as the current code does).
		// TODO https://github.com/FerretDB/FerretDB/issues/2983
		ctx, cancel = context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
	}

	// closer accumulates all things that should be closed / canceled.
	closer := iterator.NewMultiCloser(iterator.CloserFunc(cancel))

	var keepTx pgx.Tx
	var iter types.DocumentsIterator
	err = dbPool.InTransactionKeep(ctx, func(tx pgx.Tx) error {
		keepTx = tx

		var queryRes pgdb.QueryResults
		iter, queryRes, err = pgdb.QueryDocuments(ctx, tx, qp)
		if err != nil {
			return lazyerrors.Error(err)
		}

		closer.Add(iter)

		iter = common.FilterIterator(iter, closer, params.Filter)

		if !queryRes.SortPushdown {
			iter, err = common.SortIterator(iter, closer, params.Sort)
			if err != nil {
				var pathErr *types.PathError
				if errors.As(err, &pathErr) && pathErr.Code() == types.ErrPathElementEmpty {
					return commonerrors.NewCommandErrorMsgWithArgument(
						commonerrors.ErrPathContainsEmptyElement,
						"Empty field names in path are not allowed",
						document.Command(),
					)
				}

				return lazyerrors.Error(err)
			}
		}

		iter = common.SkipIterator(iter, closer, params.Skip)

		iter = common.LimitIterator(iter, closer, params.Limit)

		iter, err = common.ProjectionIterator(iter, closer, params.Projection, params.Filter)
		if err != nil {
			return lazyerrors.Error(err)
		}

		return nil
	})

	if err != nil {
		closer.Close()
		return nil, lazyerrors.Error(err)
	}

	closer.Add(iterator.CloserFunc(func() {
		// It does not matter if we commit or rollback the read transaction,
		// but we should close it.
		// ctx could be canceled already.
		_ = keepTx.Rollback(context.Background())
	}))

	cursor := h.cursors.NewCursor(ctx, &cursor.NewParams{
		Iter:       iterator.WithClose(iterator.Interface[struct{}, *types.Document](iter), closer.Close),
		DB:         params.DB,
		Collection: params.Collection,
		Username:   username,
	})

	cursorID := cursor.ID

	firstBatchDocs, err := iterator.ConsumeValuesN(iterator.Interface[struct{}, *types.Document](cursor), int(params.BatchSize))
	if err != nil {
		cursor.Close()
		return nil, lazyerrors.Error(err)
	}

	firstBatch := types.MakeArray(len(firstBatchDocs))
	for _, doc := range firstBatchDocs {
		firstBatch.Append(doc)
	}

	if params.SingleBatch || firstBatch.Len() < int(params.BatchSize) {
		// Support tailable cursors.
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
				"ns", qp.DB+"."+qp.Collection,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// fetchParams is used to pass parameters to fetchAndFilterDocs.
type fetchParams struct {
	tx                    pgx.Tx
	qp                    *pgdb.QueryParams
	disableFilterPushdown bool
}

// fetchAndFilterDocs fetches documents from the database and filters them using the provided sqlParam.Filter.
func fetchAndFilterDocs(ctx context.Context, fp *fetchParams) ([]*types.Document, error) {
	// filter is used to filter documents on the FerretDB side,
	// qp.Filter is used to filter documents on the PostgreSQL side (query pushdown).
	filter := fp.qp.Filter

	if fp.disableFilterPushdown {
		fp.qp.Filter = nil
	}

	iter, _, err := pgdb.QueryDocuments(ctx, fp.tx, fp.qp)
	if err != nil {
		return nil, err
	}

	closer := iterator.NewMultiCloser(iter)
	defer closer.Close()

	f := common.FilterIterator(iter, closer, filter)

	return iterator.ConsumeValues(iterator.Interface[struct{}, *types.Document](f))
}
