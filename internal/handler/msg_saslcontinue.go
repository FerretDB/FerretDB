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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
	"go.uber.org/zap"
)

// MsgSASLContinue implements `saslContinue` command.
func (h *Handler) MsgSASLContinue(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var payload []byte

	binaryPayload, err := common.GetRequiredParam[types.Binary](document, "payload")
	if err == nil {
		payload = binaryPayload.B
	}

	// c=biws,r=F68dvyilmZIMEFz+3CisPp1HxZ32mqLt66Rcm4+X1R+wPNI0xhXAEayWgwSapNbT,p=N/8JIXLonwJz2i6yNKGX3lZLMGTvf8PAfNmmVrQGxis=
	conv := conninfo.Get(ctx).Conv()
	response, err := conv.Step(string(payload))
	h.L.Debug(
		"saslContinue",
		zap.String("payload", string(payload)),
		zap.String("response", response),
		zap.Error(err),
		zap.Bool("nil error", err == nil),
	)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"conversationId", int32(1),
			"done", true,
			"payload", string(payload),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
