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

package handler

import (
	"context"
	"errors"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDropAllUsersFromDatabase implements `dropAllUsersFromDatabase` command.
func (h *Handler) MsgDropAllUsersFromDatabase(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "writeConcern", "comment")

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	adminDB, err := h.b.Database("admin")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	users, err := adminDB.Collection("system.users")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	qr, err := users.Query(ctx, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var ids []any

	defer qr.Iter.Close()

	filter := must.NotFail(types.NewDocument("db", dbName))

	for {
		_, v, err := qr.Iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		matches, err := common.FilterDocument(v, filter)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if matches {
			ids = append(ids, must.NotFail(v.Get("_id")))
		}
	}

	var deleted int32

	if len(ids) > 0 {
		res, err := users.DeleteAll(ctx, &backends.DeleteAllParams{
			IDs: ids,
		})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		deleted = res.Deleted
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"n", deleted,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
