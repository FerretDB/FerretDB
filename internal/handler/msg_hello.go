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
	"strings"
	"time"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/handler/session"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// MsgHello implements `hello` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgHello(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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

	resp, err := h.hello(connCtx, doc, h.TCPHost, h.ReplSetName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return wire.NewOpMsg(must.NotFail(resp.Encode()))
}

// hello checks client metadata and returns hello's document fields.
// It also returns response for deprecated `isMaster` and `ismaster` commands.
func (h *Handler) hello(ctx context.Context, spec wirebson.AnyDocument, tcpHost, name string) (*wirebson.Document, error) {
	doc, err := spec.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = checkClientMetadata(ctx, doc); err != nil {
		return nil, lazyerrors.Error(err)
	}

	res := must.NotFail(wirebson.NewDocument())

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

	if name != "" {
		// That does not work for TLS-only setups, IPv6 addresses, etc.
		// The proper solution is to support `replSetInitiate` command.
		// TODO https://github.com/FerretDB/FerretDB/issues/3936
		if strings.HasPrefix(tcpHost, ":") {
			tcpHost = "localhost" + tcpHost
		}

		must.NoError(res.Add("setName", name))
		must.NoError(res.Add("hosts", must.NotFail(wirebson.NewArray(tcpHost))))
	}

	must.NoError(res.Add("maxBsonObjectSize", maxBsonObjectSize))
	must.NoError(res.Add("maxMessageSizeBytes", int32(wire.MaxMsgLen)))
	must.NoError(res.Add("maxWriteBatchSize", maxWriteBatchSize))
	must.NoError(res.Add("localTime", time.Now()))
	must.NoError(res.Add("logicalSessionTimeoutMinutes", session.LogicalSessionTimeoutMinutes))
	must.NoError(res.Add("connectionId", connectionID))
	must.NoError(res.Add("minWireVersion", minWireVersion))
	must.NoError(res.Add("maxWireVersion", maxWireVersion))
	must.NoError(res.Add("readOnly", false))
	must.NoError(res.Add("saslSupportedMechs", wirebson.MustArray("SCRAM-SHA-256")))

	authV := doc.Get("speculativeAuthenticate")
	if authV == nil {
		must.NoError(res.Add("ok", float64(1)))

		return res, nil
	}

	authAny, ok := authV.(wirebson.AnyDocument)
	if !ok {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrTypeMismatch,
			fmt.Sprintf("speculativeAuthenticate type wrong; expected: document; got: %T", authV),
			doc.Command(),
		)
	}

	auth, err := authAny.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, err = getRequiredParam[string](auth, "db"); err != nil {
		h.L.DebugContext(ctx, "No `db` in `speculativeAuthenticate`", logging.Error(err))
		must.NoError(res.Add("ok", float64(1)))

		return res, nil
	}

	authRes, err := h.saslStart(ctx, auth)
	if err != nil {
		h.L.DebugContext(ctx, "Speculative authentication failed", logging.Error(err))

		// unsuccessful speculative authentication leave `speculativeAuthenticate` field unset
		// and let `saslStart` return an error
		must.NoError(res.Add("ok", float64(1)))

		return res, nil
	}

	must.NoError(res.Add("speculativeAuthenticate", authRes))
	must.NoError(res.Add("ok", float64(1)))

	h.L.DebugContext(ctx, "Speculative authentication passed")

	return res, nil
}
