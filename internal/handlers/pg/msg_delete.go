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
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDelete implements HandlerInterface.
func (h *Handler) MsgDelete(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

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
		return nil, common.NewCommandErrorMsgWithArgument(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			document.Command(),
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

		var limited bool
		if limit == 1 {
			limited = true
		}

		del, err := execDelete(ctx, dbPool, &sp, limited)
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

// execDelete fetches documents, filters them out, limits them (if needed) and deletes them.
// If limit is true, only the first matched document is chosen for deletion, otherwise all matched documents are chosen.
// It returns the number of deleted documents or an error.
func execDelete(ctx context.Context, dbPool *pgdb.Pool, sp *pgdb.SQLParam, limit bool) (int32, error) {
	var deleted int32
	err := dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		iter, err := pgdb.GetDocuments(ctx, tx, sp)
		if err != nil {
			return err
		}

		defer iter.Close()

		resDocs := make([]*types.Document, 0, 16)

		for {
			var doc *types.Document
			if _, doc, err = iter.Next(); err != nil {
				if errors.Is(err, iterator.ErrIteratorDone) {
					break
				}

				return err
			}

			var matches bool
			if matches, err = common.FilterDocument(doc, sp.Filter); err != nil {
				return err
			}

			if !matches {
				continue
			}

			resDocs = append(resDocs, doc)

			// if limit is set, no need to fetch all the documents
			if limit {
				break
			}
		}

		// if no documents matched, there is nothing to delete
		if len(resDocs) == 0 {
			return nil
		}

		rowsDeleted, err := deleteDocuments(ctx, dbPool, sp, resDocs)
		if err != nil {
			return err
		}

		deleted = int32(rowsDeleted)

		return nil
	})
	if err != nil {
		return 0, err
	}

	return deleted, nil
}

// deleteDocuments deletes documents by _id.
func deleteDocuments(ctx context.Context, dbPool *pgdb.Pool, sp *pgdb.SQLParam, docs []*types.Document) (int64, error) {
	ids := make([]any, len(docs))
	for i, doc := range docs {
		id := must.NotFail(doc.Get("_id"))
		ids[i] = id
	}

	var rowsDeleted int64
	err := dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
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
