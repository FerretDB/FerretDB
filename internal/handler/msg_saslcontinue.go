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

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
	"go.uber.org/zap"
)

// MsgSASLContinue implements `saslContinue` command.
func (h *Handler) MsgSASLContinue(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var payload []byte

	binaryPayload, err := common.GetRequiredParam[types.Binary](document, "payload")
	if err == nil {
		payload = binaryPayload.B
	}

	sconv := conninfo.Get(ctx).Conv()

	adminDB, err := h.b.Database("admin")
	must.NoError(err)

	users, err := adminDB.Collection("system.users")
	must.NoError(err)

	q, err := users.Query(ctx, &backends.QueryParams{
		Filter: must.NotFail(types.NewDocument(
			"user", sconv.Conv.Username(),
		)),
		Limit: int64(1), // assume there's only 'test.username' user for now
	})
	must.NoError(err)

	var doc *types.Document

	defer q.Iter.Close()

	for {
		_, v, err := q.Iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		doc = v
	}

	scramCredentials, err := doc.GetByPath(
		types.Path{}.Append("credentials").Append(sconv.Mechanism),
	)

	if scramCredentials == nil {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrMechanismUnavailable,
			fmt.Sprintf("Unable to use %s based authentication for user without any %s credentials registered", sconv.Mechanism, sconv.Mechanism),
			sconv.Conv.Username(),
		)
	}

	if err != nil {
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrAuthenticationFailed,
			fmt.Sprintf("Unable to validate %s authentication due to corrupted stored credentials", sconv.Mechanism),
			sconv.Mechanism,
		)
	}

	storedKey := must.NotFail(scramCredentials.(*types.Document).Get("storedKey"))
	decodedStoredKey, err := base64.StdEncoding.DecodeString(storedKey.(string))
	must.NoError(err)

	// just match storedKey for now
	if !bytes.Equal(sconv.StoredKey, decodedStoredKey) {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrAuthenticationFailed,
			"SCRAM authentication failed, storedKey mismatch",
		)
	}

	response, err := sconv.Conv.Step(string(payload))
	must.NoError(err)

	h.L.Debug(
		"saslContinue",
		zap.String("payload", string(payload)),
		zap.String("response", response),
		zap.String("user", sconv.Conv.Username()),
		zap.Bool("authenticated", sconv.Conv.Valid()),
	)

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"conversationId", int32(1),
			"done", true,
			"payload", response,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
