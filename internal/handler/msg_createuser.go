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
	"log/slog"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// MsgCreateUser implements `createUser` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgCreateUser(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, _, err = h.s.CreateOrUpdateByLSID(connCtx, spec); err != nil {
		return nil, err
	}

	doc, err := spec.DecodeDeep()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/913
	doc.Remove("mechanisms")

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/911
	roles, _ := doc.Get("roles").(*wirebson.Array)
	if roles == nil || roles.Len() == 0 {
		roles = wirebson.MustArray(
			wirebson.MustDocument("role", "clusterAdmin", "db", "admin"),
			wirebson.MustDocument("role", "readWriteAnyDatabase", "db", "admin"),
		)

		h.L.WarnContext(
			connCtx, "Adding default roles",
			slog.Any("user", doc.Get(doc.Command())), slog.String("roles", roles.LogMessage()),
		)

		must.NoError(doc.Replace("roles", roles))
		spec = must.NotFail(doc.Encode())
	}

	var res wirebson.RawDocument

	err = h.Pool.WithConn(func(conn *pgx.Conn) error {
		res, err = documentdb_api.CreateUser(connCtx, conn, h.L, spec)
		return err
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return wire.NewOpMsg(res)
}
