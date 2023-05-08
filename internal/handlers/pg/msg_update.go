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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgUpdate implements HandlerInterface.
func (h *Handler) MsgUpdate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	params, err := common.GetUpdateParams(document, h.L)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		return pgdb.CreateCollectionIfNotExists(ctx, tx, params.DB, params.Collection)
	})

	switch {
	case err == nil:
		// do nothing
	case errors.Is(err, pgdb.ErrInvalidCollectionName), errors.Is(err, pgdb.ErrInvalidDatabaseName):
		msg := fmt.Sprintf("Invalid namespace: %s.%s", params.DB, params.Collection)
		return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrInvalidNamespace, msg, document.Command())
	default:
		return nil, lazyerrors.Error(err)
	}

	var matched, modified int32
	var upserted types.Array

	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		for _, u := range params.Updates {
			qp := pgdb.QueryParams{
				DB:         params.DB,
				Collection: params.Collection,
				Filter:     u.Filter,
				Comment:    u.Comment,
			}

			resDocs, err := fetchAndFilterDocs(ctx, &fetchParams{tx, &qp, h.DisableFilterPushdown})
			if err != nil {
				return err
			}

			if len(resDocs) == 0 {
				if !u.Upsert {
					// nothing to do, continue to the next update operation
					continue
				}

				doc := u.Filter.DeepCopy()
				if _, err = common.UpdateDocument(doc, u.Update); err != nil {
					return err
				}
				if !doc.Has("_id") {
					doc.Set("_id", types.NewObjectID())
				}

				upserted.Append(must.NotFail(types.NewDocument(
					"index", int32(0), // TODO
					"_id", must.NotFail(doc.Get("_id")),
				)))

				if err = insertDocument(ctx, dbPool, &qp, doc); err != nil {
					return err
				}

				matched++
				continue
			}

			if len(resDocs) > 1 && !u.Multi {
				resDocs = resDocs[:1]
			}

			matched += int32(len(resDocs))

			for _, doc := range resDocs {
				changed, err := common.UpdateDocument(doc, u.Update)
				if err != nil {
					return err
				}

				if !changed {
					continue
				}

				rowsChanged, err := updateDocument(ctx, tx, &qp, doc)
				if err != nil {
					return err
				}
				modified += int32(rowsChanged)
			}
		}

		return nil
	})

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := must.NotFail(types.NewDocument(
		"n", matched,
	))

	if upserted.Len() != 0 {
		res.Set("upserted", &upserted)
	}

	res.Set("nModified", modified)
	res.Set("ok", float64(1))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{res},
	}))

	return &reply, nil
}

// updateDocument updates documents by _id.
func updateDocument(ctx context.Context, tx pgx.Tx, qp *pgdb.QueryParams, doc *types.Document) (int64, error) {
	id := must.NotFail(doc.Get("_id"))

	res, err := pgdb.SetDocumentByID(ctx, tx, qp, id, doc)
	if err == nil {
		return res, nil
	}

	return 0, commonerrors.CheckError(err)
}
