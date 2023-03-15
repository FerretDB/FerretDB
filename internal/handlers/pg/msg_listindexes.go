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

// MsgListIndexes implements HandlerInterface.
func (h *Handler) MsgListIndexes(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "comment", "cursor")

	var db string

	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	var collectionParam any

	if collectionParam, err = document.Get(document.Command()); err != nil {
		return nil, err
	}

	collection, ok := collectionParam.(string)
	if !ok {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
			document.Command(),
		)
	}

	var indexes []pgdb.IndexParams

	err = dbPool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		indexes, err = pgdb.ListIndexes(ctx, tx, db, collection)
		return err
	})

	switch {
	case err == nil:
		// do nothing
	case err == pgdb.ErrTableNotExist:
		return nil, commonerrors.NewCommandErrorMsg(
			commonerrors.ErrNamespaceNotFound,
			fmt.Sprintf("ns does not exist: %s.%s", db, collection),
		)
	default:
		return nil, lazyerrors.Error(err)
	}

	firstBatch := must.NotFail(types.NewArray(
		must.NotFail(types.NewDocument(
			"v", int32(2),
			"key", must.NotFail(types.NewDocument(
				"_id", int32(1),
			)),
			"name", "_id_",
		)),
	))

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"id", int64(0),
				"ns", fmt.Sprintf("%s.%s", db, collection),
				"firstBatch", firstBatch,
			)),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
