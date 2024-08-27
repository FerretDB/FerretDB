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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// MsgSASLContinue implements `saslContinue` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgSASLContinue(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	doc, err := opMsgDocument(msg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	res, err := h.saslContinue(connCtx, doc)
	if err != nil {
		return nil, err
	}

	if msg, err = documentOpMsg(res); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return msg, nil
}

// saslContinue executes [scram.Step] on payload to move the conversation forward
// and returns the response containing the payload.
// If the conversation is already valid, it returns response with empty payload
// without executing [scram.Step].
func (h *Handler) saslContinue(connCtx context.Context, doc *types.Document) (*types.Document, error) {
	payload, err := common.GetRequiredParam[types.Binary](doc, "payload")
	if err != nil {
		return nil, err
	}

	_, _, conv, _ := conninfo.Get(connCtx).Auth()

	if conv == nil {
		h.L.WarnContext(connCtx, "saslContinue: no conversation to continue")

		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrProtocolError,
			"No SASL session state found",
			"saslContinue",
		)
	}

	valid := conv.Valid()

	attrs := []any{
		slog.String("username", conv.Username()),
		slog.Bool("valid", valid),
		slog.Bool("done", conv.Done()),
	}

	if valid {
		h.L.DebugContext(connCtx, "saslContinue: conversation success", attrs...) //nolint:sloglint // slice of attrs

		conninfo.Get(connCtx).SetBypassBackendAuth()

		return must.NotFail(types.NewDocument(
			"conversationId", int32(1),
			"done", valid,
			"payload", types.Binary{},
			"ok", float64(1),
		)), nil
	}

	response, err := conv.Step(string(payload.B))

	attrs = append(attrs, slog.Bool("step_valid", conv.Valid()), slog.Bool("step_done", conv.Done()))

	if err != nil {
		if h.L.Enabled(connCtx, slog.LevelDebug) {
			attrs = append(attrs, logging.Error(err))
		}

		h.L.WarnContext(connCtx, "saslContinue: step failed", attrs...) //nolint:sloglint // attrs is not key-value pairs

		conninfo.Get(connCtx).SetAuth("", "", nil, "")

		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrAuthenticationFailed,
			"Authentication failed.",
			"saslContinue",
		)
	}

	h.L.DebugContext(connCtx, "saslContinue: step succeed", attrs...) //nolint:sloglint // attrs is not key-value pairs

	return must.NotFail(types.NewDocument(
		"conversationId", int32(1),
		"done", valid, // for compatibility, assign the validity of the conversation before [Step] was called
		"payload", types.Binary{B: []byte(response)},
		"ok", float64(1),
	)), nil
}
