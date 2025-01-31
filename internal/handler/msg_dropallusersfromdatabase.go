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

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// MsgDropAllUsersFromDatabase implements `dropAllUsersFromDatabase` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgDropAllUsersFromDatabase(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, _, err = h.s.CreateOrUpdateByLSID(connCtx, spec); err != nil {
		return nil, err
	}

	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	conn, err := h.Pool.Acquire()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer conn.Release()

	usersInfoSpec := must.NotFail(wirebson.MustDocument(
		"usersInfo", int32(1),
		"$db", dbName,
	).Encode())

	usersInfo, err := documentdb_api.UsersInfo(connCtx, conn.Conn(), h.L, usersInfoSpec)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	usersInfoDoc, err := usersInfo.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	usersV, ok := usersInfoDoc.Get("users").(wirebson.AnyArray)
	if !ok {
		return wire.MustOpMsg("n", int32(0), "ok", float64(1)), nil
	}

	users, err := usersV.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var n int32

	for userV := range users.Values() {
		var user *wirebson.Document

		if user, err = userV.(wirebson.AnyDocument).Decode(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		if userDB := user.Get("db").(string); userDB != dbName {
			continue
		}

		username := user.Get("user").(string)
		dropUserSpec := must.NotFail(wirebson.MustDocument(
			"dropUser", username,
			"$db", dbName,
		).Encode())

		if _, err = documentdb_api.DropUser(connCtx, conn.Conn(), h.L, dropUserSpec); err != nil {
			return nil, lazyerrors.Error(err)
		}

		n++
	}

	return wire.MustOpMsg("n", n, "ok", float64(1)), nil
}
