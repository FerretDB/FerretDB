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

package hana

import (
	"context"
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/hana/hanadb"
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

	ignoredFields := []string{
		"writeConcern",
	}
	common.Ignored(document, h.L, ignoredFields...)

	params, err := common.GetInsertParams(document, h.L)
	if err != nil {
		return nil, err
	}

	qp := hanadb.QueryParams{
		DB:         params.DB,
		Collection: params.Collection,
	}

	inserted, insErrors := insertMany(ctx, dbPool, &qp, params.Docs, params.Ordered)

	replyDoc := must.NotFail(types.NewDocument(
		"n", inserted,
		"ok", float64(1),
	))

	if insErrors.Len() > 0 {
		replyDoc = insErrors.Document()
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{replyDoc},
	}))

	return &reply, nil
}

func insertMany(ctx context.Context, dbPool *hanadb.Pool, qp *hanadb.QueryParams, docs *types.Array, ordered bool) (int32, *commonerrors.WriteErrors) { //nolint:lll // argument list is too long
	var inserted int32
	var insErrors commonerrors.WriteErrors

	// TODO: Bulk Insert
	// Attempt to insert all the documents in the same request to make insert faster.
	/*if err := dbPool.InsertManyDocuments(ctx, qp docs); err == nil {
		return int32(docs.Len()), &insErrors
	}*/

	// If the transaction failed, attempt to insert each document separately.
	for i := 0; i < docs.Len(); i++ {
		doc := must.NotFail(docs.Get(i))

		err := insertOne(ctx, dbPool, qp, doc.(*types.Document))

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

// insertOne checks if database and collection exist, create them if needed and attempts to insertDocument the given doc.
func insertOne(ctx context.Context, dbPool *hanadb.Pool, qp *hanadb.QueryParams, doc *types.Document) error {
	toInsert := doc

	if !toInsert.Has("_id") {
		// Make a copy so that original document could be sent to the proxy as it is.
		toInsert = doc.DeepCopy()

		toInsert.Set("_id", types.NewObjectID())
	}

	err := dbPool.InsertOne(ctx, qp, toInsert)

	switch {
	case err == nil:
		return nil
	case errors.Is(err, hanadb.ErrInvalidCollectionName), errors.Is(err, hanadb.ErrInvalidDatabaseName):
		msg := fmt.Sprintf("Invalid namespace: %s.%s", qp.DB, qp.Collection)
		return commonerrors.NewCommandErrorMsg(commonerrors.ErrInvalidNamespace, msg)

	// TODO: set up some sort of metadata table to keep track of '_ids' so we can track duplicates
	/*case errors.Is(err, hanzdb.ErrUniqueViolation):
	// TODO Extend message for non-_id unique indexes in https://github.com/FerretDB/FerretDB/issues/2045
	idMasrshaled := must.NotFail(json.Marshal(must.NotFail(d.Get("_id"))))

	return commonerrors.NewWriteErrorMsg(
		commonerrors.ErrDuplicateKey,
		fmt.Sprintf(
			`E11000 duplicate key error collection: %s.%s index: _id_ dup key: { _id: %s }`,
			qp.DB, qp.Collection, idMasrshaled,
		),
	)
	*/
	default:
		return lazyerrors.Error(err)
	}
}
