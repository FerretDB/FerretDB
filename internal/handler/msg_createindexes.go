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

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api_internal"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// MsgCreateIndexes implements `createIndexes` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgCreateIndexes(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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

	v := doc.Get("indexes")
	if v == nil {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrLocation40414,
			"BSON field 'createIndexes.indexes' is missing but a required field",
			"createIndexes",
		)
	}

	conn, err := h.Pool.Acquire()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer conn.Release()

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1147
	// resRaw, _, _, err := documentdb_api.CreateIndexesBackground(connCtx, conn.Conn(), h.L, dbName, spec)
	resRaw, err := documentdb_api_internal.CreateIndexesNonConcurrently(connCtx, conn.Conn(), h.L, dbName, spec, true)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/292

	res, err := resRaw.DecodeDeep()
	if err != nil {
		h.L.WarnContext(connCtx, "MsgCreateIndexes failed to decode response", logging.Error(err))
		return wire.NewOpMsg(resRaw)
	}

	lazyRes := slog.Any("res", logging.LazyString(res.LogMessage))

	h.L.DebugContext(connCtx, "MsgCreateIndexes raw response", lazyRes)

	raw, _ := res.Get("raw").(*wirebson.Document)
	if raw == nil {
		h.L.WarnContext(connCtx, "MsgCreateIndexes: unexpected response", lazyRes)
		return wire.NewOpMsg(resRaw)
	}

	defaultShard, _ := raw.Get("defaultShard").(*wirebson.Document)
	if defaultShard == nil {
		h.L.WarnContext(connCtx, "MsgCreateIndexes: unexpected response", lazyRes)
		return wire.NewOpMsg(resRaw)
	}

	c, _ := defaultShard.Get("code").(int32)
	code := mongoerrors.MapWrappedCode(c)

	if code != 0 {
		errMsg, _ := defaultShard.Get("errmsg").(string)
		return nil, mongoerrors.NewWithArgument(code, errMsg, doc.Command())
	}

	resOk := defaultShard.Get("ok").(int32)
	must.NoError(defaultShard.Replace("ok", float64(resOk)))

	return wire.NewOpMsg(defaultShard)
}
