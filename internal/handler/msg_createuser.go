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
	"fmt"

	"github.com/google/uuid"
	"github.com/xdg-go/scram"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgCreateUser implements `createUser` command.
func (h *Handler) MsgCreateUser(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3777
	// TODO https://github.com/FerretDB/FerretDB/issues/3778
	// TODO https://github.com/FerretDB/FerretDB/issues/3784
	if dbName != "$external" && !document.Has("pwd") {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrBadValue,
			"Must provide a 'pwd' field for all user documents, except those with '$external' as the user's source db",
		)
	}

	var username string
	username, err = common.GetRequiredParam[string](document, document.Command())

	if err != nil {
		return nil, err
	}

	if username == "" {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrBadValue,
			"User document needs 'user' field to be non-empty",
		)
	}

	if err = common.UnimplementedNonDefault(document, "customData", func(v any) bool {
		if v == nil || v == types.Null {
			return true
		}

		cd, ok := v.(*types.Document)
		return ok && cd.Len() == 0
	}); err != nil {
		return nil, err
	}

	if _, err = common.GetRequiredParam[*types.Array](document, "roles"); err != nil {
		var ce *handlererrors.CommandError
		if errors.As(err, &ce) && ce.Code() == handlererrors.ErrBadValue {
			return nil, handlererrors.NewCommandErrorMsg(
				handlererrors.ErrMissingField,
				"BSON field 'createUser.roles' is missing but a required field",
			)
		}

		return nil, lazyerrors.Error(err)
	}

	if err = common.UnimplementedNonDefault(document, "roles", func(v any) bool {
		r, ok := v.(*types.Array)
		return ok && r.Len() == 0
	}); err != nil {
		return nil, err
	}

	if err = common.UnimplementedNonDefault(document, "digestPassword", func(v any) bool {
		if v == nil || v == types.Null {
			return true
		}

		dp, ok := v.(bool)
		return ok && dp
	}); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "writeConcern", "authenticationRestrictions", "comment")

	mechanisms, err := document.Get("mechanisms")
	must.NoError(err)

	supportedMechanisms := map[string]struct{}{
		"PLAIN":         {},
		"SCRAM-SHA-256": {},
	}

	for _, mechanism := range mechanisms.([]string) {
		if _, ok := supportedMechanisms[mechanism]; !ok {
			return nil, lazyerrors.Errorf("unsupported mechanism %s", mechanism)
		}
	}

	pwd, err := document.Get("pwd")
	must.NoError(err)

	// ignore the authzID
	scramClient, err := scram.SHA256.NewClient(username, pwd.(string), "")
	must.NoError(err)

	// XXX local authorization stores users and their respective roles in the database
	// we most likely need to generate the SHA-1 and SHA-256 credentials ourselves
	// instead of relying on an authentication exchange
	_ = scramClient

	id := uuid.New()
	saved := must.NotFail(types.NewDocument(
		"_id", dbName+"."+username,
		"credentials", types.MakeDocument(0),
		"user", username,
		"db", dbName,
		"roles", types.MakeArray(0),
		"userId", types.Binary{Subtype: types.BinaryUUID, B: must.NotFail(id.MarshalBinary())},
	))

	// Users are saved in the "admin" database.
	adminDB, err := h.b.Database("admin")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	users, err := adminDB.Collection("system.users")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, err = users.InsertAll(ctx, &backends.InsertAllParams{
		Docs: []*types.Document{saved},
	})
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeInsertDuplicateID) {
			return nil, handlererrors.NewCommandErrorMsg(
				handlererrors.ErrUserAlreadyExists,
				fmt.Sprintf("User \"%s@%s\" already exists", username, dbName),
			)
		}

		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
