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

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/pg"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgUpdate modifies an existing document or documents in a collection.
func (h *storage) MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	m := document.Map()
	collection := m["update"].(string)
	docs, _ := m["updates"].(types.Array)
	db := m["$db"].(string)

	var selected, updated int32
	for _, doc := range docs {
		docM := doc.(types.Document).Map()

		sql := fmt.Sprintf(`SELECT _jsonb FROM %s`, pgx.Identifier{db, collection}.Sanitize())
		var placeholder pg.Placeholder

		whereSQL, args, err := where(docM["q"].(types.Document), &placeholder)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		sql += whereSQL

		rows, err := h.pgPool.Query(ctx, sql, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var updateDocs types.Array

		for {
			updateDoc, err := nextRow(rows)
			if err != nil {
				return nil, err
			}
			if updateDoc == nil {
				break
			}

			updateDocs = append(updateDocs, *updateDoc)
		}

		selected += int32(len(updateDocs))

		for i, updateDoc := range updateDocs {
			d := updateDoc.(types.Document)

			for updateOp, updateV := range docM["u"].(types.Document).Map() {
				switch updateOp {
				case "$set":
					for k, v := range updateV.(types.Document).Map() {
						if err := d.Set(k, v); err != nil {
							return nil, lazyerrors.Error(err)
						}
					}
				default:
					return nil, lazyerrors.Errorf("unhandled operation %q", updateOp)
				}
			}

			updateDocs[i] = d
		}

		for _, updateDoc := range updateDocs {
			sql = fmt.Sprintf("UPDATE %s SET _jsonb = $1 WHERE _jsonb->'_id' = $2", pgx.Identifier{db, collection}.Sanitize())
			d := updateDoc.(types.Document)
			db, err := bson.MustConvertDocument(d).MarshalJSON()
			if err != nil {
				return nil, err
			}

			idb, err := bson.ObjectID(d.Map()["_id"].(types.ObjectID)).MarshalJSON()
			if err != nil {
				return nil, err
			}
			tag, err := h.pgPool.Exec(ctx, sql, db, idb)
			if err != nil {
				return nil, err
			}

			updated += int32(tag.RowsAffected())
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"n", selected,
			"nModified", updated,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
