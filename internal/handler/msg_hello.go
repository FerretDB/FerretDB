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
	"errors"
	"strings"
	"time"

	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgHello implements `hello` command.
func (h *Handler) MsgHello(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	doc, err := msg.Document()
	if err != nil {
		return nil, err
	}

	if err := common.CheckClientMetadata(ctx, doc); err != nil {
		return nil, lazyerrors.Error(err)
	}

	saslSupportedMechs, err := common.GetOptionalParam(doc, "saslSupportedMechs", "")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	resp := must.NotFail(types.NewDocument(
		"isWritablePrimary", true,
		"maxBsonObjectSize", int32(h.MaxBsonObjectSizeBytes),
		"maxMessageSizeBytes", int32(wire.MaxMsgLen),
		"maxWriteBatchSize", int32(100000),
		"localTime", time.Now(),
		"connectionId", int32(42),
		"minWireVersion", common.MinWireVersion,
		"maxWireVersion", common.MaxWireVersion,
		"readOnly", false,
	))

	if saslSupportedMechs == "" {
		resp.Set("ok", float64(1))
		must.NoError(reply.SetSections(wire.MakeOpMsgSection(resp)))

		return &reply, nil
	}

	db, username, ok := strings.Cut(saslSupportedMechs, ".")
	if !ok {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrBadValue,
			"UserName must contain a '.' separated database.user pair",
		)
	}

	mechs := []string{"PLAIN"}
	if h.EnableNewAuth {
		mechs, err = h.getUserSupportedMechs(ctx, db, username)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	saslSupportedMechsResp := must.NotFail(types.NewArray())
	for _, k := range mechs {
		saslSupportedMechsResp.Append(k)
	}

	if saslSupportedMechsResp.Len() != 0 {
		resp.Set("saslSupportedMechs", saslSupportedMechsResp)
	}

	resp.Set("ok", float64(1))
	must.NoError(reply.SetSections(wire.MakeOpMsgSection(resp)))

	return &reply, nil
}

// getUserSupportedMechs for a given user.
func (h *Handler) getUserSupportedMechs(ctx context.Context, db, username string) ([]string, error) {
	adminDB, err := h.b.Database("admin")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	usersCol, err := adminDB.Collection("system.users")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	filter, err := usersInfoFilter(false, false, db, []usersInfoPair{
		{username: username, db: db},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

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

		if !matches {
			continue
		}

		if v.Has("credentials") {
			credentials := must.NotFail(v.Get("credentials")).(*types.Document)
			return credentials.Keys(), nil
		}
	}

	return nil, nil
}
