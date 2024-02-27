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
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// CmdQuery implements deprecated OP_QUERY message handling.
func (h *Handler) CmdQuery(ctx context.Context, query *wire.OpQuery) (*wire.OpReply, error) {
	q := query.Query()
	cmd := q.Command()
	collection := query.FullCollectionName

	v, _ := q.Get("speculativeAuthenticate")
	if v != nil && (cmd == "ismaster" || cmd == "isMaster" || cmd == "hello") {
		reply, err := common.IsMaster(ctx, q, h.TCPHost, h.ReplSetName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		replyDoc := must.NotFail(reply.Document())

		document := v.(*types.Document)

		doc, err := h.speculativeAuthenticate(ctx, document)
		if err == nil {
			// speculative authenticate response field is only set if the authentication is successful,
			// for an unsuccessful authentication, saslStart will return an error
			replyDoc.Set("speculativeAuthenticate", doc)
		}

		reply.SetDocument(replyDoc)

		return reply, nil
	}

	if (cmd == "ismaster" || cmd == "isMaster") && strings.HasSuffix(collection, ".$cmd") {
		return common.IsMaster(ctx, query.Query(), h.TCPHost, h.ReplSetName)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3008

	// database name typically is either "$external" or "admin"

	if cmd == "saslStart" && strings.HasSuffix(collection, ".$cmd") {
		var emptyPayload types.Binary
		var reply wire.OpReply
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

// speculativeAuthenticate uses db and mechanism to authenticate and returns the document
// to assign for op query speculativeAuthenticate response field if the authentication is successful.
func (h *Handler) speculativeAuthenticate(ctx context.Context, document *types.Document) (*types.Document, error) {
	dbName, err := common.GetRequiredParam[string](document, "db")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	mechanism, err := common.GetRequiredParam[string](document, "mechanism")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	switch mechanism {
	case "PLAIN":
		username, password, err := saslStartPlain(document)
		if err != nil {
			return nil, err
		}

		if h.EnableNewAuth {
			conninfo.Get(ctx).SetBypassBackendAuth()
		}

		conninfo.Get(ctx).SetAuth(username, password)

		var emptyPayload types.Binary

		return must.NotFail(types.NewDocument(
			"conversationId", int32(1),
			"done", true,
			"payload", emptyPayload,
		)), nil
	case "SCRAM-SHA-1", "SCRAM-SHA-256":
		if !h.EnableNewAuth {
			return nil, handlererrors.NewCommandErrorMsg(
				handlererrors.ErrAuthenticationFailed,
				"SCRAM authentication is not enabled",
			)
		}

		response, err := h.saslStartSCRAM(ctx, dbName, mechanism, document)
		if err != nil {
			return nil, err
		}

		conninfo.Get(ctx).SetBypassBackendAuth()

		binResponse := types.Binary{
			B: []byte(response),
		}

		return must.NotFail(types.NewDocument(
			"conversationId", int32(1),
			"done", false,
			"payload", binResponse,
		)), nil
	default:
		return nil, lazyerrors.Errorf("unsupported mechanism %s", mechanism)
	}
}
