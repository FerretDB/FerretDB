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

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// CmdQuery implements deprecated OP_QUERY message handling.
func (h *Handler) CmdQuery(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	q := query.Query()
	cmd := q.Command()
	collection := query.FullCollectionName

	if !strings.HasSuffix(collection, ".$cmd") {
		return wire.NewOpReply(must.NotFail(bson.NewDocument(
			"$err", "OP_QUERY is no longer supported. The client driver may require an upgrade.",
			"code", int32(handlererrors.ErrOpQueryCollectionSuffixMissing),
			"ok", float64(0),
		)))
	}

	switch cmd {
	case "hello", "ismaster", "isMaster":
		reply, err := h.hello(ctx, q, h.TCPHost, h.ReplSetName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		v := q.Get("speculativeAuthenticate")
		if v == nil {
			return wire.NewOpReply(must.NotFail(reply.Encode()))
		}

		docV, ok := v.(bson.AnyDocument)
		if !ok {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrTypeMismatch,
				fmt.Sprintf("speculativeAuthenticate type wrong; expected: document; got: %T", v),
				"OpQuery: "+q.Command(),
			)
		}

		authDoc, err := docV.Decode()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		dbName, err := getRequiredParam[string](authDoc, "db")
		if err != nil {
			h.SLOG.DebugContext(ctx, "No `db` in `speculativeAuthenticate`", logging.Error(err))

			return wire.NewOpReply(must.NotFail(reply.Encode()))
		}

		speculativeAuthenticate, err := h.saslStart(ctx, dbName, authDoc)
		if err != nil {
			h.SLOG.DebugContext(ctx, "Speculative authentication failed", logging.Error(err))

			// unsuccessful speculative authentication leave `speculativeAuthenticate` field unset
			// and let `saslStart` return an error
			return wire.NewOpReply(must.NotFail(reply.Encode()))
		}

		must.NoError(reply.Add("speculativeAuthenticate", speculativeAuthenticate))

		// saslSupportedMechs is used by the client as default mechanisms if `mechanisms` is unset
		must.NoError(reply.Add("saslSupportedMechs", must.NotFail(bson.NewArray("SCRAM-SHA-1", "SCRAM-SHA-256"))))

		return wire.NewOpReply(must.NotFail(reply.Encode()))
	}

	return nil, handlererrors.NewCommandErrorMsgWithArgument(
		handlererrors.ErrUnsupportedOpQueryCommand,
		fmt.Sprintf("Unsupported OP_QUERY command: %s. The client driver may require an upgrade.", cmd),
		"OpQuery: "+cmd,
	)
}
