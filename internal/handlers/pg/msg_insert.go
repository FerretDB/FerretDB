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
	"encoding/json"
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

// MsgInsert implements HandlerInterface.
func (h *Handler) MsgInsert(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "writeConcern", "bypassDocumentValidation", "comment")

	var qp pgdb.QueryParam

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

	var docs *types.Array
	if docs, err = common.GetOptionalParam(document, "documents", docs); err != nil {
		return nil, err
	}

	ordered := true
	if ordered, err = common.GetOptionalParam(document, "ordered", ordered); err != nil {
		return nil, err
	}

	inserted, insErrors := insertMany(ctx, dbPool, &qp, docs, ordered)

	replyDoc := must.NotFail(types.NewDocument(
		"ok", float64(1),
	))

	if insErrors.Len() > 0 {
		replyDoc = insErrors.Document()
	}

	replyDoc.Set("n", inserted)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{replyDoc},
	}))

	return &reply, nil
}

// insertMany inserts many documents into the collection one by one.
//
// If insert is ordered, and a document fails to insert, handling of the remaining documents will be stopped.
// If insert is unordered, a document fails to insert, handling of the remaining documents will be continued.
//
// It always returns the number of successfully inserted documents and a document with errors.
func insertMany(ctx context.Context, dbPool *pgdb.Pool, qp *pgdb.QueryParam, docs *types.Array, ordered bool) (int32, *common.WriteErrors) { //nolint:lll // argument list is too long
	var inserted int32
	var insErrors commonerrors.WriteErrors

	for i := 0; i < docs.Len(); i++ {
		doc := must.NotFail(docs.Get(i))

		err := insertDocument(ctx, dbPool, qp, doc)

		var we *commonerrors.WriteErrors

		switch {
		case err == nil:
			inserted++
			continue
		case errors.As(err, &we):
			insErrors.Merge(we, int32(i))
		default:
			insErrors.Append(err, int32(i))
		}

		if ordered {
			return inserted, &insErrors
		}
	}

	return inserted, &insErrors
}

// insertDocument prepares and executes actual INSERT request to Postgres.
func insertDocument(ctx context.Context, dbPool *pgdb.Pool, qp *pgdb.QueryParam, doc any) error {
	d, ok := doc.(*types.Document)
	if !ok {
		return commonerrors.NewCommandErrorMsg(
			commonerrors.ErrBadValue,
			fmt.Sprintf("document has invalid type %s", common.AliasFromType(doc)),
		)
	}

	err := dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		return pgdb.InsertDocument(ctx, tx, qp.DB, qp.Collection, d)
	})

	switch {
	case err == nil:
		return nil

	case errors.Is(err, pgdb.ErrInvalidCollectionName), errors.Is(err, pgdb.ErrInvalidDatabaseName):
		msg := fmt.Sprintf("Invalid namespace: %s.%s", qp.DB, qp.Collection)
		return commonerrors.NewCommandErrorMsg(commonerrors.ErrInvalidNamespace, msg)

	case errors.Is(err, pgdb.ErrUniqueViolation):
		// TODO Extend message for non-_id unique indexes in https://github.com/FerretDB/FerretDB/issues/1509
		idMasrshaled := must.NotFail(json.Marshal(must.NotFail(d.Get("_id"))))

		return commonerrors.NewWriteErrorMsg(
			commonerrors.ErrDuplicateKey,
			fmt.Sprintf(
				`E11000 duplicate key error collection: %s.%s index: _id_ dup key: { _id: %s }`,
				qp.DB, qp.Collection, idMasrshaled,
			),
		)

	default:
		return commonerrors.CheckError(err)
	}
}
