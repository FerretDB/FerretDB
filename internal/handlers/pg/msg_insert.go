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
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.l, "ordered", "writeConcern", "bypassDocumentValidation", "comment")

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
		return nil, common.NewErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
		)
	}

	var docs *types.Array
	if docs, err = common.GetOptionalParam(document, "documents", docs); err != nil {
		return nil, err
	}

	var ordered bool
	if ordered, err = common.GetOptionalParam(document, "ordered", false); err != nil {
		return nil, err
	}

	var inserted int32
	// According to MongoDB documentation, when `ordered` flag is true, we must insert
	// the documents in batch mode
	// https://www.mongodb.com/docs/manual/reference/method/db.collection.insert/
	if ordered {
		err = h.insert(ctx, &sp, docs.Slice())
		if err != nil {
			return nil, err
		}

		inserted += int32(docs.Len())
	} else {
		for i := 0; i < docs.Len(); i++ {
			doc, err := docs.Get(i)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			err = h.insert(ctx, &sp, []any{doc})
			if err != nil {
				return nil, err
			}

			inserted++
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"n", inserted,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}

// insert prepares and executes actual INSERT request to Postgres.
func (h *Handler) insert(ctx context.Context, sp *pgdb.SQLParam, docs []any) error {
	data := make([]*types.Document, len(docs))
	for i, doc := range docs {
		d, ok := doc.(*types.Document)
		if !ok {
			return common.NewErrorMsg(
				common.ErrBadValue,
				fmt.Sprintf("document has invalid type %s", common.AliasFromType(doc)),
			)
		}
		data[i] = d
	}

	err := h.pgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		if err := pgdb.InsertDocument(ctx, tx, sp.DB, sp.Collection, data); err != nil {
			if errors.Is(pgdb.ErrInvalidTableName, err) ||
				errors.Is(pgdb.ErrInvalidDatabaseName, err) {
				msg := fmt.Sprintf("Invalid namespace: %s.%s", sp.DB, sp.Collection)
				return common.NewErrorMsg(common.ErrInvalidNamespace, msg)
			}
			return lazyerrors.Error(err)
		}
		return nil
	})
	return err
}
