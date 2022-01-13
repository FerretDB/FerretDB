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

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListCollections retrieves information (i.e. the name and options)
// about the collections and views in a database.
func (h *Handler) MsgListCollections(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	m := document.Map()

	filter, ok := m["filter"].(types.Document)
	if ok && len(filter.Map()) != 0 {
		return nil, common.NewErrorMessage(common.ErrNotImplemented, "MsgListCollections: filter is not supported")
	}

	cursor, ok := m["cursor"].(types.Document)
	if ok && len(cursor.Map()) != 0 {
		return nil, common.NewErrorMessage(common.ErrNotImplemented, "MsgListCollections: cursor is not supported")
	}

	nameOnly, ok := m["nameOnly"].(bool)
	if ok && !nameOnly {
		return nil, common.NewErrorMessage(common.ErrNotImplemented, "MsgListCollections: nameOnly=false is not supported")
	}

	db, ok := m["$db"].(string)
	if !ok {
		return nil, lazyerrors.New("no db")
	}

	names, err := h.pgPool.Tables(ctx, db)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	collections := types.MakeArray(len(names))
	for _, n := range names {
		d := types.MustMakeDocument(
			"name", n,
			"type", "collection",
		)
		if err = collections.Append(d); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"cursor", types.MustMakeDocument(
				"id", int64(0),
				"ns", db+".$cmd.listCollections",
				"firstBatch", collections,
			),
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
