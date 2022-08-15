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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tigrisdb"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgInsert implements HandlerInterface.
func (h *Handler) MsgInsert(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "ordered", "writeConcern", "bypassDocumentValidation", "comment")

	var fp fetchParam
	if fp.db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}
	collectionParam, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}
	var ok bool
	if fp.collection, ok = collectionParam.(string); !ok {
		return nil, common.NewErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
		)
	}

	var docs *types.Array
	if docs, err = common.GetOptionalParam(document, "documents", docs); err != nil {
		return nil, err
	}

	var inserted int32
	for i := 0; i < docs.Len(); i++ {
		doc, err := docs.Get(i)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		err = h.insert(ctx, fp, doc.(*types.Document))
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		inserted++
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"n", inserted,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// insert checks if database and collection exist, create them if needed and attempts to insert the given doc.
// Collection creation and insert are performed in a single transaction.
func (h *Handler) insert(ctx context.Context, fp fetchParam, doc *types.Document) error {
	_, err := h.db.CreateDatabaseIfNotExists(ctx, fp.db)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return h.db.InTransaction(ctx, fp.db, func(tx driver.Tx) error {
		schema, err := tjson.DocumentSchema(doc)
		if err != nil {
			return lazyerrors.Error(err)
		}
		schema.Title = fp.collection
		b := must.NotFail(schema.Marshal())
		h.L.Sugar().Debugf("Schema:\n%s", b)

		_, err = tigrisdb.CreateCollectionIfNotExist(ctx, tx, fp.collection, b)
		if err != nil {
			return lazyerrors.Error(err)
		}

		b, err = tjson.Marshal(doc)
		if err != nil {
			return lazyerrors.Error(err)
		}
		h.L.Sugar().Debugf("Document:\n%s", b)

		_, err = tx.Insert(ctx, fp.collection, []driver.Document{b})
		if err != nil {
			return lazyerrors.Error(err)
		}
		return nil
	})
}
