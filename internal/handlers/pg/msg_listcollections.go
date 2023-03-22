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

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/pg/pgdb"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListCollections implements HandlerInterface.
func (h *Handler) MsgListCollections(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	dbPool, err := h.DBPool(ctx)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var filter *types.Document
	if filter, err = common.GetOptionalParam(document, "filter", filter); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "comment", "authorizedCollections")

	db, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	nameOnly, err := common.GetBoolOptionalParam(document, "nameOnly")
	if err != nil {
		return nil, err
	}

	var names []string

	err = dbPool.InTransaction(ctx, func(tx pgx.Tx) error {
		var err error

		names, err = pgdb.Collections(ctx, tx, db)
		if err != nil && !errors.Is(err, pgdb.ErrSchemaNotExist) {
			return lazyerrors.Error(err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	collections := types.MakeArray(len(names))
	for _, n := range names {
		d := must.NotFail(types.NewDocument(
			"name", n,
			"type", "collection",
		))

		var matches bool

		if matches, err = common.FilterDocument(d, filter); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if !matches {
			continue
		}

		if nameOnly {
			d = must.NotFail(types.NewDocument(
				"name", n,
			))
		}

		collections.Append(d)
	}

	var reply wire.OpMsg

	switch {
	case nameOnly:
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"cursor", must.NotFail(types.NewDocument(
					"firstBatch", collections,
				)),
				"ok", float64(1),
			))},
		}))
	default:
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"cursor", must.NotFail(types.NewDocument(
					"id", int64(0),
					"ns", db+".$cmd.listCollections",
					"firstBatch", collections,
				)),
				"ok", float64(1),
			))},
		}))
	}

	return &reply, nil
}
