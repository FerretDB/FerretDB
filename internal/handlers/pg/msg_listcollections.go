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
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/1405
	if err = common.UnimplementedNonDefault(document, "filter", func(v any) bool {
		d, ok := v.(*types.Document)
		return ok && d.Len() == 0
	}); err != nil {
		return nil, err
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/301
	// if err = common.UnimplementedNonDefault(document, "nameOnly", func(v any) bool {
	// 	nameOnly, ok := v.(bool)
	// 	return ok && !nameOnly
	// }); err != nil {
	// 	return nil, err
	// }

	common.Ignored(document, h.L, "comment", "authorizedCollections")

	var db string
	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	var names []string

	err = h.PgPool.InTransaction(ctx, func(tx pgx.Tx) error {
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
		if err = collections.Append(d); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursor", must.NotFail(types.NewDocument(
				"id", int64(0),
				"ns", db+".$cmd.listCollections",
				"firstBatch", collections,
			)),
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
