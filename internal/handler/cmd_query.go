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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// CmdQuery implements deprecated OP_QUERY message handling.
func (h *Handler) CmdQuery(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	body := query.Query()
	cmd := body.Command()
	collection := query.FullCollectionName

	// both are valid and are allowed to be run against any database as we don't support authorization yet
	if (cmd == "ismaster" || cmd == "isMaster") && strings.HasSuffix(collection, ".$cmd") {
		return common.IsMaster(ctx, body, h.TCPHost, h.ReplSetName)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3008

	// database name typically is either "$external" or "admin"

	if cmd == "saslStart" && strings.HasSuffix(collection, ".$cmd") {
		mechanism, err := common.GetRequiredParam[string](body, "mechanism")
		if err == nil && mechanism == "PLAIN" {
			username, password, err := saslStartPlain(body)
			if err != nil {
				return nil, err
			}

			conninfo.Get(ctx).SetAuth(username, password)
		}

		var emptyPayload types.Binary
		reply := wire.OpReply{
			NumberReturned: 1,
		}
		reply.SetDocument(must.NotFail(types.NewDocument(
			"conversationId", int32(1),
			"done", true,
			"payload", emptyPayload,
			"ok", float64(1),
		)))

		return &reply, nil
	}

	return nil, handlererrors.NewCommandErrorMsgWithArgument(
		handlererrors.ErrNotImplemented,
		fmt.Sprintf("CmdQuery: unhandled command %q for collection %q", cmd, collection),
		"OpQuery: "+cmd,
	)
}
