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

	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgSASLContinue implements `saslContinue` command.
func (h *Handler) MsgSASLContinue(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	doc, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var payload []byte

	binaryPayload, err := common.GetRequiredParam[types.Binary](doc, "payload")
	if err != nil {
		return nil, err
	}

	payload = binaryPayload.B

	_, _, conv := conninfo.Get(ctx).Auth()

	if conv == nil {
		h.L.Warn("saslContinue: no conversation to continue")

		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrAuthenticationFailed,
			"Authentication failed.",
			"saslContinue",
		)
	}

	response, err := conv.Step(string(payload))

	fields := []zap.Field{
		zap.String("username", conv.Username()),
		zap.Bool("valid", conv.Valid()),
		zap.Bool("done", conv.Done()),
	}

	if err != nil {
		if h.L.Level().Enabled(zap.DebugLevel) {
			fields = append(fields, zap.Error(err))
		}

		h.L.Warn("saslContinue: step failed", fields...)

		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrAuthenticationFailed,
			"Authentication failed.",
			"saslContinue",
		)
	}

	h.L.Debug("saslContinue: step succeed", fields...)

	if conv.Valid() {
		conninfo.Get(ctx).SetBypassBackendAuth()
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.MakeOpMsgSection(
		must.NotFail(types.NewDocument(
			"conversationId", int32(1),
			"done", conv.Done(),
			"payload", types.Binary{B: []byte(response)},
			"ok", float64(1),
		)),
	)))

	return &reply, nil
}
