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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonparams"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
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

	db, err := commonparams.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := commonparams.GetRequiredParam[string](document, command)
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

	querier := dbPool.Driver.UseDatabase(db)

	stats, err := tigrisdb.FetchStats(ctx, querier, collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	pairs := []any{
		"ns", db + "." + collection,
		"size", int32(stats.Size) / scale,
		"count", stats.NumObjects,
	}

	// If there are objects in the collection, calculate the average object size.
	if stats.NumObjects > 0 {
		pairs = append(pairs, "avgObjSize", int32(stats.Size)/stats.NumObjects)
	}

	pairs = append(pairs,
		"storageSize", int32(stats.Size)/scale,
		"nindexes", int32(0),
		"totalIndexSize", int32(0),
		"totalSize", int32(stats.Size)/scale,
		"scaleFactor", scale,
		"ok", float64(1),
	)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(pairs...))},
	}))

	return &reply, nil
}
