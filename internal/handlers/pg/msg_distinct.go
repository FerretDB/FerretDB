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

	qp := pgdb.QueryParam{
		DB:              dp.DB,
		Collection:      dp.Collection,
		Filter:          dp.Filter,
		Comment:         dp.Comment,
		DisablePushdown: h.DisablePushdown,
	}

	var resDocs []*types.Document
	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		resDocs, err = h.fetchAndFilterDocs(ctx, tx, &qp)
		return err
	})

	if err != nil {
		return nil, err
	}

	distinct, err := common.FilterDistinctValues(resDocs, dp.Key)
	if err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"values", distinct,
			"ok", float64(1),
		))},
	}))

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
