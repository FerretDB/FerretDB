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

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
)

// MsgDropIndexes implements `dropIndexes` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgDropIndexes(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, _, err = h.s.CreateOrUpdateByLSID(connCtx, spec); err != nil {
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

	if index := doc.Get("index"); index == nil {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrLocation40414,
			"BSON field 'dropIndexes.index' is missing but a required field",
			"dropIndexes",
		)
	}

	conn, err := h.Pool.Acquire()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer conn.Release()

	res, err := documentdb_api.DropIndexes(connCtx, conn.Conn(), h.L, dbName, spec, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// this currently fails due to
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/306
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/643
	if msg, err = wire.NewOpMsg(res); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return msg, nil
}
