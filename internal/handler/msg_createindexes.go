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

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api_internal"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// msgCreateIndexes implements `createIndexes` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgCreateIndexes(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
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
			"indexes",
		)
	}

	var res wirebson.AnyDocument

	err = h.p.WithConn(connCtx, func(conn *pgx.Conn) error {
		res, err = h.createIndexes(connCtx, conn, doc.Command(), dbName, req.DocumentRaw())
		return err
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return middleware.ResponseDoc(req, res)
}

// createIndexes calls DocumentDB API to create indexes, decodes and maps embedded error to command error if any.
// It returns a document for createIndexes response.
func (h *Handler) createIndexes(connCtx context.Context, conn *pgx.Conn, command, dbName string, spec wirebson.RawDocument) (wirebson.AnyDocument, error) { //nolint:lll // for readability
	// TODO https://github.com/documentdb/documentdb/issues/25
	// resRaw, _, _, err := documentdb_api.CreateIndexesBackground(connCtx, conn.Conn(), h.L, dbName, spec)
	resRaw, err := documentdb_api_internal.CreateIndexesNonConcurrently(connCtx, conn, h.L, dbName, spec, true)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/292

	res, err := resRaw.DecodeDeep()
	if err != nil {
		h.L.WarnContext(connCtx, "CreateIndexes failed to decode response", logging.Error(err), slog.String("command", command))
		return resRaw, nil
	}

	lazyRes := slog.Any("res", logging.LazyString(res.LogMessage))

	h.L.DebugContext(connCtx, "CreateIndexes raw response", lazyRes, slog.String("command", command))

	raw, _ := res.Get("raw").(*wirebson.Document)
	if raw == nil {
		h.L.WarnContext(connCtx, "CreateIndexes: unexpected response", lazyRes, slog.String("command", command))
		return res, nil
	}

	defaultShard, _ := raw.Get("defaultShard").(*wirebson.Document)
	if defaultShard == nil {
		h.L.WarnContext(connCtx, "CreateIndexes: unexpected response", lazyRes, slog.String("command", command))
		return res, nil
	}

	c, _ := defaultShard.Get("code").(int32)
	code := mongoerrors.MapWrappedCode(c)

	if code != 0 {
		errMsg, _ := defaultShard.Get("errmsg").(string)
		return nil, mongoerrors.New(code, errMsg)
	}

	resOk := defaultShard.Get("ok").(int32)
	must.NoError(defaultShard.Replace("ok", float64(resOk)))

	return defaultShard, nil
}
