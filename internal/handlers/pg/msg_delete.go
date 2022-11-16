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

	if err := common.Unimplemented(document, "let"); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "writeConcern")

	var deletes *types.Array
	if deletes, err = common.GetOptionalParam(document, "deletes", deletes); err != nil {
		return nil, err
	}

	ordered := true
	if ordered, err = common.GetOptionalParam(document, "ordered", ordered); err != nil {
		return nil, err
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
		return nil, common.NewCommandErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
		)
	}

	// get comment from options.Delete().SetComment() method
	if sp.Comment, err = common.GetOptionalParam(document, "comment", sp.Comment); err != nil {
		return nil, err
	}

	var deleted int32
	var delErrors common.WriteErrors

	// process every delete filter
	for i := 0; i < deletes.Len(); i++ {
		// get document with filter
		deleteDoc, err := common.AssertType[*types.Document](must.NotFail(deletes.Get(i)))
		if err != nil {
			return nil, err
		}

		filter, limit, err := h.prepareDeleteParams(deleteDoc)
		if err != nil {
			return nil, err
		}

		// get comment from query, e.g. db.collection.DeleteOne({"_id":"string", "$comment: "test"})
		if sp.Comment, err = common.GetOptionalParam(filter, "$comment", sp.Comment); err != nil {
			return nil, err
		}

		sp.Filter = filter

		del, err := h.execDelete(ctx, &sp, filter, limit)
		if err == nil {
			deleted += del
			continue
		}

		delErrors.Append(err, int32(i))

		if ordered {
			break
		}
	}

	replyDoc := must.NotFail(types.NewDocument(
		"ok", float64(1),
	))

	if delErrors.Len() > 0 {
		replyDoc = delErrors.Document()
	}

	replyDoc.Set("n", deleted)

	var reply wire.OpMsg

	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{replyDoc},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// prepareDeleteParams extracts query filter and limit from delete document.
func (h *Handler) prepareDeleteParams(deleteDoc *types.Document) (*types.Document, int64, error) {
	var err error

	if err = common.Unimplemented(deleteDoc, "collation", "hint"); err != nil {
		return nil, 0, err
	}

	// get filter from document
	var filter *types.Document
	if filter, err = common.GetOptionalParam(deleteDoc, "q", filter); err != nil {
		return nil, 0, err
	}

	l, err := deleteDoc.Get("limit")
	if err != nil {
		return nil, 0, common.NewCommandErrorMsgWithArgument(
			common.ErrMissingField,
			"BSON field 'delete.deletes.limit' is missing but a required field",
			"limit",
		)
	}

	var limit int64
	if limit, err = common.GetWholeNumberParam(l); err != nil || limit < 0 || limit > 1 {
		return nil, 0, common.NewCommandErrorMsgWithArgument(
			common.ErrFailedToParse,
			fmt.Sprintf("The limit field in delete objects must be 0 or 1. Got %v", l),
			"limit",
		)
	}

	return filter, limit, nil
}

// execDelete fetches documents, filter them out and limiting with the given limit value.
// It returns the number of deleted documents or an error.
func (h *Handler) execDelete(ctx context.Context, sp *pgdb.SQLParam, filter *types.Document, limit int64) (int32, error) {
	var err error

	resDocs := make([]*types.Document, 0, 16)

	var deleted int32
	err = h.PgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		// fetch current items from collection
		fetchedChan, err := h.PgPool.QueryDocuments(ctx, tx, sp)
		if err != nil {
			return err
		}
		defer func() {
			// Drain the channel to prevent leaking goroutines.
			// TODO Offer a better design instead of channels: https://github.com/FerretDB/FerretDB/issues/898.
			for range fetchedChan {
			}
		}()

		// iterate through every row and delete matching ones
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

			// if no field is matched in a row, go to the next one
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
		return 0, err
	}

	return deleted, nil
}

// delete deletes documents by _id.
func (h *Handler) delete(ctx context.Context, sp *pgdb.SQLParam, docs []*types.Document) (int64, error) {
	ids := make([]any, len(docs))
	for i, doc := range docs {
		id := must.NotFail(doc.Get("_id"))
		ids[i] = id
	}

	var rowsDeleted int64
	err := h.PgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		var err error
		rowsDeleted, err = pgdb.DeleteDocumentsByID(ctx, tx, sp, ids)
		return err
	})
	if err != nil {
		// TODO check error code
		return 0, common.NewCommandError(common.ErrNamespaceNotFound, fmt.Errorf("delete: ns not found: %w", err))
	}

	return rowsDeleted, nil
}
