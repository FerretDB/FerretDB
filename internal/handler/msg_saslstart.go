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
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
	"github.com/xdg-go/scram"
	"go.uber.org/zap"
)

// MsgSASLStart implements `saslStart` command.
func (h *Handler) MsgSASLStart(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3008

	// database name typically is either "$external" or "admin"
	// we can't use it to query the database
	_ = dbName

	mechanism, err := common.GetRequiredParam[string](document, "mechanism")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var username, password string

	var response string

	plain := true

	switch {
	case mechanism == "PLAIN":
		username, password, err = saslStartPlain(document)
		if err != nil {
			return nil, err
		}

	case mechanism == "SCRAM-SHA-256":
		// TODO finish SCRAM negotiation
		response, err = saslStartSCRAM(document)
		if err != nil {
			return nil, err
		}

		plain = false

	// to reduce connection overhead time, clients may use a hello command to complete their authentication exchange
	// if so, the saslStart command may be embedded under the speculativeAuthenticate field
	case document.Has("speculativeAuthenticate"):
		// TODO finish SCRAM negotiation
		response, err = saslStartSCRAM(document)
		if err != nil {
			return nil, err
		}

		plain = false

	default:
		msg := fmt.Sprintf("Unsupported authentication mechanism %q.\n", mechanism) +
			"See https://docs.ferretdb.io/security/authentication/ for more details."
		return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrAuthenticationFailed, msg, "mechanism")
	}

	// {"payload": "n,,n=username,r=xv/n+51PvakMHhHjIa8va/hXQnzZ/n3W"}
	h.L.Debug(
		"SCRAM", zap.String("response", response),
	)

	conninfo.Get(ctx).SetAuth(username, password)

	var emptyPayload types.Binary
	var reply wire.OpMsg
	d := must.NotFail(types.NewDocument(
		"conversationId", int32(1),
		"done", true,
		"payload", emptyPayload,
		"ok", float64(1),
	))

	if !plain {
		d.Set("payload", response)
	}

	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{d},
	}))

	return &reply, nil
}

// saslStartPlain extracts username and password from PLAIN `saslStart` payload.
func saslStartPlain(doc *types.Document) (string, string, error) {
	var payload []byte

	// some drivers send payload as a string
	stringPayload, err := common.GetRequiredParam[string](doc, "payload")
	if err == nil {
		if payload, err = base64.StdEncoding.DecodeString(stringPayload); err != nil {
			return "", "", handlererrors.NewCommandErrorMsgWithArgument(
				handlererrors.ErrBadValue,
				fmt.Sprintf("Invalid payload: %v", err),
				"payload",
			)
		}
	}

	// most drivers follow spec and send payload as a binary
	binaryPayload, err := common.GetRequiredParam[types.Binary](doc, "payload")
	if err == nil {
		payload = binaryPayload.B
	}

	// as spec's payload should be binary, we return an error mentioned binary as expected type
	if payload == nil {
		return "", "", err
	}

	parts := bytes.Split(payload, []byte{0})
	if l := len(parts); l != 3 {
		return "", "", handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrTypeMismatch,
			fmt.Sprintf("Invalid payload: expected 3 parts, got %d", l),
			"payload",
		)
	}

	authzid, authcid, passwd := parts[0], parts[1], parts[2]

	// Some drivers (Go) send empty authorization identity (authzid),
	// while others (Java) set it to the same value as authentication identity (authcid)
	// (see https://www.rfc-editor.org/rfc/rfc4616.html).
	// Ignore authzid for now.
	_ = authzid

	return string(authcid), string(passwd), nil
}

func saslStartSCRAM(doc *types.Document) (string, error) {
	var payload []byte

	binaryPayload, err := common.GetRequiredParam[types.Binary](doc, "payload")
	if err == nil {
		payload = binaryPayload.B
	}

	client, err := scram.SHA256.NewClient("", "", "")
	if err != nil {
		return "", err
	}

	// 1. "client-first-message" the client sends the username for lookup
	firstMsg, err := client.NewConversation().Step(string(payload))
	if err != nil {
		return "", err
	}

	// 2. the server sends a "server-first-message" containing the salt, iteration, StoredKey, and ServerKey
	// 3. the client responds with the "client-final-message" containing the ClientProof
	// 4. the server verifies the nonce and the proof

	return firstMsg, nil
}
