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

package pg

import (
	"bytes"
	"context"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgSASLStart implements HandlerInterface.
func (h *Handler) MsgSASLStart(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	doc, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	mechanism, err := common.GetRequiredParam[string](doc, "mechanism")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if mechanism != "PLAIN" {
		return nil, common.NewCommandErrorMsgWithArgument(
			common.ErrTypeMismatch,
			"Unsupported mechanism '"+mechanism+"'",
			"mechanism",
		)
	}

	payload, err := common.GetRequiredParam[types.Binary](doc, "payload")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	parts := bytes.Split(payload.B, []byte{0})
	if l := len(parts); l != 3 {
		return nil, common.NewCommandErrorMsgWithArgument(
			common.ErrTypeMismatch,
			fmt.Sprintf("Invalid payload (expected 3 parts, got %d)", l),
			"payload",
		)
	}

	authzid, authcid, passwd := parts[0], parts[1], parts[2]

	// Some drivers (Go) send empty authorization identity (authzid),
	// while others (Java) set it to the same value as authentication identity (authcid)
	// (see https://www.rfc-editor.org/rfc/rfc4616.html).
	// Ignore authzid for now.
	_ = authzid

	conninfo.Get(ctx).SetAuth(string(authcid), string(passwd))

	if _, err = h.DBPool(ctx); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var emptyPayload types.Binary
	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"conversationId", int32(1),
			"done", true,
			"payload", types.Binary{},
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
