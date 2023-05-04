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

package tigris

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"
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

	params, err := common.GetInsertParams(document, h.L)
	if err != nil {
		return nil, err
	}

	qp := tigrisdb.QueryParams{
		DB:         params.DB,
		Collection: params.Collection,
	}

	inserted, insErrors := insertMany(ctx, dbPool, &qp, params.Docs, params.Ordered)

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
func insertMany(ctx context.Context, dbPool *tigrisdb.TigrisDB, qp *tigrisdb.QueryParams, docs *types.Array, ordered bool) (int32, *commonerrors.WriteErrors) { //nolint:lll // argument list is too long
	var inserted int32
	var insErrors commonerrors.WriteErrors

	// Attempt to insert all the documents in the same request to make insert faster.
	if err := dbPool.InsertManyDocuments(ctx, qp.DB, qp.Collection, docs); err == nil {
		return int32(docs.Len()), &insErrors
	}

	// If the transaction failed, attempt to insert each document separately.
	for i := 0; i < docs.Len(); i++ {
		doc := must.NotFail(docs.Get(i))

		err := insertDocument(ctx, dbPool, qp, doc.(*types.Document))

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

// insertDocument checks if database and collection exist, create them if needed and attempts to insertDocument the given doc.
func insertDocument(ctx context.Context, dbPool *tigrisdb.TigrisDB, qp *tigrisdb.QueryParams, doc *types.Document) error {
	err := dbPool.InsertDocument(ctx, qp.DB, qp.Collection, doc)

	var driverErr *driver.Error

	switch {
	case err == nil:
		return nil

	case errors.As(err, &driverErr):
		switch {
		case tigrisdb.IsInvalidArgument(err):
			return commonerrors.NewCommandErrorMsg(commonerrors.ErrDocumentValidationFailure, err.Error())

		case tigrisdb.IsAlreadyExists(err):
			// TODO Extend message for non-_id unique indexes in https://github.com/FerretDB/FerretDB/issues/2045
			idMasrshaled := must.NotFail(json.Marshal(must.NotFail(doc.Get("_id"))))

			return commonerrors.NewWriteErrorMsg(
				commonerrors.ErrDuplicateKey,
				fmt.Sprintf(
					`E11000 duplicate key error collection: %s.%s index: _id_ dup key: { _id: %s }`,
					qp.DB, qp.Collection, idMasrshaled,
				),
			)

		default:
			return lazyerrors.Error(err)
		}

	default:
		return commonerrors.CheckError(err)
	}
}
