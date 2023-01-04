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

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDistinct implements HandlerInterface.
func (h *Handler) MsgDistinct(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dp, err := common.GetDistinctParams(document, h.L)
	if err != nil {
		return nil, err
	}

	sp := pgdb.SQLParam{
		DB:         dp.DB,
		Collection: dp.Collection,
		Filter:     dp.Filter,
		Comment:    dp.Comment,
	}

	resDocs := make([]*types.Document, 0, 16)
	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		resDocs, err = h.fetchAndFilterDocs(ctx, tx, &sp)
		return err
	})

	if err != nil {
		return nil, err
	}

	distinct := types.MakeArray(len(resDocs))

	for _, doc := range resDocs {
		var val any

		val, err = doc.Get(dp.Key)
		if err != nil {
			// if the key is not found in the current document, it should be skipped
			continue
		}

		switch v := val.(type) {
		case *types.Array:
			for i := 0; i < v.Len(); i++ {
				el, err := v.Get(i)
				if err != nil {
					return nil, lazyerrors.Error(err)
				}

				if !distinct.Contains(el) {
					err := distinct.Append(el)
					if err != nil {
						return nil, lazyerrors.Error(err)
					}
				}
			}
		default:
			if !distinct.Contains(v) {
				err := distinct.Append(v)
				if err != nil {
					return nil, lazyerrors.Error(err)
				}
			}
		}
	}

	if err = common.SortArray(distinct, types.Ascending); err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"values", distinct,
			"ok", float64(1),
		))},
	})

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
