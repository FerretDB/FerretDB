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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgCollStats implements HandlerInterface.
func (h *Handler) MsgCollStats(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	db, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := common.GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	// TODO Add proper support for scale: https://github.com/FerretDB/FerretDB/issues/1346
	var scale int32

	scale, err = common.GetOptionalPositiveNumber(document, "scale")
	if err != nil || scale == 0 {
		scale = 1
	}

	stats, err := dbPool.Stats(ctx, db, collection)

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, pgdb.ErrTableNotExist):
		// Return empty stats for non-existent collections.
		stats = new(pgdb.DBStats)
	default:
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ns", db+"."+collection,
			"count", stats.CountRows,
			"size", int32(stats.SizeTotal)/scale,
			"storageSize", int32(stats.SizeRelation)/scale,
			"totalIndexSize", int32(stats.SizeIndexes)/scale,
			"totalSize", int32(stats.SizeTotal)/scale,
			"scaleFactor", scale,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
