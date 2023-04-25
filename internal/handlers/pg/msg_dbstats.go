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
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDBStats implements HandlerInterface.
func (h *Handler) MsgDBStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	db, err := commonparams.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	scale := int32(1)

	var s any
	if s, err = document.Get("scale"); err == nil {
		if scale, err = common.GetScaleParam(command, s); err != nil {
			return nil, err
		}
	}

	var stats *pgdb.DBStats

	if err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		stats, err = pgdb.CalculateDBStats(ctx, tx, db)
		return err
	}); err != nil {
		return nil, lazyerrors.Error(err)
	}

	pairs := []any{
		"db", db,
		"collections", stats.CountCollections,
		// TODO https://github.com/FerretDB/FerretDB/issues/176
		"views", int32(0),
		"objects", stats.CountObjects,
	}

	if stats.CountObjects > 0 {
		pairs = append(pairs, "avgObjSize", stats.SizeCollections/stats.CountObjects)
	}

	pairs = append(pairs,
		"dataSize", stats.SizeCollections/scale,
		"storageSize", stats.SizeCollections/scale,
		"indexes", stats.CountIndexes,
		"indexSize", stats.SizeIndexes/scale,
		"totalSize", stats.SizeTotal/scale,
		"scaleFactor", float64(scale),
		"ok", float64(1),
	)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(pairs...))},
	}))

	return &reply, nil
}
