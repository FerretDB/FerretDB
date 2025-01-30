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

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// collectionNameRe validates collection names.
var collectionNameRe = regexp.MustCompile("^[^\\.$\x00][^$\x00]{0,234}$")

// MsgCreate implements `create` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgCreate(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, _, err = h.s.CreateOrUpdateByLSID(connCtx, spec); err != nil {
		return nil, err
	}

	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dbName, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	collectionName, err := getRequiredParam[string](doc, "create")
	if err != nil {
		return nil, err
	}

	if !collectionNameRe.MatchString(collectionName) ||
		!utf8.ValidString(collectionName) {
		msg := fmt.Sprintf("Invalid collection name: %s", collectionName)
		return nil, mongoerrors.NewWithArgument(mongoerrors.ErrInvalidNamespace, msg, "create")
	}

	conn, err := h.Pool.Acquire()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer conn.Release()

	_, err = documentdb_api.CreateCollection(connCtx, conn.Conn(), h.L, dbName, collectionName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := wirebson.MustDocument(
		"ok", float64(1),
	)

	return wire.NewOpMsg(must.NotFail(res.Encode()))
}
