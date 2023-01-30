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

	var docs *types.Array
	if docs, err = common.GetOptionalParam(document, "documents", docs); err != nil {
		return nil, err
	}

	ordered := true
	if ordered, err = common.GetOptionalParam(document, "ordered", ordered); err != nil {
		return nil, err
	}

	inserted, insErrors := insertMany(ctx, dbPool, &sp, docs, ordered)

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
func insertMany(ctx context.Context, dbPool *pgdb.Pool, sp *pgdb.SQLParam, docs *types.Array, ordered bool) (int32, *common.WriteErrors) { //nolint:lll // argument list is too long
	var inserted int32
	var insErrors common.WriteErrors

	for i := 0; i < docs.Len(); i++ {
		doc := must.NotFail(docs.Get(i))

		err := insertDocument(ctx, dbPool, sp, doc)

		var we *common.WriteErrors

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
func insertDocument(ctx context.Context, dbPool *pgdb.Pool, sp *pgdb.SQLParam, doc any) error {
	d, ok := doc.(*types.Document)
	if !ok {
		return common.NewCommandErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("document has invalid type %s", common.AliasFromType(doc)),
		)
	}

	err := dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		return pgdb.InsertDocument(ctx, tx, sp.DB, sp.Collection, d)
	})
	if err == nil {
		return nil
	}

	if errors.Is(err, pgdb.ErrInvalidCollectionName) || errors.Is(err, pgdb.ErrInvalidDatabaseName) {
		msg := fmt.Sprintf("Invalid namespace: %s.%s", sp.DB, sp.Collection)
		return common.NewCommandErrorMsg(common.ErrInvalidNamespace, msg)
	}

	return common.CheckError(err)
}
