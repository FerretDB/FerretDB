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
	"time"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// msgDataSize implements `dataSize` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgDataSize(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	cmd := doc.Command()

	v := doc.Get(cmd)

	ns, ok := v.(string)
	if !ok {
		msg := fmt.Sprintf("BSON field 'dataSize.dataSize' is the wrong type '%s', expected type 'string'", aliasFromType(v))
		return nil, lazyerrors.Error(mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, cmd))
	}

	db, collection, err := splitNamespace(ns, cmd)
	if err != nil {
		return nil, err
	}

	started := time.Now()

	var pageRaw wirebson.RawDocument

	err = h.p.WithConn(func(conn *pgx.Conn) error {
		pageRaw, err = documentdb_api.CollStats(connCtx, conn, h.L, db, collection, float64(1))
		return err
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	page, err := pageRaw.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	count := page.Get("count").(int32)
	size := page.Get("totalSize")

	res := wirebson.MakeDocument(5)

	if count != 0 {
		must.NoError(res.Add("estimate", false))
	}

	must.NoError(res.Add("millis", time.Since(started).Milliseconds()))
	must.NoError(res.Add("numObjects", count))
	must.NoError(res.Add("ok", float64(1)))
	must.NoError(res.Add("size", size))

	return middleware.ResponseDoc(req, res)
}
