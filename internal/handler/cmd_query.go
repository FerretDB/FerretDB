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
	"slices"
	"strings"

	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// CmdQuery implements deprecated OP_QUERY message handling.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) CmdQuery(connCtx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	q := query.Query()
	cmd := q.Command()
	collection := query.FullCollectionName

	suffix := ".$cmd"
	if !strings.HasSuffix(collection, suffix) {
		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/527
		return wire.NewOpReply(must.NotFail(wirebson.NewDocument(
			"$err", "OP_QUERY is no longer supported. The client driver may require an upgrade.",
			"code", int32(mongoerrors.ErrLocation5739101),
			"ok", float64(0),
		)))
	}

	if query.NumberToReturn != 1 && query.NumberToReturn != -1 {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrLocation16979,
			fmt.Sprintf("Bad numberToReturn (%d) for $cmd type ns - can only be 1 or -1", query.NumberToReturn),
			"OpQuery: "+cmd,
		)
	}

	switch cmd {
	case "hello", "ismaster", "isMaster":
		reply, err := h.hello(connCtx, q, h.TCPHost, h.ReplSetName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return wire.NewOpReply(must.NotFail(reply.Encode()))

	case "saslStart":
		if slices.Contains(q.FieldNames(), "$db") {
			return nil, mongoerrors.NewWithArgument(
				mongoerrors.ErrLocation40621,
				"$db is not allowed in OP_QUERY requests",
				"OpQuery: "+cmd,
			)
		}

		reply, err := h.saslStart(connCtx, q)
		if err != nil {
			return nil, err
		}

		must.NoError(reply.Add("ok", float64(1)))

		return wire.NewOpReply(must.NotFail(reply.Encode()))

	case "saslContinue":
		if slices.Contains(q.FieldNames(), "$db") {
			return nil, mongoerrors.NewWithArgument(
				mongoerrors.ErrLocation40621,
				"$db is not allowed in OP_QUERY requests",
				"OpQuery: "+cmd,
			)
		}

		reply, err := h.saslContinue(connCtx, q)
		if err != nil {
			return nil, err
		}

		return wire.NewOpReply(must.NotFail(reply.Encode()))
	}

	return nil, mongoerrors.NewWithArgument(
		mongoerrors.ErrUnsupportedOpQueryCommand,
		fmt.Sprintf("Unsupported OP_QUERY command: %s. The client driver may require an upgrade.", cmd),
		"OpQuery: "+cmd,
	)
}
