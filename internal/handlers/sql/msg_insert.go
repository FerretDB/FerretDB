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

package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgInsert inserts a document or documents into a collection.
func (h *storage) MsgInsert(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	m := document.Map()
	collection := m[document.Command()].(string)
	db := m["$db"].(string)
	docs, _ := m["documents"].(*types.Array)

	ordered, ok := m["ordered"].(bool)
	if !ok {
		ordered = true
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/200
	_ = ordered

	var inserted int32
	for i := 0; i < docs.Len(); i++ {
		doc, err := docs.Get(i)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		d := doc.(types.Document)
		m := d.Map()

		sql := fmt.Sprintf("INSERT INTO %s (", pgx.Identifier{db, collection}.Sanitize())
		var args []any

		for _, k := range d.Keys() {
			// TODO
			if k == "_id" {
				continue
			}

			if len(args) != 0 {
				sql += ", "
			}

			sql += pgx.Identifier{k}.Sanitize()
			args = append(args, m[k])
		}

		sql += ") VALUES ("
		var placeholder pg.Placeholder
		for i := range args {
			if i != 0 {
				sql += ", "
			}
			sql += placeholder.Next()
		}

		sql += ")"

		_, err = h.pgPool.Exec(ctx, sql, args...)
		if err != nil {
			return nil, err
		}

		inserted++
	}

	var res wire.OpMsg
	err = res.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"n", inserted,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &res, nil
}
