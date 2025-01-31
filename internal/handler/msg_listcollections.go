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
	"sort"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// MsgListCollections implements `listCollections` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgListCollections(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
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

	userID, sessionID, err := h.s.CreateOrUpdateByLSID(connCtx, spec)
	if err != nil {
		return nil, err
	}

	page, cursorID, err := h.Pool.ListCollections(connCtx, dbName, spec)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	h.s.AddCursor(connCtx, userID, sessionID, cursorID)

	// Sort the first page as a partial workaround for
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/822

	resp, err := page.DecodeDeep()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	cursor := resp.Get("cursor").(*wirebson.Document)
	firstBatch := cursor.Get("firstBatch").(*wirebson.Array)

	sort.Sort(firstBatch.SortInterface(func(a, b any) bool {
		an := a.(*wirebson.Document).Get("name").(string)
		bn := b.(*wirebson.Document).Get("name").(string)

		return an < bn
	}))

	page, err = resp.Encode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if msg, err = wire.NewOpMsg(page); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return msg, nil
}
