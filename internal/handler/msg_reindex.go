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
	"log/slog"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api_internal"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// MsgReIndex implements `reIndex` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgReIndex(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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

	v := doc.Get(doc.Command())

	collection, ok := v.(string)
	if !ok {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrInvalidNamespace,
			fmt.Sprintf("collection name has invalid type %T", v),
			"reIndex",
		)
	}

	if collection == "" {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrInvalidNamespace,
			fmt.Sprintf("Invalid namespace specified '%s.'", dbName),
			"reIndex",
		)
	}

	conn, err := h.Pool.Acquire()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer conn.Release()

	listIndexesSpec := must.NotFail(wirebson.MustDocument(
		"listIndexes", collection,
		// use large batchSize to get all results in one batch
		"cursor", wirebson.MustDocument("batchSize", int32(10000)),
	).Encode())

	page, cursorID, err := h.Pool.ListIndexes(connCtx, dbName, listIndexesSpec)
	if err != nil {
		var e *mongoerrors.Error
		if errors.As(err, &e) && e.Code == 26 {
			h.L.DebugContext(connCtx, "MsgReIndex: nothing to re-index as namespace does not exist")
			return wire.MustOpMsg("ok", float64(1)), nil
		}

		return nil, lazyerrors.Error(err)
	}

	if cursorID != 0 {
		h.L.ErrorContext(connCtx, "MsgReIndex: too many indexes some indexes are not re-indexed")
	}

	pageDoc, err := page.DecodeDeep()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	lazyRes := slog.Any("res", logging.LazyString(pageDoc.LogMessage))
	h.L.DebugContext(connCtx, "MsgReIndex: ListIndexes response", lazyRes)

	indexes := pageDoc.Get("cursor").(*wirebson.Document).Get("firstBatch").(*wirebson.Array)

	dropSpec := must.NotFail(wirebson.MustDocument(
		"dropIndexes", collection,
		"index", "*", // drops all but default _id index
	).Encode())

	res, err := documentdb_api.DropIndexes(connCtx, conn.Conn(), h.L, dbName, dropSpec, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// this currently fails due to
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/306
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/643
	resDoc, err := res.DecodeDeep()
	if err != nil {
		h.L.DebugContext(connCtx, "MsgReIndex: failed to decode DropIndexes response", logging.Error(err))
	}

	lazyRes = slog.Any("res", logging.LazyString(resDoc.LogMessage))
	h.L.DebugContext(connCtx, "MsgReIndex: DropIndexes response", lazyRes)

	createSpec := must.NotFail(wirebson.MustDocument(
		"createIndexes", collection,
		"indexes", indexes,
	).Encode())

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1147
	// resRaw, _, _, err := documentdb_api.CreateIndexesBackground(connCtx, conn.Conn(), h.L, dbName, createSpec)
	resRaw, err := documentdb_api_internal.CreateIndexesNonConcurrently(connCtx, conn.Conn(), h.L, dbName, createSpec, true)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/292

	resDoc, err = resRaw.DecodeDeep()
	if err != nil {
		h.L.WarnContext(connCtx, "MsgReIndex: failed to decode CreateIndexes response", logging.Error(err))
		return wire.NewOpMsg(resRaw)
	}

	lazyRes = slog.Any("res", logging.LazyString(resDoc.LogMessage))
	h.L.DebugContext(connCtx, "MsgReIndex: CreateIndexes response", lazyRes)

	raw, _ := resDoc.Get("raw").(*wirebson.Document)
	if raw == nil {
		h.L.WarnContext(connCtx, "MsgReIndex: unexpected CreateIndexes response", lazyRes)
		return wire.NewOpMsg(resRaw)
	}

	defaultShard, _ := raw.Get("defaultShard").(*wirebson.Document)
	if defaultShard == nil {
		h.L.WarnContext(connCtx, "MsgReIndex: unexpected CreateIndexes response", lazyRes)
		return wire.NewOpMsg(resRaw)
	}

	c, _ := defaultShard.Get("code").(int32)
	code := mongoerrors.MapWrappedCode(c)

	if code != 0 {
		errMsg, _ := defaultShard.Get("errmsg").(string)
		return nil, mongoerrors.NewWithArgument(code, errMsg, doc.Command())
	}

	resOk := defaultShard.Get("ok").(int32)

	return wire.MustOpMsg("ok", resOk), nil
}
