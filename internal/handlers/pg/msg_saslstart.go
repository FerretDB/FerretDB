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
	"context"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
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

	var username, password string

	switch mechanism {
	case "PLAIN":
		username, password, err = common.SASLStartPlain(doc)
	default:
		err = commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrTypeMismatch,
			"Unsupported mechanism '"+mechanism+"'",
			"mechanism",
		)
	}
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	conninfo.Get(ctx).SetAuth(username, password)

	if _, err = h.DBPool(ctx); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var emptyPayload types.Binary
	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"conversationId", int32(1),
			"done", true,
			"payload", emptyPayload,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
