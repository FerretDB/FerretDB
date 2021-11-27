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

package jsonb1

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgFind selects documents in a collection or view and returns a cursor to the selected documents.
func (h *storage) MsgFind(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	m := document.Map()
	collection := m["find"].(string)
	db := m["$db"].(string)

	projection, ok := m["projection"].(types.Document)
	if ok && len(projection.Map()) != 0 {
		return nil, common.NewErrorMessage(common.ErrNotImplemented, "MsgFind: projection is not supported")
	}

	filter, _ := m["filter"].(types.Document)
	sort, _ := m["sort"].(types.Document)
	limit, _ := m["limit"].(int32)

	sql := fmt.Sprintf(`SELECT _jsonb FROM %s`, pgx.Identifier{db, collection}.Sanitize())
	var args []interface{}
	var placeholder pg.Placeholder

	whereSQL, args, err := where(filter, &placeholder)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	sql += whereSQL

	sortMap := sort.Map()
	if len(sortMap) > 0 {
		sql += " ORDER BY"

		for i, k := range sort.Keys() {
			if i != 0 {
				sql += ","
			}

			sql += " _jsonb->" + placeholder.Next()
			args = append(args, k)

			order := sortMap[k].(int32)
			if order > 0 {
				sql += " ASC"
			} else {
				sql += " DESC"
			}
		}
	}

	switch {
	case limit == 0:
		// undefined or zero - no limit
	case limit > 0:
		sql += " LIMIT " + placeholder.Next()
		args = append(args, limit)
	default:
		// TODO https://github.com/FerretDB/FerretDB/issues/79
		return nil, common.NewErrorMessage(common.ErrNotImplemented, "MsgFind: negative limit values are not supported")
	}

	rows, err := h.pgPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	var docs types.Array

	for {
		doc, err := nextRow(rows)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		if doc == nil {
			break
		}

		docs = append(docs, *doc)
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"cursor", types.MustMakeDocument(
				"firstBatch", docs,
				"id", int64(0), // TODO
				"ns", db+"."+collection,
			),
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
