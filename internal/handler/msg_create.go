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
	"regexp"
	"unicode/utf8"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
)

// collectionNameRe validates collection names.
// TODO https://github.com/FerretDB/FerretDB/issues/4879
var collectionNameRe = regexp.MustCompile("^[^\\.$\x00][^$\x00]{0,234}$")

// msgCreate implements `create` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgCreate(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
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
		msg := fmt.Sprintf("Invalid collection name: %s", collectionName)
		return nil, mongoerrors.NewWithArgument(mongoerrors.ErrInvalidNamespace, msg, command)
	}

	err = h.p.WithConn(connCtx, func(conn *pgx.Conn) error {
		_, err = documentdb_api.CreateCollection(connCtx, conn, h.L, dbName, collectionName)
		return err
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return middleware.ResponseDoc(req, wirebson.MustDocument(
		"ok", float64(1),
	))
}
