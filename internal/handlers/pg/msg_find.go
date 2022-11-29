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
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
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

	if params.MaxTimeMS != 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(params.MaxTimeMS)*time.Millisecond)
		defer cancel()

		ctx = ctxWithTimeout
	}

	sp := pgdb.SQLParam{
		DB:         params.DB,
		Collection: params.Collection,
		Comment:    params.Comment,
		Filter:     params.Filter,
	}

	// get comment from query, e.g. db.collection.find({$comment: "test"})
	if sp.Filter != nil {
		if sp.Comment, err = common.GetOptionalParam(sp.Filter, "$comment", sp.Comment); err != nil {
			return nil, err
		}
	}

	resDocs := make([]*types.Document, 0, 16)
	err = h.PgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		fetchedChan, err := h.PgPool.QueryDocuments(ctx, tx, &sp)
		if err != nil {
			return err
		}
		defer func() {
			// Drain the channel to prevent leaking goroutines.
			// TODO Offer a better design instead of channels: https://github.com/FerretDB/FerretDB/issues/898.
			for range fetchedChan {
			}
		}()

		for fetchedItem := range fetchedChan {
			if fetchedItem.Err != nil {
				return fetchedItem.Err
			}

			for _, doc := range fetchedItem.Docs {
				matches, err := common.FilterDocument(doc, params.Filter)
				if err != nil {
					return err
				}

				if !matches {
					continue
				}

				resDocs = append(resDocs, doc)
			}
		}

		return nil
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
		if err = firstBatch.Append(doc); err != nil {
			return nil, err
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"firstBatch", firstBatch,
				"id", int64(0), // TODO
				"ns", sp.DB+"."+sp.Collection,
			)),
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
