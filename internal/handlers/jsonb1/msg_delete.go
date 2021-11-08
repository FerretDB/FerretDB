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

package jsonb1

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"

	"github.com/MangoDB-io/MangoDB/internal/handlers/common"
	"github.com/MangoDB-io/MangoDB/internal/pg"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

func (h *storage) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, common.NewError(common.ErrInternalError, err)
	}

	m := document.Map()
	collection := m[document.Command()].(string)
	db := m["$db"].(string)
	docs, _ := m["deletes"].(types.Array)

	var deleted int32
	for _, doc := range docs {
		d := doc.(types.Document).Map()

		sql := fmt.Sprintf(`DELETE FROM %s`, pgx.Identifier{db, collection}.Sanitize())
		var placeholder pg.Placeholder

		elSQL, args, err := where(d["q"].(types.Document), &placeholder)
		if err != nil {
			return nil, common.NewError(common.ErrNotImplemented, err)
		}

		// TODO
		limit, _ := d["limit"].(int32)
		_ = limit
		// if limit != 0 {
		// 	return nil, common.NewError(common.ErrNotImplemented, fmt.Errorf("limit for delete is not supported"), header, msg)
		// }

		sql += elSQL

		tag, err := h.pgPool.Exec(ctx, sql, args...)
		if err != nil {
			// TODO check error code
			return nil, common.NewError(common.ErrNamespaceNotFound, fmt.Errorf("ns not found"))
		}

		deleted += int32(tag.RowsAffected())
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"n", deleted,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, common.NewError(common.ErrInternalError, err)
	}

	return &reply, nil
}
