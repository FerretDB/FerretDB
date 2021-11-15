// Copyright 2021 Baltoro OÜ.
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

package shared

import (
	"context"
	"fmt"
	"sort"

	"github.com/MangoDB-io/MangoDB/internal/handlers/common"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

func (h *Handler) MsgListCollections(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, common.NewError(common.ErrInternalError, err)
	}

	m := document.Map()

	filter, ok := m["filter"].(types.Document)
	if ok && len(filter.Map()) != 0 {
		return nil, common.NewError(common.ErrNotImplemented, fmt.Errorf("filter is not supported"))
	}

	cursor, ok := m["cursor"].(types.Document)
	if ok && len(cursor.Map()) != 0 {
		return nil, common.NewError(common.ErrNotImplemented, fmt.Errorf("cursor is not supported"))
	}

	nameOnly, ok := m["nameOnly"].(bool)
	if ok && !nameOnly {
		return nil, common.NewError(common.ErrNotImplemented, fmt.Errorf("nameOnly=false is not supported"))
	}

	db, ok := m["$db"].(string)
	if !ok {
		return nil, common.NewError(common.ErrInternalError, fmt.Errorf("no db"))
	}

	var names []string
	rows, err := h.pgPool.Query(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = $1", db)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, err
		}

		names = append(names, name)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	sort.Strings(names)

	collections := make(types.Array, len(names))
	for i, n := range names {
		collections[i] = types.MustMakeDocument(
			"name", n,
			"type", "collection",
		)
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
		return nil, common.NewError(common.ErrInternalError, err)
	}

	return &reply, nil
}
