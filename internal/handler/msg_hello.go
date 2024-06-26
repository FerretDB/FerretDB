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
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgHello implements `hello` command.
func (h *Handler) MsgHello(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	spec, err := msg.RawDocument()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resp, err := h.hello(ctx, doc, h.TCPHost, h.ReplSetName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return wire.NewOpMsg(must.NotFail(resp.Encode()))
}

// hello checks client metadata and returns hello's document fields.
// It also returns response for deprecated `isMaster` and `ismaster` commands.
func (h *Handler) hello(ctx context.Context, spec bson.AnyDocument, tcpHost, name string) (*bson.Document, error) {
	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = checkClientMetadata(ctx, doc); err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := must.NotFail(bson.NewDocument())

	switch doc.Command() {
	case "hello":
		must.NoError(res.Add("isWritablePrimary", true))
	case "isMaster", "ismaster":
		if helloOk := doc.Get("helloOk"); helloOk != nil {
			must.NoError(res.Add("helloOk", true))
		}

		must.NoError(res.Add("ismaster", true))
	default:
		panic(fmt.Sprintf("unexpected command: %q", doc.Command()))
	}

	saslSupportedMechs, err := getOptionalParam(doc, "saslSupportedMechs", "")
	if err != nil {
		return nil, err
	}

	var resSupportedMechs *bson.Array

	if saslSupportedMechs != "" {
		db, username, ok := strings.Cut(saslSupportedMechs, ".")
		if !ok {
			return nil, handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				"UserName must contain a '.' separated database.user pair",
			)
		}

		resSupportedMechs = must.NotFail(bson.NewArray("PLAIN"))

		if h.EnableNewAuth {
			if resSupportedMechs, err = h.getUserSupportedMechs(ctx, db, username); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}
	}

	if name != "" {
		// That does not work for TLS-only setups, IPv6 addresses, etc.
		// The proper solution is to support `replSetInitiate` command.
		// TODO https://github.com/FerretDB/FerretDB/issues/3936
		if strings.HasPrefix(tcpHost, ":") {
			tcpHost = "localhost" + tcpHost
		}

		must.NoError(res.Add("setName", name))
		must.NoError(res.Add("hosts", must.NotFail(bson.NewArray(tcpHost))))
	}

	must.NoError(res.Add("maxBsonObjectSize", int32(h.MaxBsonObjectSizeBytes)))
	must.NoError(res.Add("maxMessageSizeBytes", int32(wire.MaxMsgLen)))
	must.NoError(res.Add("maxWriteBatchSize", int32(100000)))
	must.NoError(res.Add("localTime", time.Now()))
	must.NoError(res.Add("logicalSessionTimeoutMinutes", int32(30)))
	must.NoError(res.Add("connectionId", int32(42)))
	must.NoError(res.Add("minWireVersion", common.MinWireVersion))
	must.NoError(res.Add("maxWireVersion", common.MaxWireVersion))
	must.NoError(res.Add("readOnly", false))

	if resSupportedMechs != nil && resSupportedMechs.Len() != 0 {
		must.NoError(res.Add("saslSupportedMechs", resSupportedMechs))
	}

	must.NoError(res.Add("ok", float64(1)))

	return res, nil
}

// getUserSupportedMechs returns supported mechanisms for the given user.
// If the user was not found, it returns nil.
func (h *Handler) getUserSupportedMechs(ctx context.Context, db, username string) (*bson.Array, error) {
	adminDB, err := h.b.Database("admin")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	usersCol, err := adminDB.Collection("system.users")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	filter, err := usersInfoFilter(false, false, db, []usersInfoPair{
		{username: username, db: db},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	qr, err := usersCol.Query(ctx, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer qr.Iter.Close()

	for {
		_, v, err := qr.Iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		matches, err := common.FilterDocument(v, filter)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if !matches {
			continue
		}

		credentialsV, _ := v.Get("credentials")
		if credentialsV == nil {
			return nil, nil
		}

		credentials := credentialsV.(*types.Document)

		supportedMechs := bson.MakeArray(len(credentials.Keys()))
		for _, mechanism := range credentials.Keys() {
			must.NoError(supportedMechs.Add(mechanism))
		}

		return supportedMechs, nil
	}

	return nil, nil
}
