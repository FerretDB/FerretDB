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
	"fmt"

	"github.com/FerretDB/wire"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// MsgKillCursors implements `killCursors` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgKillCursors(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := bson.OpMsgDocument(msg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	db, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := common.GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	username := conninfo.Get(connCtx).Username()

	cursors, err := common.GetRequiredParam[*types.Array](document, "cursors")
	if err != nil {
		return nil, err
	}

	iter := cursors.Iterator()
	defer iter.Close()

	var ids []int64
	cursorsKilled := types.MakeArray(0)
	cursorsNotFound := types.MakeArray(0)
	cursorsAlive := types.MakeArray(0)
	cursorsUnknown := types.MakeArray(0)

	for {
		i, v, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.ErrIteratorDone) {
				break
			}

			return nil, lazyerrors.Error(err)
		}

		id, ok := v.(int64)
		if !ok {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrTypeMismatch,
				fmt.Sprintf(
					"BSON field 'killCursors.cursors.%d' is the wrong type '%s', expected type 'long'",
					i,
					handlerparams.AliasFromType(v),
				),
				command,
			)
		}

		ids = append(ids, id)
	}

	for _, id := range ids {
		cursor := h.cursors.Get(id)
		if cursor == nil || cursor.DB != db || cursor.Collection != collection || cursor.Username != username {
			cursorsNotFound.Append(id)
			continue
		}

		h.cursors.CloseAndRemove(cursor)
		cursorsKilled.Append(id)
	}

	return bson.NewOpMsg(
		must.NotFail(types.NewDocument(
			"cursorsKilled", cursorsKilled,
			"cursorsNotFound", cursorsNotFound,
			"cursorsAlive", cursorsAlive,
			"cursorsUnknown", cursorsUnknown,
			"ok", float64(1),
		)),
	)
}
