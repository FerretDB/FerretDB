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
	"log/slog"

	"github.com/FerretDB/wire"
	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/internal/bson"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/logging"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// MsgSASLStart implements `saslStart` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) MsgSASLStart(connCtx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := bson.OpMsgDocument(msg)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	replyDoc, err := h.saslStart(connCtx, dbName, document)
	if err != nil {
		return nil, err
	}

	replyDoc.Set("ok", float64(1))

	return bson.NewOpMsg(replyDoc)
}

// saslStart starts authentication and returns a document used for the response.
// If EnableNewAuth is set SCRAM mechanisms are supported, otherwise `PLAIN` mechanism is supported.
func (h *Handler) saslStart(ctx context.Context, dbName string, document *types.Document) (*types.Document, error) {
	// TODO https://github.com/FerretDB/FerretDB/issues/3008
	mechanism, err := common.GetRequiredParam[string](document, "mechanism")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if h.EnableNewAuth {
		switch mechanism {
		case "SCRAM-SHA-1", "SCRAM-SHA-256":
			var response string

			if response, err = h.saslStartSCRAM(ctx, dbName, mechanism, document); err != nil {
				return nil, err
			}

			return must.NotFail(types.NewDocument(
				"conversationId", int32(1),
				"done", false,
				"payload", types.Binary{B: []byte(response)},
			)), nil
		default:
			msg := fmt.Sprintf("Unsupported authentication mechanism %q.\n", mechanism) +
				"See https://docs.ferretdb.io/security/authentication/ for more details."
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrAuthenticationFailed, msg, "mechanism")
		}
	}

	switch mechanism {
	case "PLAIN":
		if err = saslStartPlain(ctx, dbName, document); err != nil {
			return nil, err
		}

		var emptyPayload types.Binary

		return must.NotFail(types.NewDocument(
			"conversationId", int32(1),
			"done", true,
			"payload", emptyPayload,
		)), nil
	default:
		msg := fmt.Sprintf("Unsupported authentication mechanism %q.\n", mechanism) +
			"See https://docs.ferretdb.io/security/authentication/ for more details."
		return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrAuthenticationFailed, msg, "mechanism")
	}
}

// saslStartPlain extracts username and password from PLAIN `saslStart` payload.
func saslStartPlain(ctx context.Context, dbName string, doc *types.Document) error {
	var payload []byte

	// some drivers send payload as a string
	stringPayload, err := common.GetRequiredParam[string](doc, "payload")
	if err == nil {
		if payload, err = base64.StdEncoding.DecodeString(stringPayload); err != nil {
			return handlererrors.NewCommandErrorMsgWithArgument(
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
		return err
	}

	fields := bytes.Split(payload, []byte{0})
	if l := len(fields); l != 3 {
		return handlererrors.NewCommandErrorMsgWithArgument(
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

	conninfo.Get(ctx).SetAuth(string(authcid), string(passwd), nil, dbName)

	return nil
}

// scramCredentialLookup looks up an user's credentials in the database.
func (h *Handler) scramCredentialLookup(ctx context.Context, dbName, username, mechanism string) (*scram.StoredCredentials, error) { //nolint:lll // for readability
	adminDB, err := h.b.Database("admin")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	usersCol, err := adminDB.Collection("system.users")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/174
	filter := must.NotFail(types.NewDocument("_id", dbName+"."+username))

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

	h.L.WarnContext(ctx, "scramCredentialLookup: failed", slog.String("user", username))

	return nil, handlererrors.NewCommandErrorMsgWithArgument(
		handlererrors.ErrAuthenticationFailed,
		"Authentication failed.",
		"scramCredentialLookup",
	)
}

// saslStartSCRAM extracts the initial challenge and attempts to move the
// authentication conversation forward returning a challenge response.
func (h *Handler) saslStartSCRAM(ctx context.Context, dbName, mechanism string, doc *types.Document) (string, error) {
	var payload []byte

	// most drivers follow spec and send payload as a binary
	binaryPayload, err := common.GetRequiredParam[types.Binary](doc, "payload")
	if err != nil {
		return "", err
	}

	payload = binaryPayload.B

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
		cred, lookupErr := h.scramCredentialLookup(ctx, dbName, username, mechanism)
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

	attrs := []any{
		slog.String("username", conv.Username()),
		slog.Bool("valid", conv.Valid()),
		slog.Bool("done", conv.Done()),
	}

	if err != nil {
		if h.L.Enabled(ctx, slog.LevelDebug) {
			attrs = append(attrs, logging.Error(err))
		}

		h.L.WarnContext(ctx, "saslStartSCRAM: step failed", attrs...) //nolint:sloglint // attrs is not key-value pairs

		return "", err
	}

	h.L.DebugContext(ctx, "saslStartSCRAM: step succeed", attrs...) //nolint:sloglint // attrs is not key-value pairs

	conninfo.Get(ctx).SetAuth(conv.Username(), "", conv, dbName)

	return response, nil
}
