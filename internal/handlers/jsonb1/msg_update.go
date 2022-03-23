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

	"github.com/FerretDB/FerretDB/internal/fjson"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgUpdate modifies an existing document or documents in a collection.
func (s *storage) MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err := common.Unimplemented(document, "let"); err != nil {
		return nil, err
	}
	common.Ignored(document, s.l, "ordered", "writeConcern", "bypassDocumentValidation", "comment")

	command := document.Command()

	var db, collection string
	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}
	if collection, err = common.GetRequiredParam[string](document, command); err != nil {
		return nil, err
	}

	var updates *types.Array
	if updates, err = common.GetOptionalParam(document, "updates", updates); err != nil {
		return nil, err
	}

	var selected, updated int32
	for i := 0; i < updates.Len(); i++ {
		update, err := common.AssertType[*types.Document](must.NotFail(updates.Get(i)))
		if err != nil {
			return nil, err
		}

		unimplementedFields := []string{
			"c",
			"upsert",
			"multi",
			"collation",
			"arrayFilters",
			"hint",
		}
		if err := common.Unimplemented(update, unimplementedFields...); err != nil {
			return nil, err
		}

		var q, u *types.Document
		if q, err = common.GetOptionalParam(update, "q", q); err != nil {
			return nil, err
		}
		if u, err = common.GetOptionalParam(update, "u", u); err != nil {
			return nil, err
		}

		fetchedDocs, err := s.fetch(ctx, db, collection)
		if err != nil {
			return nil, err
		}

		resDocs := make([]*types.Document, 0, 16)
		for _, doc := range fetchedDocs {
			matches, err := common.FilterDocument(doc, q)
			if err != nil {
				return nil, err
			}

			if !matches {
				continue
			}

			resDocs = append(resDocs, doc)
		}

		if len(resDocs) == 0 {
			continue
		}

		selected += int32(len(resDocs))

		for _, doc := range resDocs {
			for _, updateOp := range u.Keys() {
				updateV := must.NotFail(u.Get(updateOp))
				switch updateOp {
				case "$set":
					setDoc, err := common.AssertType[*types.Document](updateV)
					if err != nil {
						return nil, err
					}

					for _, setKey := range setDoc.Keys() {
						setValue := must.NotFail(setDoc.Get(setKey))
						if err = doc.Set(setKey, setValue); err != nil {
							return nil, lazyerrors.Error(err)
						}
					}

				default:
					return nil, lazyerrors.Errorf("unhandled operation %q", updateOp)
				}
			}

			sql := fmt.Sprintf("UPDATE %s SET _jsonb = $1 WHERE _jsonb->'_id' = $2", pgx.Identifier{db, collection}.Sanitize())
			id := must.NotFail(doc.Get("_id"))
			tag, err := s.pgPool.Exec(ctx, sql, must.NotFail(fjson.Marshal(doc)), must.NotFail(fjson.Marshal(id)))
			if err != nil {
				return nil, err
			}

			updated += int32(tag.RowsAffected())
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
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
