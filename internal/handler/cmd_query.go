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

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// CmdQuery implements deprecated OP_QUERY message handling.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) CmdQuery(connCtx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	q, err := bson.ToDocument(query.Query())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	cmd := q.Command()
	collection := query.FullCollectionName

	suffix := ".$cmd"
	if !strings.HasSuffix(collection, suffix) {
		return wire.NewOpReply(must.NotFail(bson.FromDocument(must.NotFail(types.NewDocument(
			"$err", "OP_QUERY is no longer supported. The client driver may require an upgrade.",
			"code", int32(handlererrors.ErrOpQueryCollectionSuffixMissing),
			"ok", float64(0),
		)))))
	}

	db := strings.TrimSuffix(collection, suffix)

	if query.NumberToReturn != 1 && query.NumberToReturn != -1 {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrBadNumberToReturn,
			fmt.Sprintf("Bad numberToReturn (%d) for $cmd type ns - can only be 1 or -1", query.NumberToReturn),
			"OpQuery: "+cmd,
		)
	}

	switch cmd {
	case "hello", "ismaster", "isMaster":
		var reply *types.Document
		reply, err = h.hello(connCtx, q, h.TCPHost, h.ReplSetName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		v, _ := q.Get("speculativeAuthenticate")
		if v == nil {
			return wire.NewOpReply(must.NotFail(bson.FromDocument(reply)))
		}

		authDoc, ok := v.(*types.Document)
		if !ok {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrTypeMismatch,
				fmt.Sprintf("speculativeAuthenticate type wrong; expected: document; got: %T", v),
				"OpQuery: "+q.Command(),
			)
		}

		dbName, err := common.GetRequiredParam[string](authDoc, "db")
		if err != nil {
			h.L.DebugContext(connCtx, "No `db` in `speculativeAuthenticate`", logging.Error(err))

			return wire.NewOpReply(must.NotFail(bson.FromDocument(reply)))
		}

		speculativeAuthenticate, err := h.saslStart(connCtx, dbName, authDoc)
		if err != nil {
			h.L.DebugContext(connCtx, "Speculative authentication failed", logging.Error(err))

			// unsuccessful speculative authentication leave `speculativeAuthenticate` field unset
			// and let `saslStart` return an error
			return wire.NewOpReply(must.NotFail(bson.FromDocument(reply)))
		}

		reply.Set("speculativeAuthenticate", speculativeAuthenticate)

		// ok field is the last field
		reply.Remove("ok")
		reply.Set("ok", float64(1))

		h.L.DebugContext(connCtx, "Speculative authentication passed")

		return wire.NewOpReply(must.NotFail(bson.FromDocument(reply)))
	case "saslContinue":
		if slices.Contains(q.Keys(), "$db") {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrOpQueryInvalidField,
				"$db is not allowed in OP_QUERY requests",
				"OpQuery: "+cmd,
			)
		}

		var reply *types.Document

		if reply, err = h.saslContinue(connCtx, q); err != nil {
			return nil, err
		}

		return wire.NewOpReply(must.NotFail(bson.FromDocument(reply)))
	case "saslStart":
		if slices.Contains(q.Keys(), "$db") {
			return nil, handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrOpQueryInvalidField,
				"$db is not allowed in OP_QUERY requests",
				"OpQuery: "+cmd,
			)
		}

		var reply *types.Document

		if reply, err = h.saslStart(connCtx, db, q); err != nil {
			return nil, err
		}

		reply.Set("ok", float64(1))

		return wire.NewOpReply(must.NotFail(bson.FromDocument(reply)))
	}

	return nil, handlererrors.NewCommandErrorMsgWithArgument(
		handlererrors.ErrUnsupportedOpQueryCommand,
		fmt.Sprintf("Unsupported OP_QUERY command: %s. The client driver may require an upgrade.", cmd),
		"OpQuery: "+cmd,
	)
}
