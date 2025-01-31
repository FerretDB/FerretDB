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

	"github.com/FerretDB/FerretDB/v2/internal/handler/session"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// MsgKillSessions implements `killSessions` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgKillSessions(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	userID, _, err := h.s.CreateOrUpdateByLSID(connCtx, spec)
	if err != nil {
		return nil, err
	}

	ids, err := getSessionIDsParam(doc, doc.Command())
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		// with access control enabled, all other users sessions are killed
		// TODO https://github.com/FerretDB/FerretDB/issues/3974
		cursorIDs := h.s.DeleteSessionsByUserIDs([]session.UserID{userID})

		for _, cursorID := range cursorIDs {
			_ = h.Pool.KillCursor(connCtx, cursorID)
		}

		return wire.MustOpMsg(
			"ok", float64(1),
		), nil
	}

	cursorIDs := h.s.DeleteSessionsByIDs(userID, ids)

	for _, cursorID := range cursorIDs {
		_ = h.Pool.KillCursor(connCtx, cursorID)
	}

	return wire.MustOpMsg(
		"ok", float64(1),
	), nil
}
