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
	common.Ignored(document, h.l, "writeConcern")

	var deletes *types.Array
	if deletes, err = common.GetOptionalParam(document, "deletes", deletes); err != nil {
		return nil, err
	}

	ordered := true
	if ordered, err = common.GetOptionalParam(document, "ordered", ordered); err != nil {
		return nil, err
	}

	var deleted int32
	processQuery := func(i int) error {
		// get document with filter
		d, err := common.AssertType[*types.Document](must.NotFail(deletes.Get(i)))
		if err != nil {
			return err
		}

		if err := common.Unimplemented(d, "collation", "hint"); err != nil {
			return err
		}

		// get filter from document
		var filter *types.Document
		if filter, err = common.GetOptionalParam(d, "q", filter); err != nil {
			return err
		}

		var limit int64

		l, err := d.Get("limit")
		if err != nil {
			return common.NewErrorMsg(
				common.ErrMissingField,
				"BSON field 'delete.deletes.limit' is missing but a required field",
			)
		}

		if limit, err = common.GetWholeNumberParam(l); err != nil || limit < 0 || limit > 1 {
			return common.NewErrorMsg(
				common.ErrFailedToParse,
				fmt.Sprintf("The limit field in delete objects must be 0 or 1. Got %v", l),
			)
		}

		var sp pgdb.SQLParam
		if sp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
			return err
		}
		collectionParam, err := document.Get(document.Command())
		if err != nil {
			return err
		}
		var ok bool
		if sp.Collection, ok = collectionParam.(string); !ok {
			return common.NewErrorMsg(
				common.ErrBadValue,
				fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			)
		}

		// get comment from options.Delete().SetComment() method
		if sp.Comment, err = common.GetOptionalParam(document, "comment", sp.Comment); err != nil {
			return err
		}

		// get comment from query, e.g. db.collection.DeleteOne({"_id":"string", "$comment: "test"})
		if sp.Comment, err = common.GetOptionalParam(filter, "$comment", sp.Comment); err != nil {
			return err
		}

		resDocs := make([]*types.Document, 0, 16)

		return h.pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
			// fetch current items from collection
			fetchedChan, err := h.pgPool.QueryDocuments(ctx, tx, &sp)
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

				rowsDeleted, err := h.delete(ctx, &sp, resDocs)
				if err != nil {
					return err
				}

				deleted += int32(rowsDeleted)
			}

			return nil
		})
	}

	var delErrors common.WriteErrors

	// process every delete filter
	for i := 0; i < deletes.Len(); i++ {
		err := processQuery(i)
		switch err.(type) {
		case nil:
			continue

		case *common.CommandError:
			// command errors should be return immediately
			return nil, err

		default:
			// write errors and others require to be handled in array
			delErrors.Append(err, int32(i))

			// Delete statements in the `deletes` field are not transactional.
			// It means that we run each delete statement separately.
			// If `ordered` is set as `true`, we don't execute the remaining statements
			// after the first failure.
			// If `ordered` is set as `false`, we execute all the statements and return
			// the list of errors corresponding to the failed statements.
			if !ordered {
				continue
			}
		}

		// send response if ordered is true
		break
	}

	replyDoc := must.NotFail(types.NewDocument(
		"ok", float64(1),
	))

	if len(delErrors) > 0 {
		replyDoc = delErrors.Document()
	}

	must.NoError(replyDoc.Set("n", deleted))

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{replyDoc},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// delete deletes documents by _id.
func (h *Handler) delete(ctx context.Context, sp *pgdb.SQLParam, docs []*types.Document) (int64, error) {
	ids := make([]any, len(docs))
	for i, doc := range docs {
		id := must.NotFail(doc.Get("_id"))
		ids[i] = id
	}

	rowsDeleted, err := h.pgPool.DeleteDocumentsByID(ctx, sp, ids)
	if err != nil {
		// TODO check error code
		return 0, common.NewError(common.ErrNamespaceNotFound, fmt.Errorf("delete: ns not found: %w", err))
	}
	return rowsDeleted, nil
}
