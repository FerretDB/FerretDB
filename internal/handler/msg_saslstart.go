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
	"log/slog"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"
	"github.com/jackc/pgx/v5"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/documentdb/documentdb_api_internal"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/logging"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
	"github.com/FerretDB/FerretDB/v2/internal/util/scram"
)

// MsgSASLStart implements `saslStart` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgSASLStart(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
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

	res, err := h.saslStart(connCtx, doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	must.NoError(res.Add("ok", float64(1)))

	return wire.NewOpMsg(res)
}

// saslStart starts SCRAM conversation.
// It returns the document containing authentication payload used for the response.
func (h *Handler) saslStart(ctx context.Context, doc *wirebson.Document) (*wirebson.Document, error) {
	if !h.Auth {
		h.L.WarnContext(ctx, "saslStart is called when authentication is disabled")
	}

	mechanism, err := getRequiredParam[string](doc, "mechanism")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if mechanism != "SCRAM-SHA-256" {
		msg := fmt.Sprintf(
			"Received authentication for mechanism %s which is not enabled",
			mechanism,
		)

		return nil, mongoerrors.NewWithArgument(mongoerrors.ErrMechanismUnavailable, msg, "mechanism")
	}

	payload, err := getRequiredParam[wirebson.Binary](doc, "payload")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	optionsV, err := getOptionalParamAny(doc, "options", wirebson.MustDocument())

	optionsDoc, ok := optionsV.(wirebson.AnyDocument)
	if !ok {
		msg := fmt.Sprintf("BSON field 'saslStart.options' is the wrong type '%T', expected type 'object'", optionsV)
		return nil, lazyerrors.Error(mongoerrors.NewWithArgument(mongoerrors.ErrTypeMismatch, msg, "options"))
	}

	options, err := optionsDoc.Decode()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	skipEmptyExchange, err := getOptionalParam(options, "skipEmptyExchange", false)
	if err != nil {
		h.L.DebugContext(
			ctx, "saslStart: skipEmptyExchange",
			slog.String("options", optionsDoc.LogMessage()), logging.Error(err),
		)

		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrAuthenticationFailed,
			"Authentication failed.",
			"saslStart",
		)
	}

	steps := 2

	if skipEmptyExchange {
		steps = 1
	}

	conninfo.Get(ctx).SetSteps(steps)

	conv := scram.NewConv(h.L)
	username, err := conv.ClientFirst(string(payload.B))
	h.L.DebugContext(
		ctx, "saslStart: client first",
		slog.String("payload", string(payload.B)), slog.String("username", username), logging.Error(err),
	)
	if err != nil {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrAuthenticationFailed,
			"Authentication failed.",
			"saslStart",
		)
	}

	var res wirebson.RawDocument

	err = h.Pool.WithConn(func(conn *pgx.Conn) error {
		res, err = documentdb_api_internal.ScramSha256GetSaltAndIterations(ctx, conn, h.L, username)
		return err
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	resDoc, err := res.DecodeDeep()
	h.L.DebugContext(
		ctx, "saslStart: salt and iterations",
		slog.Any("res", logging.LazyString(resDoc.LogMessage)), logging.Error(err),
	)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	payloadS, err := conv.ServerFirst(res)
	h.L.DebugContext(
		ctx, "saslStart: server first",
		slog.String("payload", payloadS), logging.Error(err),
	)
	if err != nil {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrAuthenticationFailed,
			"Authentication failed.",
			"saslStart",
		)
	}

	if conninfo.Get(ctx).SetConv(conv) {
		h.L.WarnContext(ctx, "saslStart: replaced existing SCRAM conversation")
	}

	return wirebson.MustDocument(
		"conversationId", int32(1),
		"done", false,
		"payload", wirebson.Binary{B: []byte(payloadS)},
	), nil
}
