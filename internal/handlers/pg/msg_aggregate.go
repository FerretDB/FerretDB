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
	"github.com/FerretDB/FerretDB/internal/handlers/common/aggregations"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgAggregate implements HandlerInterface.
func (h *Handler) MsgAggregate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// FIXME
	common.Ignored(document, h.L, "cursor")

	if err = common.Unimplemented(document, "explain", "bypassDocumentValidation", "hint"); err != nil {
		return nil, err
	}

	if err = common.Unimplemented(document, "readConcern", "writeConcern"); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "allowDiskUse", "maxTimeMS", "collation", "comment", "let")

	var sp pgdb.SQLParam

	if sp.DB, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	collection, err := document.Get(document.Command())
	if err != nil {
		return nil, err
	}

	var ok bool
	if sp.Collection, ok = collection.(string); !ok {
		return nil, common.NewCommandErrorMsgWithArgument(
			commonerrors.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collection)),
			document.Command(),
		)
	}

	pipeline, err := common.GetRequiredParam[*types.Array](document, "pipeline")
	if err != nil {
		return nil, err
	}

	st := make([]aggregations.Stage, pipeline.Len())
	iter := pipeline.Iterator()
	defer iter.Close()

	for {
		_, s, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return nil, lazyerrors.Error(err)
		}

		ss, err := aggregations.NewStage(s.(*types.Document))
		if err != nil {
			return nil, err
		}

		st = append(st, ss)
	}

	var docs []*types.Document
	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		docs, err = h.fetchAndFilterDocs(ctx, tx, &sp)
		return err
	})

	if err != nil {
		return nil, err
	}

	for _, s := range st {
		if docs, err = s.Process(ctx, docs); err != nil {
			return nil, err
		}
	}

	// TODO
	_ = docs

	return nil, commonerrors.NewCommandErrorMsg(commonerrors.ErrNotImplemented, "`aggregate` command is not implemented yet")
}
