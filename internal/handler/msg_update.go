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
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// MsgUpdate implements `update` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgUpdate(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	opID := h.operations.Start("update")
	defer h.operations.Stop(opID)

	spec, seq := msg.RawSections()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, spec); err != nil {
		return nil, err
	}

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/78
	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	collection, _ := doc.Get(doc.Command()).(string)
	h.operations.Update(opID, dbName, collection, doc)

	var res wirebson.RawDocument

	err = h.Pool.WithConn(func(conn *pgx.Conn) error {
		res, _, err = documentdb_api.Update(connCtx, conn, h.L, dbName, spec, seq)
		return err
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return wire.NewOpMsg(mongoerrors.MapWriteErrors(connCtx, res))
}
