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
	"errors"
	"fmt"

	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgSASLStart implements `saslStart` command.
func (h *Handler) MsgSASLStart(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, err = common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3008

	mechanism, err := common.GetRequiredParam[string](document, "mechanism")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var (
		reply wire.OpMsg

		username, password string
	)

	switch mechanism {
	case "PLAIN":
		username, password, err = saslStartPlain(document)
		if err != nil {
			return nil, err
		}

		conninfo.Get(ctx).SetAuth(username, password)

		var emptyPayload types.Binary
		must.NoError(reply.SetSections(wire.MakeOpMsgSection(
			must.NotFail(types.NewDocument(
				"conversationId", int32(1),
				"done", true,
				"payload", emptyPayload,
				"ok", float64(1),
			)),
		)))

	case "SCRAM-SHA-1", "SCRAM-SHA-256":
		response, err := h.saslStartSCRAM(ctx, mechanism, document)
		if err != nil {
			return nil, err
		}

		conninfo.Get(ctx).SetBypassBackendAuth()

		binResponse := types.Binary{
			B: []byte(response),
		}

		must.NoError(reply.SetSections(wire.MakeOpMsgSection(
			must.NotFail(types.NewDocument(
				"ok", float64(1),
				"conversationId", int32(1),
				"done", false,
				"payload", binResponse,
			)),
		)))

	default:
		msg := fmt.Sprintf("Unsupported authentication mechanism %q.\n", mechanism) +
			"See https://docs.ferretdb.io/security/authentication/ for more details."
		return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrAuthenticationFailed, msg, "mechanism")
	}

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

// scramCredentialLookup looks up an user's credentials in the database.
func (h *Handler) scramCredentialLookup(ctx context.Context, username, dbName, mechanism string) (
	*scram.StoredCredentials, error,
) {
	adminDB, err := h.b.Database("admin")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	usersCol, err := adminDB.Collection("system.users")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var filter *types.Document

	filter, err = usersInfoFilter(false, false, "", []usersInfoPair{
		{username: username, db: dbName},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// Filter isn't being passed to the query as we are filtering after retrieving all data
	// from the database due to limitations of the internal/backends filters.
	qr, err := usersCol.Query(ctx, nil)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer qr.Iter.Close()

	for {
		_, v, err := qr.Iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		matches, err := common.FilterDocument(v, filter)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if matches {
			credentials := must.NotFail(v.Get("credentials")).(*types.Document)

			if !credentials.Has(mechanism) {
				return nil, handlererrors.NewCommandErrorMsgWithArgument(
					handlererrors.ErrMechanismUnavailable,
					fmt.Sprintf(
						"Unable to use %s based authentication for user without any %s credentials registered",
						mechanism,
						mechanism,
					),
					mechanism,
				)
			}

			cred := must.NotFail(credentials.Get(mechanism)).(*types.Document)

			salt := must.NotFail(base64.StdEncoding.DecodeString(must.NotFail(cred.Get("salt")).(string)))
			storedKey := must.NotFail(base64.StdEncoding.DecodeString(must.NotFail(cred.Get("storedKey")).(string)))
			serverKey := must.NotFail(base64.StdEncoding.DecodeString(must.NotFail(cred.Get("serverKey")).(string)))

			return &scram.StoredCredentials{
				KeyFactors: scram.KeyFactors{
					Salt:  string(salt),
					Iters: int(must.NotFail(cred.Get("iterationCount")).(int32)),
				},
				StoredKey: storedKey,
				ServerKey: serverKey,
			}, nil
		}
	}

	return nil, handlererrors.NewCommandErrorMsg(
		handlererrors.ErrAuthenticationFailed,
		"Authentication failed.",
	)
}

// saslStartSCRAM extracts the initial challenge and attempts to move the
// authentication conversation forward returning a challenge response.
func (h *Handler) saslStartSCRAM(ctx context.Context, mechanism string, doc *types.Document) (string, error) {
	var payload []byte

	// most drivers follow spec and send payload as a binary
	binaryPayload, err := common.GetRequiredParam[types.Binary](doc, "payload")
	if err != nil {
		return "", err
	}

	payload = binaryPayload.B

	dbName, err := common.GetRequiredParam[string](doc, "$db")
	if err != nil {
		return "", err
	}

	var f scram.HashGeneratorFcn

	switch mechanism {
	case "SCRAM-SHA-1":
		f = scram.SHA1
	case "SCRAM-SHA-256":
		f = scram.SHA256
	default:
		panic("unsupported SCRAM mechanism")
	}

	scramServer, err := f.NewServer(func(username string) (scram.StoredCredentials, error) {
		cred, lookupErr := h.scramCredentialLookup(ctx, username, dbName, mechanism)
		if lookupErr != nil {
			return scram.StoredCredentials{}, lookupErr
		}

		return *cred, nil
	})
	if err != nil {
		return "", err
	}

	conv := scramServer.NewConversation()

	response, err := conv.Step(string(payload))
	if err != nil {
		return "", err
	}

	conninfo.Get(ctx).SetConv(conv)

	return response, nil
}
