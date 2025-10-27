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
	"unicode/utf8"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// msgDrop implements `drop` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgDrop(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	command := doc.Command()

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	collectionName, err := getRequiredParam[string](doc, command)
	if err != nil {
		return nil, err
	}

	if !collectionNameRe.MatchString(collectionName) ||
		!utf8.ValidString(collectionName) {
		msg := fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collectionName)
		return nil, mongoerrors.NewWithArgument(mongoerrors.ErrInvalidNamespace, msg, command)
	}

	// Should we manually close all cursors for the collection?
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/17

	var dropped bool

	err = h.p.WithConn(func(conn *pgx.Conn) error {
		dropped, err = documentdb_api.DropCollection(connCtx, conn, h.L, dbName, collectionName, nil, nil, false)
		return err
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := wirebson.MakeDocument(3)
	if dropped {
		must.NoError(res.Add("nIndexesWas", int32(1))) // TODO https://github.com/FerretDB/FerretDB/issues/2337
		must.NoError(res.Add("ns", dbName+"."+collectionName))
	}

	must.NoError(res.Add("ok", float64(1)))

	return middleware.ResponseDoc(req, res)
}
