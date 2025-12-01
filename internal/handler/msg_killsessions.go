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

	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/handler/session"
)

// msgKillSessions implements `killSessions` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgKillSessions(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	userID, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc)
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
			_ = h.p.KillCursor(connCtx, cursorID)
		}

		return middleware.ResponseDoc(req, wirebson.MustDocument(
			"ok", float64(1),
		))
	}

	cursorIDs := h.s.DeleteSessionsByIDs(userID, ids)

	for _, cursorID := range cursorIDs {
		_ = h.p.KillCursor(connCtx, cursorID)
	}

	return middleware.ResponseDoc(req, wirebson.MustDocument(
		"ok", float64(1),
	))
}
