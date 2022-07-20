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

package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDelete implements HandlerInterface.
func (h *Handler) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.l, "comment") // TODO https://github.com/FerretDB/FerretDB/issues/849
	if err := common.Unimplemented(document, "let"); err != nil {
		return nil, err
	}
	common.Ignored(document, h.l, "ordered") // TODO https://github.com/FerretDB/FerretDB/issues/848
	common.Ignored(document, h.l, "writeConcern")

	var deletes *types.Array
	if deletes, err = common.GetOptionalParam(document, "deletes", deletes); err != nil {
		return nil, err
	}

	var deleted int32
	for i := 0; i < deletes.Len(); i++ {
		d, err := common.AssertType[*types.Document](must.NotFail(deletes.Get(i)))
		if err != nil {
			return nil, err
		}

		if err := common.Unimplemented(d, "collation", "hint"); err != nil {
			return nil, err
		}

		var filter *types.Document
		if filter, err = common.GetOptionalParam(d, "q", filter); err != nil {
			return nil, err
		}

		var limit int64
		if l, _ := d.Get("limit"); l != nil {
			if limit, err = common.GetWholeNumberParam(l); err != nil {
				return nil, err
			}
		}

		var sp pgdb.SQLParam
		if sp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
			return nil, err
		}
		collectionParam, err := document.Get(document.Command())
		if err != nil {
			return nil, err
		}
		var ok bool
		if sp.Collection, ok = collectionParam.(string); !ok {
			return nil, common.NewErrorMsg(
				common.ErrBadValue,
				fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			)
		}

		resDocs := make([]*types.Document, 0, 16)
		err = h.pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
			fetchedChan, err := h.pgPool.QueryDocuments(ctx, tx, sp)
			if err != nil {
				return err
			}
			defer func() {
				// Drain the channel to prevent leaking goroutines.
				// TODO Offer a better design instead of channels: https://github.com/FerretDB/FerretDB/issues/898.
				for range fetchedChan {
				}
			}()

			for fetchedItem := range fetchedChan {
				if fetchedItem.Err != nil {
					return fetchedItem.Err
				}

				for _, doc := range fetchedItem.Docs {
					matches, err := common.FilterDocument(doc, filter)
					if err != nil {
						return err
					}

					if !matches {
						continue
					}

					resDocs = append(resDocs, doc)
				}

				if resDocs, err = common.LimitDocuments(resDocs, limit); err != nil {
					return err
				}

				if len(resDocs) == 0 {
					continue
				}

				rowsDeleted, err := h.delete(ctx, sp, resDocs)
				if err != nil {
					return err
				}

				deleted += int32(rowsDeleted)
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"n", deleted,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// delete deletes documents by _id.
func (h *Handler) delete(ctx context.Context, sp pgdb.SQLParam, docs []*types.Document) (int64, error) {
	ids := make([]any, len(docs))
	for i, doc := range docs {
		id := must.NotFail(doc.Get("_id"))
		ids[i] = id
	}

	rowsDeleted, err := h.pgPool.DeleteDocumentsByID(ctx, sp.DB, sp.Collection, ids)
	if err != nil {
		// TODO check error code
		return 0, common.NewError(common.ErrNamespaceNotFound, fmt.Errorf("delete: ns not found: %w", err))
	}
	return rowsDeleted, nil
}
