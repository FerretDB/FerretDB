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
	"errors"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
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

	if params.MaxTimeMS != 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
		defer cancel()

		ctx = ctxWithTimeout
	}

	qp := &tigrisdb.QueryParams{
		DB:         params.DB,
		Collection: params.Collection,
		Filter:     params.Filter,
	}

	if !h.DisablePushdown {
		qp.Filter = params.Filter
	}

	var resDocs []*types.Document

	err = func() error {
		if params.BatchSize == 0 {
			return nil
		}

		var iter types.DocumentsIterator

		if iter, err = dbPool.QueryDocuments(ctx, qp); err != nil {
			return lazyerrors.Error(err)
		}

		defer iter.Close()

		iter = common.FilterIterator(iter, params.Filter)

		iter, err = common.SortIterator(iter, params.Sort)
		if err != nil {
			var pathErr *types.DocumentPathError
			if errors.As(err, &pathErr) && pathErr.Code() == types.ErrDocumentPathEmptyKey {
				return commonerrors.NewCommandErrorMsgWithArgument(
					commonerrors.ErrPathContainsEmptyElement,
					"Empty field names in path are not allowed",
					document.Command(),
				)
			}

			return lazyerrors.Error(err)
		}

		// SortIterator should be closed
		defer iter.Close()

		iter = common.LimitIterator(iter, params.Limit)

		iter = common.SkipIterator(iter, params.Skip)

		if iter, err = common.ProjectionIterator(iter, params.Projection); err != nil {
			return lazyerrors.Error(err)
		}

		resDocs, err = iterator.ConsumeValues(iterator.Interface[struct{}, *types.Document](iter))

		return err
	}()

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	firstBatch := types.MakeArray(len(resDocs))
	for _, doc := range resDocs {
		firstBatch.Append(doc)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"firstBatch", firstBatch,
				"id", int64(0), // TODO
				"ns", qp.DB+"."+qp.Collection,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// fetchParams is used to pass parameters to fetchAndFilterDocs.
type fetchParams struct {
	dbPool          *tigrisdb.TigrisDB
	qp              *tigrisdb.QueryParams
	disablePushdown bool
}

// fetchAndFilterDocs fetches documents from the database and filters them using the provided QueryParams.Filter.
func fetchAndFilterDocs(ctx context.Context, fp *fetchParams) ([]*types.Document, error) {
	// filter is used to filter documents on the FerretDB side,
	// qp.Filter is used to filter documents on the Tigris side (query pushdown).
	filter := fp.qp.Filter

	if fp.disablePushdown {
		fp.qp.Filter = nil
	}

	iter, err := fp.dbPool.QueryDocuments(ctx, fp.qp)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer iter.Close()

	f := common.FilterIterator(iter, filter)

	return iterator.ConsumeValues(iterator.Interface[struct{}, *types.Document](f))
}
