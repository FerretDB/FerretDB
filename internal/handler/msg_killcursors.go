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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgKillCursors implements `killCursors` command.
func (h *Handler) MsgKillCursors(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
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

	username, _ := conninfo.Get(ctx).Auth()

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
		c := h.cursors.Get(id)
		if c == nil || c.DB != db || c.Collection != collection || c.Username != username {
			cursorsNotFound.Append(id)
			continue
		}

		h.cursors.CloseAndRemove(c)
		cursorsKilled.Append(id)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"cursorsKilled", cursorsKilled,
			"cursorsNotFound", cursorsNotFound,
			"cursorsAlive", cursorsAlive,
			"cursorsUnknown", cursorsUnknown,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
