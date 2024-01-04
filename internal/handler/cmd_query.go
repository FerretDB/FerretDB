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
	scramutil "github.com/FerretDB/FerretDB/internal/util/scram"
	"github.com/FerretDB/FerretDB/internal/wire"
	"go.uber.org/zap"
)

// CmdQuery implements deprecated OP_QUERY message handling.
func (h *Handler) CmdQuery(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	cmd := query.Query.Command()
	collection := query.FullCollectionName

	doc := query.Query

	var response string

	var sconv *scramutil.ScramConveration

	// to reduce connection overhead time, clients may use a hello command to complete their authentication exchange
	// if so, the saslStart command may be embedded under the speculativeAuthenticate field
	if (cmd == "ismaster" || cmd == "hello") && doc.Has("speculativeAuthenticate") {
		reply, err := common.IsMaster(ctx, doc)
		must.NoError(err)

		v, _ := doc.Get("speculativeAuthenticate")
		d := v.(*types.Document)
		payload, _ := d.Get("payload")

		response, sconv, err = saslStartSCRAM(d)
		must.NoError(err)

		conninfo.Get(ctx).SetConv(sconv)
		conninfo.Get(ctx).SetAuth(sconv.Conv.Username(), "password") // hack

		h.L.Debug(
			"OP_QUERY",
			zap.String("command", cmd),
			zap.Any("payload", payload),
			zap.String("response", response),
			zap.String("user", sconv.Conv.Username()),
			zap.Bool("authenticated", sconv.Conv.Valid()),
		)

		// create a speculative conversation document for SCRAM authentication
		reply.Documents[0].Set("speculativeAuthenticate", must.NotFail(types.NewDocument(
			"helloOk", true,
			"conversationId", int32(1),
			"done", false,
			"payload", response,
			"ok", float64(1),
		)))

		return reply, nil
	}

	// both are valid and are allowed to be run against any database as we don't support authorization yet
	if (cmd == "ismaster" || cmd == "isMaster") && strings.HasSuffix(collection, ".$cmd") {
		return common.IsMaster(ctx, query.Query)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3008

	// database name typically is either "$external" or "admin"

	if cmd == "saslStart" && strings.HasSuffix(collection, ".$cmd") {
		var emptyPayload types.Binary
		return &wire.OpReply{
			NumberReturned: 1,
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"conversationId", int32(1),
				"done", true,
				"payload", emptyPayload,
				"ok", float64(1),
			))},
		}, nil
	}

	return nil, handlererrors.NewCommandErrorMsgWithArgument(
		handlererrors.ErrNotImplemented,
		fmt.Sprintf("CmdQuery: unhandled command %q for collection %q", cmd, collection),
		"OpQuery: "+cmd,
	)
}
