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
			"Unsupported mechanism 'PLAIN'",
			"mechanism",
		)
	}

	payload, err := common.GetRequiredParam[types.Binary](doc, "payload")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	parts := bytes.Split(payload.B, []byte{0})
	if len(parts) != 3 || len(parts[0]) > 0 {
		return nil, common.NewCommandErrorMsgWithArgument(
			common.ErrTypeMismatch,
			"Invalid payload",
			"payload",
		)
	}

	conninfo.Get(ctx).SetAuth(string(parts[1]), string(parts[2]))

	if _, err = h.DBPool(ctx); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"done", true,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
