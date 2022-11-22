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
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgCount implements HandlerInterface.
func (h *Handler) MsgCount(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	unimplementedFields := []string{
		"skip",
		"collation",
	}
	if err := common.Unimplemented(document, unimplementedFields...); err != nil {
		return nil, err
	}

	ignoredFields := []string{
		"hint",
		"readConcern",
		"comment",
	}
	common.Ignored(document, h.L, ignoredFields...)

	var filter *types.Document
	if filter, err = common.GetOptionalParam(document, "query", filter); err != nil {
		return nil, err
	}

	var limit int64
	if l, _ := document.Get("limit"); l != nil {
		if limit, err = common.GetWholeNumberParam(l); err != nil {
			return nil, err
		}
	}

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
		return nil, common.NewCommandErrorMsg(
			common.ErrBadValue,
			fmt.Sprintf("collection name has invalid type %s", common.AliasFromType(collectionParam)),
		)
	}

	sp.Filter = filter

	resDocs := make([]*types.Document, 0, 16)
	err = h.PgPool.InTransaction(ctx, func(tx pgx.Tx) error {
		it, ierr := h.PgPool.QueryDocuments(ctx, tx, &sp)
		if ierr != nil {
			return ierr
		}

		if it == nil {
			// no documents found
			return nil
		}

		for {
			_, doc, ierr := it.Next()

			switch {
			case ierr == nil:
				// do nothing
			case errors.Is(ierr, iterator.ErrIteratorDone):
				// no more documents
				return nil
			case errors.Is(ierr, context.Canceled), errors.Is(ierr, context.DeadlineExceeded):
				h.L.Warn(
					"context canceled, stopping fetching",
					zap.String("db", sp.DB), zap.String("collection", sp.Collection), zap.Error(ierr),
				)
				return nil
			default:
				return ierr
			}

			matches, ierr := common.FilterDocument(doc, filter)
			if ierr != nil {
				return ierr
			}

			if !matches {
				continue
			}

			resDocs = append(resDocs, doc)
		}
	})

	if err != nil {
		return nil, err
	}

	if resDocs, err = common.LimitDocuments(resDocs, limit); err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"n", int32(len(resDocs)),
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
