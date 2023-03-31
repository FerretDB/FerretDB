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
	"strings"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDataSize implements HandlerInterface.
func (h *Handler) MsgDataSize(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := common.Unimplemented(document, "keyPattern", "min", "max"); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "estimate")

	m := document.Map()
	target, ok := m["dataSize"].(string)
	if !ok {
		return nil, lazyerrors.New("no target collection")
	}
	targets := strings.Split(target, ".")
	if len(targets) != 2 {
		return nil, lazyerrors.New("target collection must be like: 'database.collection'")
	}
	db, collection := targets[0], targets[1]

	started := time.Now()
	stats, err := dbPool.Stats(ctx, db, collection)
	elapses := time.Since(started)

	addEstimate := true
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, lazyerrors.Error(err)
		}

		// return zeroes for non-existent collection
		stats = new(pgdb.DBStats)
		addEstimate = false
	}

	var pairs []any
	if addEstimate {
		pairs = append(pairs, "estimate", false)
	}
	pairs = append(pairs,
		"size", int32(stats.SizeTotal),
		"numObjects", stats.CountRows,
		"millis", int32(elapses.Milliseconds()),
		"ok", float64(1),
	)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(pairs...))},
	}))

	return &reply, nil
}
