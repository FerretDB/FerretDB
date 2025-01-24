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
	"fmt"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// MsgKillCursors implements `killCursors` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgKillCursors(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	document, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	command := document.Command()

	db, err := getRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := getRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	username := conninfo.Get(connCtx).Conv().Username()

	userID, _, err := h.s.CreateOrUpdateByLSID(connCtx, spec)
	if err != nil {
		return nil, err
	}

	cursorsV, err := getRequiredParamAny(document, "cursors")
	if err != nil {
		return nil, err
	}

	curArr, ok := cursorsV.(wirebson.AnyArray)
	if !ok {
		msg := fmt.Sprintf(`required parameter "cursors" has type %T (expected array)`, cursorsV)
		return nil, lazyerrors.Error(mongoerrors.NewWithArgument(mongoerrors.ErrBadValue, msg, command))
	}

	cursors, err := curArr.Decode()
	if !ok {
		return nil, lazyerrors.Error(err)
	}

	var ids []int64
	cursorsKilled := wirebson.MakeArray(0)
	cursorsNotFound := wirebson.MakeArray(0)
	cursorsAlive := wirebson.MakeArray(0)
	cursorsUnknown := wirebson.MakeArray(0)

	for i := range cursors.Len() {
		v := cursors.Get(i)

		id, ok := v.(int64)
		if !ok {
			return nil, mongoerrors.NewWithArgument(
				mongoerrors.ErrTypeMismatch,
				fmt.Sprintf(
					"BSON field 'killCursors.cursors.%d' is the wrong type '%s', expected type 'long'",
					i,
					aliasFromType(v),
				),
				command,
			)
		}

		ids = append(ids, id)
	}

	for _, id := range ids {
		// Should we check database and collection names?
		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/17
		_, _, _ = db, collection, username

		if err = h.s.DeleteCursor(userID, id, db); err != nil {
			return nil, err
		}

		if deleted := h.Pool.KillCursor(connCtx, id); !deleted {
			must.NoError(cursorsNotFound.Add(id))
			continue
		}

		must.NoError(cursorsKilled.Add(id))
	}

	res := must.NotFail(wirebson.NewDocument(
		"cursorsKilled", cursorsKilled,
		"cursorsNotFound", cursorsNotFound,
		"cursorsAlive", cursorsAlive,
		"cursorsUnknown", cursorsUnknown,
		"ok", float64(1),
	))

	return wire.NewOpMsg(must.NotFail(res.Encode()))
}
