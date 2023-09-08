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

package hana

import (
	"context"
	"time"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/clientconn/cursor"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/hana/hanadb"
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

	qp := hanadb.QueryParams{
		DB:         params.DB,
		Collection: params.Collection,
	}

	cancel := func() {}
	if params.MaxTimeMS != 0 {
		// It is not clear if maxTimeMS affects only find, or both find and getMore (as the current code does).
		// TODO https://github.com/FerretDB/FerretDB/issues/2983
		ctx, cancel = context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
	}

	// closer accumulates all things that should be closed / canceled.
	closer := iterator.NewMultiCloser(iterator.CloserFunc(cancel))

	// Get iterator for documents
	docIter, err := dbPool.QueryDocuments(ctx, &qp)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	closer.Add(docIter)

	docIter = common.FilterIterator(docIter, closer, params.Filter)
	docIter = common.SkipIterator(docIter, closer, params.Skip)

	docIter = common.LimitIterator(docIter, closer, params.Limit)

	docIter, err = common.ProjectionIterator(docIter, closer, params.Projection, params.Filter)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	username, _ := conninfo.Get(ctx).Auth()

	cursor := h.cursors.NewCursor(ctx, &cursor.NewParams{
		Iter:       iterator.WithClose(iterator.Interface[struct{}, *types.Document](docIter), closer.Close),
		DB:         params.DB,
		Collection: params.Collection,
		Username:   username,
	})

	cursorID := cursor.ID

	firstBatchDocs, err := iterator.ConsumeValuesN(iterator.Interface[struct{}, *types.Document](docIter), int(params.BatchSize))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	firstBatch := types.MakeArray(len(firstBatchDocs))
	for _, doc := range firstBatchDocs {
		firstBatch.Append(doc)
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
