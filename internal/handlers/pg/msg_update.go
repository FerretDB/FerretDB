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

	if err := common.Unimplemented(document, "let"); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "ordered", "writeConcern", "bypassDocumentValidation")

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

	var updates *types.Array
	if updates, err = common.GetOptionalParam(document, "updates", updates); err != nil {
		return nil, err
	}

	err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		return pgdb.CreateCollectionIfNotExists(ctx, tx, qp.DB, qp.Collection)
	})
	if err != nil {
		if errors.Is(err, pgdb.ErrInvalidCollectionName) ||
			errors.Is(err, pgdb.ErrInvalidDatabaseName) {
			msg := fmt.Sprintf("Invalid namespace: %s.%s", qp.DB, qp.Collection)
			return nil, commonerrors.NewCommandErrorMsg(commonerrors.ErrInvalidNamespace, msg)
		}
		return nil, err
	}

	var matched, modified int32
	var upserted types.Array

	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		for i := 0; i < updates.Len(); i++ {
			update, err := common.AssertType[*types.Document](must.NotFail(updates.Get(i)))
			if err != nil {
				return err
			}

			unimplementedFields := []string{
				"c",
				"collation",
				"arrayFilters",
			}
			if err := common.Unimplemented(update, unimplementedFields...); err != nil {
				return err
			}

			common.Ignored(update, h.L, "hint")

			var q, u *types.Document
			var upsert bool
			var multi bool
			if q, err = common.GetOptionalParam(update, "q", q); err != nil {
				return err
			}
			if u, err = common.GetOptionalParam(update, "u", u); err != nil {
				// TODO check if u is an array of aggregation pipeline stages
				return err
			}

			// get comment from options.Update().SetComment() method
			if qp.Comment, err = common.GetOptionalParam(document, "comment", qp.Comment); err != nil {
				return err
			}

			// get comment from query, e.g. db.collection.UpdateOne({"_id":"string", "$comment: "test"},{$set:{"v":"foo""}})
			if qp.Comment, err = common.GetOptionalParam(q, "$comment", qp.Comment); err != nil {
				return err
			}

			if u != nil {
				if err = common.ValidateUpdateOperators(document.Command(), u); err != nil {
					return err
				}
			}

			if upsert, err = common.GetOptionalParam(update, "upsert", upsert); err != nil {
				return err
			}

			if multi, err = common.GetOptionalParam(update, "multi", multi); err != nil {
				return err
			}

			qp.Filter = q

			resDocs, err := fetchAndFilterDocs(ctx, &fetchParams{tx, &qp, h.DisablePushdown})
			if err != nil {
				return err
			}

			if len(resDocs) == 0 {
				if !upsert {
					// nothing to do, continue to the next update operation
					continue
				}

				doc := q.DeepCopy()
				if _, err = common.UpdateDocument(doc, u); err != nil {
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

			if len(resDocs) > 1 && !multi {
				resDocs = resDocs[:1]
			}

			matched += int32(len(resDocs))

			for _, doc := range resDocs {
				changed, err := common.UpdateDocument(doc, u)
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
		return nil, err
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
