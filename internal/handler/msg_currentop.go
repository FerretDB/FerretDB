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

package handler

import (
	"context"
	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// MsgCurrentOp implements `currentOp` command.
//
// The passed context is canceled when the client connection is closed.
//
// TODO https://github.com/FerretDB/FerretDB/issues/3974
func (h *Handler) MsgCurrentOp(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, _, err = h.s.CreateOrUpdateByLSID(connCtx, spec); err != nil {
		return nil, err
	}

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/78
	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	currentOpSpec := must.NotFail(wirebson.MustDocument(
		"aggregate", int32(1),
		"pipeline", wirebson.MustArray(wirebson.MustDocument("$currentOp", wirebson.MustDocument())),
		// use large batchSize to get all results in one batch
		"cursor", wirebson.MustDocument("batchSize", int32(10000)),
	).Encode())

	page, cursorID, err := h.Pool.Aggregate(connCtx, dbName, currentOpSpec)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if cursorID != 0 {
		h.L.WarnContext(connCtx, "MsgCurrentOp: too many in-progress operations; not all operations are shown")

		_ = h.Pool.KillCursor(connCtx, cursorID)
	}

	res, err := page.DecodeDeep()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	ops := res.Get("cursor").(*wirebson.Document).Get("firstBatch").(*wirebson.Array)

	return wire.MustOpMsg(
		"inprog", ops,
		"ok", float64(1),
	), nil
}
