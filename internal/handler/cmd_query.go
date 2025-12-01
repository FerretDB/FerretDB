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

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// CmdQuery implements deprecated OP_QUERY message handling.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) CmdQuery(connCtx context.Context, query *middleware.Request) (*middleware.Response, error) {
	q := query.Document()
	cmd := q.Command()
	queryBody := query.WireBody().(*wire.OpQuery)
	collection := queryBody.FullCollectionName
	toReturn := queryBody.NumberToReturn

	suffix := ".$cmd"
	if !strings.HasSuffix(collection, suffix) {
		// special case for legacy reply: $err instead of errmsg, no codeName
		return middleware.ResponseDoc(query, wirebson.MustDocument(
			"$err", "OP_QUERY is no longer supported. The client driver may require an update.",
			"code", int32(mongoerrors.ErrLocation5739101),
			"ok", float64(0),
		))
	}

	if toReturn != 1 && toReturn != -1 {
		return nil, mongoerrors.NewWithArgument(
			mongoerrors.ErrLocation16979,
			fmt.Sprintf("Bad numberToReturn (%d) for $cmd type ns - can only be 1 or -1", toReturn),
			"OpQuery: "+cmd,
		)
	}

	switch cmd {
	case "hello", "ismaster", "isMaster":
		reply, err := h.hello(connCtx, q, h.TCPHost, h.ReplSetName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return middleware.ResponseDoc(query, reply)

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

		return middleware.ResponseDoc(query, reply)

	case "saslContinue":
		if q.Get("$db") != nil {
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

		return middleware.ResponseDoc(query, reply)
	}

	return nil, mongoerrors.NewWithArgument(
		mongoerrors.ErrUnsupportedOpQueryCommand,
		fmt.Sprintf("Unsupported OP_QUERY command: %s. The client driver may require an update.", cmd),
		"OpQuery: "+cmd,
	)
}
