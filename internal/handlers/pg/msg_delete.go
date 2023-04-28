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

	params, err := common.GetDeleteParams(document, h.L)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	qp := pgdb.QueryParams{
		DB:         params.DB,
		Collection: params.Collection,
		Comment:    params.Comment,
	}

	var deleted int32
	var delErrors commonerrors.WriteErrors

	// process every delete filter
	for i, deleteParams := range params.Deletes {
		qp.Filter = deleteParams.Filter
		qp.Comment = deleteParams.Comment

		del, err := execDelete(ctx, &execDeleteParams{
			dbPool,
			&qp,
			h.DisableFilterPushdown,
			deleteParams.Limited,
		})
		if err == nil {
			deleted += del
			continue
		}

		delErrors.Append(err, int32(i))

		if params.Ordered {
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

// execDeleteParams contains parameters for execDelete function.
type execDeleteParams struct {
	dbPool                *pgdb.Pool
	qp                    *pgdb.QueryParams
	disableFilterPushdown bool
	limited               bool
}

// execDelete fetches documents, filters them out, limits them (if needed) and deletes them.
// If limit is true, only the first matched document is chosen for deletion, otherwise all matched documents are chosen.
// It returns the number of deleted documents or an error.
func execDelete(ctx context.Context, dp *execDeleteParams) (int32, error) {
	var deleted int32

	// filter is used to filter documents on the FerretDB side,
	// qp.Filter is used to filter documents on the PostgreSQL side (query pushdown).
	filter := dp.qp.Filter

	if dp.disableFilterPushdown {
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
