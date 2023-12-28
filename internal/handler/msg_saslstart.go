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

	var conv *scram.ServerConversation

	plain := true

	switch mechanism {
	case "PLAIN":
		username, password, err = saslStartPlain(document)
		if err != nil {
			return nil, err
		}

	case "SCRAM-SHA-256":
		// FIXME nonce received did not match nonce sent
		response, conv, err = saslStartSCRAM(document)
		if err != nil {
			return nil, err
		}

		plain = false

		h.L.Debug(
			"saslStart",
			zap.String("response", response),
		)

	default:
		msg := fmt.Sprintf("Unsupported authentication mechanism %q.\n", mechanism) +
			"See https://docs.ferretdb.io/security/authentication/ for more details."
		return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrAuthenticationFailed, msg, "mechanism")
	}

	conninfo.Get(ctx).SetAuth(username, password)

	var emptyPayload types.Binary
	var reply wire.OpMsg
	d := must.NotFail(types.NewDocument(
		"conversationId", int32(1),
		"done", false,
		"payload", emptyPayload,
		"ok", float64(1),
	))

	// TODO confirm if this is even needed or if speculativeAuthenticate is always used and is sent in an OP_QUERY
	if !plain {
		conninfo.Get(ctx).SetConv(conv)
		h.L.Debug(
			"conninfo",
			zap.Bool("valid", conninfo.Get(ctx).Conv().Valid()),
			zap.Bool("done", conninfo.Get(ctx).Conv().Done()),
			zap.String("username", conninfo.Get(ctx).Conv().Username()),
		)

		d.Set("payload", response)
		d.Set("done", false)

		if document.Has("speculativeAuthenticate") {
			d.Remove("conversationId")
			d.Remove("done")
			d.Remove("payload")

			// create a speculative conversation document for SCRAM authentication
			d.Set("speculativeAuthenticate", must.NotFail(
				types.NewDocument(
					"conversationId", int32(1),
					"done", false,
					"payload", response,
					"ok", float64(1),
				),
			))
		}
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

	fields := bytes.Split(payload, []byte{0})
	if l := len(fields); l != 3 {
		return "", "", handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrTypeMismatch,
			fmt.Sprintf("Invalid payload: expected 3 fields, got %d", l),
			"payload",
		)
	}

	authzid, authcid, passwd := fields[0], fields[1], fields[2]

	// Some drivers (Go) send empty authorization identity (authzid),
	// while others (Java) set it to the same value as authentication identity (authcid)
	// (see https://www.rfc-editor.org/rfc/rfc4616.html).
	// Ignore authzid for now.
	_ = authzid

	return string(authcid), string(passwd), nil
}

func saslStartSCRAM(doc *types.Document) (string, *scram.ServerConversation, error) {
	var payload []byte

	binaryPayload, err := common.GetRequiredParam[types.Binary](doc, "payload")
	if err == nil {
		payload = binaryPayload.B
	}

	// TODO store the credentials in the 'admin.system.users' namespace eventually
	// "SCRAM-SHA-256" : {
	// 	"iterationCount" : 15000,
	// 	"salt" : "0KPzHubY95siEZZQC6jYuupOqlkUjlqBFGK24w==",
	// 	"storedKey" : "tTiad1PQvXqtOWTgxwF1/Cks7niAzRLzP91vGd4VWuQ=",
	// 	"serverKey" : "h8hefOPW7nRxDNrrCGvYU9PdAEx/1oKEcCRdg3LGGRo="
	// }

	var (
		salt      = []byte{238, 53, 185, 100, 231, 51, 143, 78, 79, 227, 12, 141, 115, 109, 78, 138, 66, 46, 74, 88, 143, 55, 218, 240, 226, 193, 40, 25}
		storedKey = []byte{23, 200, 83, 46, 185, 217, 177, 203, 174, 179, 55, 235, 135, 238, 39, 186, 156, 163, 60, 14, 52, 114, 159, 160, 127, 60, 181, 30, 199, 55, 59, 119}
		serverKey = []byte{119, 131, 254, 119, 205, 67, 223, 85, 199, 194, 247, 208, 3, 120, 240, 129, 57, 164, 138, 246, 95, 93, 48, 255, 156, 16, 18, 155, 190, 195, 194, 253}
	)

	var response string

	// generate server-first-message of the form r=client-nonce|server-nonce,s=user-salt,i=iteration-count
	// salt := make([]byte, 28)
	// if _, err := io.ReadFull(rand.Reader, salt); err != nil {
	// 	return nil, nil, err
	// }

	cl := scram.CredentialLookup(func(s string) (scram.StoredCredentials, error) {
		kf := scram.KeyFactors{
			Salt:  string(salt),
			Iters: 15000,
		}

		return scram.StoredCredentials{
			KeyFactors: kf,
			StoredKey:  storedKey,
			ServerKey:  serverKey,
		}, nil
	})

	ss, err := scram.SHA256.NewServer(cl)
	must.NoError(err)

	conv := ss.NewConversation()
	response, err = conv.Step(string(payload))
	must.NoError(err)

	return response, conv, nil
}
