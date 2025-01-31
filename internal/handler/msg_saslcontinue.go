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

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api_internal"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
)

// MsgSASLContinue implements `saslContinue` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgSASLContinue(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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

	res, err := h.saslContinue(connCtx, doc)
	if err != nil {
		return nil, err
	}

	if msg, err = wire.NewOpMsg(res); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return msg, nil
}

// saslContinue continues and finishes SCRAM conversation.
// It returns the document containing authentication payload used for the response.
func (h *Handler) saslContinue(ctx context.Context, doc *wirebson.Document) (*wirebson.Document, error) {
	if !h.Auth {
		h.L.WarnContext(ctx, "saslContinue is called when authentication is disabled")
	}

	payload, err := getRequiredParam[wirebson.Binary](doc, "payload")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	conv := conninfo.Get(ctx).Conv()
	steps := conninfo.Get(ctx).DecrementSteps()

	if conv == nil || steps < 0 {
		h.L.WarnContext(ctx, "saslContinue: no conversation to continue")

		conninfo.Get(ctx).SetConv(nil)

		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrProtocolError,
			"No SASL session state found",
			"saslContinue",
		)
	}

	done := steps == 0
	if conv.Succeed() && done {
		return wirebson.MustDocument(
			"conversationId", int32(1),
			"done", true,
			"payload", wirebson.Binary{},
			"ok", float64(1),
		), nil
	}

	username := conv.Username()
	authMsg, clientProof, err := conv.ClientFinal(string(payload.B))
	h.L.DebugContext(
		ctx, "saslContinue: client final",
		slog.String("payload", string(payload.B)), slog.String("username", username),
		slog.String("auth_msg", authMsg), slog.String("client_proof", clientProof), logging.Error(err),
	)
	if err != nil {
		conninfo.Get(ctx).SetConv(nil)

		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrAuthenticationFailed,
			"Authentication failed.",
			"saslContinue",
		)
	}

	var res wirebson.RawDocument

	err = h.Pool.WithConn(func(conn *pgx.Conn) error {
		res, err = documentdb_api_internal.AuthenticateWithScramSha256(ctx, conn, h.L, username, authMsg, clientProof)
		return err
	})
	if err != nil {
		conninfo.Get(ctx).SetConv(nil)
		return nil, lazyerrors.Error(err)
	}

	resDoc, err := res.DecodeDeep()
	h.L.DebugContext(
		ctx, "saslContinue: authentication",
		slog.Any("res", logging.LazyString(resDoc.LogMessage)), logging.Error(err),
	)
	if err != nil {
		conninfo.Get(ctx).SetConv(nil)
		return nil, lazyerrors.Error(err)
	}

	payloadS, err := conv.ServerFinal(res)
	h.L.DebugContext(
		ctx, "saslContinue: server final",
		slog.String("payload", payloadS), logging.Error(err),
	)
	if err != nil {
		conninfo.Get(ctx).SetConv(nil)

		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrAuthenticationFailed,
			"Authentication failed.",
			"saslContinue",
		)
	}

	return wirebson.MustDocument(
		"conversationId", int32(1),
		"done", done,
		"payload", wirebson.Binary{B: []byte(payloadS)},
		"ok", float64(1),
	), nil
}
