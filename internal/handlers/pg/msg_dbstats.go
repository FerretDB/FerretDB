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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
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

	db, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	scale, err := common.GetScaleParam("collStats", document)
	if err != nil {
		return nil, err
	}

	stats, err := dbPool.Stats(ctx, db, "")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var avgObjSize float64
	if stats.CountRows > 0 {
		avgObjSize = float64(stats.SizeRelation) / float64(stats.CountRows)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"db", db,
			"collections", stats.CountTables,
			// TODO https://github.com/FerretDB/FerretDB/issues/176
			"views", int32(0),
			"objects", stats.CountRows,
			"avgObjSize", avgObjSize,
			"dataSize", float64(stats.SizeRelation/int64(scale)),
			"indexes", stats.CountIndexes,
			"indexSize", float64(stats.SizeIndexes/int64(scale)),
			"totalSize", float64(stats.SizeTotal/int64(scale)),
			"scaleFactor", float64(scale),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
