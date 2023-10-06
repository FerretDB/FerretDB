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

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgServerStatus implements HandlerInterface.
func (h *Handler) MsgServerStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res, err := common.ServerStatus(h.StateProvider.Get(), h.ConnMetrics)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var stats *pgdb.ServerStats

	if err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		stats, err = pgdb.CalculateServerStats(ctx, tx)
		return err
	}); err != nil {
		return nil, lazyerrors.Error(err)
	}

	res.Set("catalogStats", must.NotFail(types.NewDocument(
		"collections", stats.CountCollections,
		"capped", int32(0), // TODO https://github.com/FerretDB/FerretDB/issues/2342
		"clustered", int32(0),
		"timeseries", int32(0),
		"views", int32(0),
		"internalCollections", int32(0),
		"internalViews", int32(0),
	)))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{res},
	}))

	return &reply, nil
}
