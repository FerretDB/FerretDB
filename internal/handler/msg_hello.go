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
	doc, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resp, err := h.hello(ctx, doc, h.TCPHost, h.ReplSetName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.MakeOpMsgSection(resp)))

	return &reply, nil
}

// hello checks client metadata and returns hello's document fields.
// It also returns response for deprecated `isMaster` and `ismaster` commands.
func (h *Handler) hello(ctx context.Context, doc *types.Document, tcpHost, name string) (*types.Document, error) {
	if err := checkClientMetadata(ctx, doc); err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := must.NotFail(types.NewDocument())

	switch doc.Command() {
	case "hello":
		res.Set("isWritablePrimary", true)
	case "isMaster", "ismaster":
		if helloOk, _ := doc.Get("helloOk"); helloOk != nil {
			res.Set("helloOk", true)
		}

		res.Set("ismaster", true)
	default:
		panic(fmt.Sprintf("unexpected command: %q", doc.Command()))
	}

	saslSupportedMechs, err := common.GetOptionalParam(doc, "saslSupportedMechs", "")
	if err != nil {
		return nil, err
	}

	var resSupportedMechs *types.Array

	if saslSupportedMechs != "" {
		db, username, ok := strings.Cut(saslSupportedMechs, ".")
		if !ok {
			return nil, handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				"UserName must contain a '.' separated database.user pair",
			)
		}

		resSupportedMechs = must.NotFail(types.NewArray("PLAIN"))

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

		res.Set("setName", name)
		res.Set("hosts", must.NotFail(types.NewArray(tcpHost)))
	}

	res.Set("maxBsonObjectSize", int32(h.MaxBsonObjectSizeBytes))
	res.Set("maxMessageSizeBytes", int32(wire.MaxMsgLen))
	res.Set("maxWriteBatchSize", maxWriteBatchSize)
	res.Set("localTime", time.Now())
	res.Set("logicalSessionTimeoutMinutes", logicalSessionTimeoutMinutes)
	res.Set("connectionId", connectionID)
	res.Set("minWireVersion", common.MinWireVersion)
	res.Set("maxWireVersion", common.MaxWireVersion)
	res.Set("readOnly", false)

	if resSupportedMechs != nil && resSupportedMechs.Len() != 0 {
		res.Set("saslSupportedMechs", resSupportedMechs)
	}

	res.Set("ok", float64(1))

	return res, nil
}

// getUserSupportedMechs returns supported mechanisms for the given user.
// If the user was not found, it returns nil.
func (h *Handler) getUserSupportedMechs(ctx context.Context, db, username string) (*types.Array, error) {
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

		supportedMechs := types.MakeArray(len(credentials.Keys()))
		for _, mechanism := range credentials.Keys() {
			supportedMechs.Append(mechanism)
		}

		return supportedMechs, nil
	}

	return nil, nil
}
