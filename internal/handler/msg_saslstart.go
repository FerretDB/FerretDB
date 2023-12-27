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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

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

	var response []byte

	var conv *scram.ServerConversation

	plain := true

	switch mechanism {
	case "PLAIN":
		username, password, err = saslStartPlain(document)
		if err != nil {
			return nil, err
		}

	case "SCRAM-SHA-256":
		// TODO fix invalid-proof in SCRAM conversation
		response, conv, err = saslStartSCRAM(document)
		if err != nil {
			return nil, err
		}

		plain = false

		h.L.Debug(
			"saslStart",
			zap.String("response", string(response)),
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

	if !plain {
		conninfo.Get(ctx).SetConv(conv)
		h.L.Debug(
			"conninfo",
			zap.Bool("valid", conninfo.Get(ctx).Conv().Valid()),
			zap.Bool("done", conninfo.Get(ctx).Conv().Done()),
			zap.String("username", conninfo.Get(ctx).Conv().Username()),
		)

		d.Set("payload", types.Binary{
			Subtype: types.BinarySubtype(0),
			B:       response,
		})
		d.Set("done", false)
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

func saslStartSCRAM(doc *types.Document) ([]byte, *scram.ServerConversation, error) {
	var payload []byte

	binaryPayload, err := common.GetRequiredParam[types.Binary](doc, "payload")
	if err == nil {
		payload = binaryPayload.B
	}

	// TODO store the credentials in the 'admin.system.users' namespace eventually
	// 'SCRAM-SHA-256': {
	//     iterationCount: 15000, or at least 4096
	//     salt: '/EWnFeM5z6vZbsviI9N+DpThFxjrDhryf47cTA==',
	//     storedKey: 'jQGI8FtQZjfe/MyyaiYT8m0GlF7KxqvH5+EHhxYtXyo=',
	//     serverKey: 'I9O3QjHz++JGp4vrD79P7m+af1oXPPziZ8sTlauQEwI='
	// }

	var response string

	// generate server-first-message of the form r=client-nonce|server-nonce,s=user-salt,i=iteration-count
	salt := make([]byte, 28)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, nil, err
	}

	cl := scram.CredentialLookup(func(s string) (scram.StoredCredentials, error) {
		kf := scram.KeyFactors{
			Salt:  string(salt),
			Iters: 4096,
		}

		// https://github.com/xdg-go/scram/blob/17629a50d5ce12875d83f9095809ae43b765c303/server_conv.go#L143 the hashes are not equal
		return scram.StoredCredentials{KeyFactors: kf}, nil
	})

	ss, err := scram.SHA256.NewServer(cl)
	must.NoError(err)

	conv := ss.NewConversation()
	response, err = conv.Step(string(payload))
	must.NoError(err)

	return []byte(response), conv, nil
}
