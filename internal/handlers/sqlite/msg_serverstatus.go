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

package sqlite

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgServerStatus implements HandlerInterface.
func (h *Handler) MsgServerStatus(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	res, err := common.ServerStatus(h.StateProvider.Get(), h.ConnMetrics)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	stats, err := h.b.Stats(ctx, new(backends.ServerStatsParams))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res.Set("catalogStats", must.NotFail(types.NewDocument(
		"collections", stats.CountCollections,
		"capped", int32(0),
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
