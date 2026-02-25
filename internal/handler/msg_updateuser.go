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

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// msgUpdateUser implements `updateUser` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgUpdateUser(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	user, err := getRequiredParam[string](doc, "updateUser")
	if err != nil {
		return nil, err
	}

	updateSpec := wirebson.MustDocument(
		"updateUser", user,
	)

	if customData := doc.Get("customData"); customData != nil {
		must.NoError(updateSpec.Add("customData", customData))
	}

	if roles := doc.Get("roles"); roles != nil {
		must.NoError(updateSpec.Add("roles", roles))
	}

	if userPassword := doc.Get("pwd"); userPassword != nil {
		must.NoError(updateSpec.Add("pwd", userPassword))
	}

	if authRestrictions := doc.Get("authenticationRestrictions"); authRestrictions != nil {
		must.NoError(updateSpec.Add("authenticationRestrictions", authRestrictions))
	}

	if mechanisms := doc.Get("mechanisms"); mechanisms != nil {
		must.NoError(updateSpec.Add("mechanisms", mechanisms))
	}

	if passwordDigestor := doc.Get("passwordDigestor"); passwordDigestor != nil {
		must.NoError(updateSpec.Add("passwordDigestor", passwordDigestor))
	}

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	must.NoError(updateSpec.Add("$db", dbName))

	var res wirebson.RawDocument

	err = h.p.WithConn(connCtx, func(conn *pgx.Conn) error {
		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/859
		res, err = documentdb_api.UpdateUser(connCtx, conn, h.L, must.NotFail(updateSpec.Encode()))
		return err
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return middleware.ResponseDoc(req, res)
}
