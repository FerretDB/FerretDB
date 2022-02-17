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

package handlers

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

func formatResponse(size, rows, millis int32, showEstimate bool) (*wire.OpMsg, error) {
	var pairs []any
	if showEstimate {
		pairs = append(pairs, "estimate", false)
	}
	pairs = append(pairs,
		"size", size,
		"numObjects", rows,
		"millis", millis,
		"ok", float64(1),
	)

	var reply wire.OpMsg
	err := reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(pairs...)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// MsgDataSize returns the size of the collection in bytes.
func (h *Handler) MsgDataSize(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := common.Unimplemented(document, "keyPattern", "min", "max"); err != nil {
		return nil, err
	}
	common.Ignored(document, h.l, "estimate")

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
	stats, err := h.pgPool.TableStats(ctx, db, collection)
	elapses := time.Since(started)
	millis := int32(elapses.Milliseconds())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return formatResponse(0, 0, millis, false)
		}
		return nil, lazyerrors.Error(err)
	}

	return formatResponse(stats.SizeTotal, stats.Rows, millis, true)
}
