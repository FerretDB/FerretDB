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

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/handler/handlerparams"
	"github.com/FerretDB/FerretDB/internal/handler/users"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/password"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgUpdateUser implements `updateUser` command.
func (h *Handler) MsgUpdateUser(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	username, err := common.GetRequiredParam[string](document, document.Command())
	if err != nil {
		return nil, err
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

	if _, err = common.GetOptionalParam[*types.Array](document, "roles", nil); err != nil {
		var ce *handlererrors.CommandError
		if errors.As(err, &ce) && ce.Code() == handlererrors.ErrBadValue {
			return nil, handlererrors.NewCommandErrorMsg(
				handlererrors.ErrMissingField,
				"BSON field 'updateUser.roles' is missing but a required field",
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

	common.Ignored(document, h.L, "writeConcern", "authenticationRestrictions", "comment")

	defMechanisms := must.NotFail(types.NewArray("SCRAM-SHA-1", "SCRAM-SHA-256"))

	mechanisms, err := common.GetOptionalParam(document, "mechanisms", defMechanisms)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	iter := mechanisms.Iterator()
	defer iter.Close()

	for {
		var v any
		_, v, err = iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		switch v {
		case "SCRAM-SHA-1", "SCRAM-SHA-256":
			// do nothing
		default:
			return nil, handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				fmt.Sprintf("Unknown auth mechanism '%s'", v),
			)
		}
	}

	var credentials *types.Document

	if document.Has("pwd") {
		pwd, _ := document.Get("pwd")
		userPassword, ok := pwd.(string)

		if !ok {
			return nil, handlererrors.NewCommandErrorMsg(
				handlererrors.ErrTypeMismatch,
				fmt.Sprintf("BSON field 'updateUser.pwd' is the wrong type '%s', expected type 'string'",
					handlerparams.AliasFromType(pwd),
				),
			)
		}

		if userPassword == "" {
			return nil, handlererrors.NewCommandErrorMsg(
				handlererrors.ErrSetEmptyPassword,
				"Password cannot be empty",
			)
		}

		credentials, err = users.MakeCredentials(username, password.WrapPassword(userPassword), mechanisms)
		if err != nil {
			return nil, err
		}
	}

	adminDB, err := h.b.Database("admin")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	usersCol, err := adminDB.Collection("system.users")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	filter, err := usersInfoFilter(false, false, "", []usersInfoPair{{username: username, db: dbName}})
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

	var saved *types.Document

	for {
		var v *types.Document
		_, v, err = qr.Iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		var matches bool
		matches, err = common.FilterDocument(v, filter)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		if matches {
			saved = v
			break
		}
	}

	if saved == nil {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrUserNotFound,
			fmt.Sprintf("User %s@%s not found", username, dbName),
		)
	}

	var changes bool

	if credentials != nil {
		changes = true

		saved.Set("credentials", credentials)
	}

	if !changes {
		return nil, handlererrors.NewCommandErrorMsg(
			handlererrors.ErrBadValue,
			"Must specify at least one field to update in updateUser",
		)
	}

	_, err = usersCol.UpdateAll(ctx, &backends.UpdateAllParams{Docs: []*types.Document{saved}})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.MakeOpMsgSection(
		must.NotFail(types.NewDocument(
			"ok", float64(1),
		)),
	)))

	return &reply, nil
}
