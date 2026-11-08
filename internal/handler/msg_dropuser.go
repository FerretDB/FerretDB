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

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// msgDropUser implements `dropUser` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgDropUser(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	if _, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc); err != nil {
		return nil, err
	}

	user, err := getRequiredParam[string](doc, "dropUser")
	if err != nil {
		return nil, err
	}

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	dropUserSpec := must.NotFail(wirebson.MustDocument(
		"dropUser", user,
		"$db", dbName,
	).Encode())

	var res wirebson.RawDocument

	err = h.p.WithConn(func(conn *pgx.Conn) error {
		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/859
		res, err = documentdb_api.DropUser(connCtx, conn, h.L, dropUserSpec)
		return err
	})
	if err != nil {
		var mErr *mongoerrors.Error
		var pgErr *pgconn.PgError

		if (errors.As(err, &mErr) && mErr.Code == int32(mongoerrors.ErrUserNotFound)) ||
			(errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UndefinedObject) {
			return nil, mongoerrors.NewWithArgument(
				mongoerrors.ErrUserNotFound,
				fmt.Sprintf("User '%s@%s' not found", user, dbName),
				"dropUser",
			)
		}

		return nil, lazyerrors.Error(err)
	}

	return middleware.ResponseDoc(req, res)
}
