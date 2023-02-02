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

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
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

	if params.MaxTimeMS != 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
		defer cancel()

		ctx = ctxWithTimeout
	}

	var batchSize int

	// Only apply batchSize if sorting is not set.
	if params.Sort == nil {
		batchSize = int(params.BatchSize)
	}

	sp := pgdb.SQLParam{
		DB:         params.DB,
		Collection: params.Collection,
		Comment:    params.Comment,
		Filter:     params.Filter,
		BatchSize:  batchSize,
		Limit:      int(params.Limit),
	}

	// get comment from query, e.g. db.collection.find({$comment: "test"})
	if sp.Filter != nil {
		if sp.Comment, err = common.GetOptionalParam(sp.Filter, "$comment", sp.Comment); err != nil {
			return nil, err
		}
	}

	if params.Sort == nil {
		var iter iterator.Interface[int, *types.Document]
		var tx pgx.Tx
		var resDocs []*types.Document

		tx, err = dbPool.Begin(ctx)
		if err != nil {
			return nil, err
		}

		resDocs, iter, err = h.getFirstBatchAndIterator(ctx, tx, &sp)
		if err != nil {
			return nil, err
		}

		if iter == nil {
			tx.Commit(ctx)
		}

		if err = common.ProjectDocuments(resDocs, params.Projection); err != nil {
			return nil, err
		}

		firstBatch, id := common.MakeFindReplyParameters(ctx, resDocs, int(params.BatchSize), iter, tx, sp.Filter)

		var reply wire.OpMsg
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"cursor", must.NotFail(types.NewDocument(
					"id", id,
					"ns", sp.DB+"."+sp.Collection,
					"firstBatch", firstBatch,
				)),
				"ok", float64(1),
			))},
		}))

		return &reply, nil
	}

	var resDocs []*types.Document

	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		resDocs, err = h.fetchAndFilterDocs(ctx, tx, &sp)

		return err
	})

	if err != nil {
		return nil, err
	}

	if err = common.SortDocuments(resDocs, params.Sort); err != nil {
		return nil, err
	}

	if resDocs, err = common.LimitDocuments(resDocs, params.Limit); err != nil {
		return nil, err
	}

	if err = common.ProjectDocuments(resDocs, params.Projection); err != nil {
		return nil, err
	}

	firstBatch := types.MakeArray(len(resDocs))
	for _, doc := range resDocs {
		firstBatch.Append(doc)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"id", int64(0),
				"ns", sp.DB+"."+sp.Collection,
				"firstBatch", firstBatch,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

func (h *Handler) fetchAndFilterDocs(ctx context.Context, tx pgx.Tx, sqlParam *pgdb.SQLParam) ([]*types.Document, error) {
	iter, err := pgdb.GetDocuments(ctx, tx, sqlParam)
	if err != nil {
		return nil, err
	}

	defer iter.Close()

	resDocs := make([]*types.Document, 0, 16)

	for {
		_, doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				return resDocs, nil
			}

			return nil, err
		}

		matches, err := common.FilterDocument(doc, sqlParam.Filter)
		if err != nil {
			return nil, err
		}

		if !matches {
			continue
		}

		resDocs = append(resDocs, doc)
	}
}

// getFirstBatchAndIterator fetches documents from the database and filters them using the provided sqlParam.Filter.
func (h *Handler) getFirstBatchAndIterator(ctx context.Context, tx pgx.Tx, sqlParam *pgdb.SQLParam) (
	[]*types.Document, iterator.Interface[int, *types.Document], error,
) {
	iter, err := pgdb.GetDocuments(ctx, tx, sqlParam)
	if err != nil {
		return nil, nil, err
	}

	closeIter := true
	defer func() {
		if closeIter {
			iter.Close()
		}
	}()

	resDocs := make([]*types.Document, 0, 16)

	for i := 0; i < sqlParam.BatchSize; {
		var doc *types.Document

		_, doc, err = iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return nil, nil, err
		}

		matches, err := common.FilterDocument(doc, sqlParam.Filter)
		if err != nil {
			return nil, nil, err
		}

		if !matches {
			continue
		}

		resDocs = append(resDocs, doc)
		i++
	}

	if sqlParam.Limit > 0 && sqlParam.BatchSize > sqlParam.Limit {
		resDocs, err = common.LimitDocuments(resDocs, int64(sqlParam.Limit))
		if err != nil {
			return nil, nil, err
		}

		return resDocs, nil, nil
	}

	if len(resDocs) < sqlParam.BatchSize {
		return resDocs, nil, nil
	}

	closeIter = false

	return resDocs, iter, nil
}
