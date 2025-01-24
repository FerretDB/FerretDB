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

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// MsgKillAllSessions implements `killAllSessions` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgKillAllSessions(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, _, err = h.s.CreateOrUpdateByLSID(connCtx, spec); err != nil {
		return nil, err
	}

	command := doc.Command()
	field := "KillAllSessionsCmd.killAllSessions"

	v := doc.Get(command)

	userIDs, err := getSessionUsersParam(v, command, field)
	if err != nil {
		return nil, err
	}

	if len(userIDs) == 0 {
		cursorIDs := h.s.DeleteAllSessions()

		for _, cursorID := range cursorIDs {
			_ = h.Pool.KillCursor(connCtx, cursorID)
		}

		return wire.MustOpMsg(
			"ok", float64(1),
		), nil
	}

	cursorIDs := h.s.DeleteSessionsByUserIDs(userIDs)

	for _, cursorID := range cursorIDs {
		_ = h.Pool.KillCursor(connCtx, cursorID)
	}

	return wire.MustOpMsg(
		"ok", float64(1),
	), nil
}
