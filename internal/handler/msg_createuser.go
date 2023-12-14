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

	// https://www.mongodb.com/docs/manual/reference/command/createUser/

	username, err := common.GetRequiredParam[string](document, document.Command())
	if err != nil {
		return nil, err
	}

	_, err = common.GetOptionalParam[string](document, "pwd", "")
	if err != nil {
		return nil, err
	}

	if err := common.UnimplementedNonDefault(document, "roles", func(v any) bool {
		roles, ok := v.(*types.Array)
		return ok && roles.Len() == 0
	}); err != nil {
		return nil, err
	}

	common.Ignored(document, h.L, "roles", "writeConcern", "authenticationRestrictions", "mechanisms")

	saved := must.NotFail(types.NewDocument(
		"user", username,
		"roles", must.NotFail(types.NewArray()), // Non-default value is currently ignored.
		"pwd", "password", // TODO: hash the password.
	))

	if document.Has("customData") {
		customData, err := common.GetOptionalParam[*types.Document](document, "customData", nil)
		if err != nil {
			return nil, err
		}
		saved.Set("customData", customData)
	}

	if document.Has("digestPassword") {
		digestPassword, err := common.GetOptionalParam[bool](document, "digestPassword", true)
		if err != nil {
			return nil, err
		}
		saved.Set("digestPassword", digestPassword)
	}

	if document.Has("comment") {
		saved.Set("comment", must.NotFail(document.Get("comment")))
	}

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	if dbName != "$external" && !document.Has("pwd") {
		return nil, handlererrors.NewCommandErrorMsg(handlererrors.ErrBadValue, "Must provide a 'pwd' field for all user documents, except those with '$external' as the user's source db")
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	collection, err := db.Collection("system.users")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, err = collection.InsertAll(ctx, &backends.InsertAllParams{
		Docs: []*types.Document{saved},
	})
	if err != nil {
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
