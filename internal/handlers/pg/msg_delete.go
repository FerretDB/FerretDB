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

	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
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

	var qp pgdb.QueryParams

	if qp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if qp.Collection, ok = collectionParam.(string); !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	// get comment from options.Delete().SetComment() method
	if qp.Comment, err = common.GetOptionalParam(document, "comment", qp.Comment); err != nil {
		return nil, err
	}

	var deleted int32
	var delErrors commonerrors.WriteErrors

	// process every delete filter
	for i := 0; i < deletes.Len(); i++ {
		// get document with filter
		deleteDoc, err := common.AssertType[*types.Document](must.NotFail(deletes.Get(i)))
		if err != nil {
			return nil, err
		}

		var limited bool
		qp.Filter, limited, err = h.prepareDeleteParams(deleteDoc)
		if err != nil {
			return nil, err
		}

		// get comment from query, e.g. db.collection.DeleteOne({"_id":"string", "$comment: "test"})
		if qp.Comment, err = common.GetOptionalParam(qp.Filter, "$comment", qp.Comment); err != nil {
			return nil, err
		}

		del, err := execDelete(ctx, &execDeleteParams{dbPool, &qp, h.DisablePushdown, limited})
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
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{replyDoc},
	}))

	return &reply, nil
}

// prepareDeleteParams extracts query filter and limit from delete document.
func (h *Handler) prepareDeleteParams(deleteDoc *types.Document) (*types.Document, bool, error) {
	var err error

	if err = common.Unimplemented(deleteDoc, "collation"); err != nil {
		return nil, false, err
	}

	common.Ignored(deleteDoc, h.L, "hint")

	// get filter from document
	var filter *types.Document
	if filter, err = common.GetOptionalParam(deleteDoc, "q", filter); err != nil {
		return nil, false, err
	}

	// TODO use `GetLimitParam`
	// https://github.com/FerretDB/FerretDB/issues/2255
	l, err := deleteDoc.Get("limit")
	if err != nil {
		return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrMissingField,
			"BSON field 'delete.deletes.limit' is missing but a required field",
			"limit",
		)
	}

	var limit int64
	if limit, err = common.GetWholeNumberParam(l); err != nil || limit < 0 || limit > 1 {
		return nil, false, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrFailedToParse,
			fmt.Sprintf("The limit field in delete objects must be 0 or 1. Got %v", l),
			"limit",
		)
	}

	return filter, limit == 1, nil
}

// execDeleteParams contains parameters for execDelete function.
type execDeleteParams struct {
	dbPool          *pgdb.Pool
	qp              *pgdb.QueryParams
	disablePushdown bool
	limited         bool
}

// execDelete fetches documents, filters them out, limits them (if needed) and deletes them.
// If limit is true, only the first matched document is chosen for deletion, otherwise all matched documents are chosen.
// It returns the number of deleted documents or an error.
func execDelete(ctx context.Context, dp *execDeleteParams) (int32, error) {
	var deleted int32

	// filter is used to filter documents on the FerretDB side,
	// qp.Filter is used to filter documents on the PostgreSQL side (query pushdown).
	filter := dp.qp.Filter

	if dp.disablePushdown {
		dp.qp.Filter = nil
	}

	err := dp.dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		iter, err := pgdb.QueryDocuments(ctx, tx, dp.qp)
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
			if matches, err = common.FilterDocument(doc, filter); err != nil {
				return err
			}

			if !matches {
				continue
			}

			resDocs = append(resDocs, doc)

			// if limit is set, no need to fetch all the documents
			if dp.limited {
				break
			}
		}

		// if no documents matched, there is nothing to delete
		if len(resDocs) == 0 {
			return nil
		}

		rowsDeleted, err := deleteDocuments(ctx, dp.dbPool, dp.qp, resDocs)
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
func deleteDocuments(ctx context.Context, dbPool *pgdb.Pool, qp *pgdb.QueryParams, docs []*types.Document) (int64, error) {
	ids := make([]any, len(docs))
	for i, doc := range docs {
		id := must.NotFail(doc.Get("_id"))
		ids[i] = id
	}

	var rowsDeleted int64
	err := dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		var err error
		rowsDeleted, err = pgdb.DeleteDocumentsByID(ctx, tx, qp, ids)
		return err
	})
	if err != nil {
		// TODO check error code
		return 0, commonerrors.NewCommandError(commonerrors.ErrNamespaceNotFound, fmt.Errorf("delete: ns not found: %w", err))
	}

	return rowsDeleted, nil
}
