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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgUsersInfo implements `usersInfo` command.
func (h *Handler) MsgUsersInfo(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	usersInfo, err := document.Get(document.Command())
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if usersInfo == nil {
		msg := fmt.Sprintf("required parameter %q is missing", document.Command())
		return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrBadValue, msg, document.Command())
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/3784
	// TODO https://github.com/FerretDB/FerretDB/issues/3777
	// TODO https://github.com/FerretDB/FerretDB/issues/3785
	if err = common.UnimplementedNonDefault(document, "filter", func(v any) bool {
		if v == nil || v == types.Null {
			return true
		}

		cd, ok := v.(*types.Document)
		return ok && cd.Len() == 0
	}); err != nil {
		return nil, err
	}

	common.Ignored(
		document, h.L,
		"showCredentials", "showCustomData", "showPrivileges",
		"showAuthenticationRestrictions", "comment", "filter",
	)

	var users []usersInfoPair

	switch user := usersInfo.(type) {
	case *types.Document, string:
		var u usersInfoPair
		if err = u.extract(user, dbName); err != nil {
			return nil, lazyerrors.Error(err)
		}

		users = append(users, u)
	case int: // {usersInfo: 1 }
		break
	case *types.Array:
		for i := 0; i < user.Len(); i++ {
			var ui any
			ui, err = user.Get(i)

			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			if ui != nil {
				var u usersInfoPair

				if err = u.extract(ui, dbName); err != nil {
					return nil, lazyerrors.Error(err)
				}

				users = append(users, u)
			}
		}
	default:
		msg := fmt.Sprintf("required parameter %q has unexpected type %T", document.Command(), usersInfo)
		return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrBadValue, msg, document.Command())
	}

	// Users are saved in the "admin" database.
	adminDB, err := h.b.Database("admin")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	usersCol, err := adminDB.Collection("system.users")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	qr, err := usersCol.Query(ctx, &backends.QueryParams{
		Filter: usersInfoQueryFilter(users),
	})

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer qr.Iter.Close()

	var res []*types.Document

	for {
		_, v, err := qr.Iter.Next()
		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		} else if err != nil {
			return nil, lazyerrors.Error(err)
		}
		res = append(res, v)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"users", res,
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}

// usersInfoPair is a pair of username and database name.
type usersInfoPair struct {
	username string
	db       string
}

// extract username and db from v.
func (p *usersInfoPair) extract(v any, dbName string) error {
	switch vt := v.(type) {
	case *types.Document:
		ui, err := vt.Get("user")
		if err != nil {
			return lazyerrors.Error(err)
		}

		var ok bool
		p.username, ok = ui.(string)

		if !ok {
			return lazyerrors.Errorf("unexpected type %T for username", ui)
		}

		db, err := vt.Get("db")

		if err != nil {
			return lazyerrors.Error(err)
		}

		if db != nil {
			p.db, ok = db.(string)
			if !ok {
				return lazyerrors.Errorf("unexpected type %T for db", db)
			}
		}

		return nil
	case string:
		p.username = vt
		p.db = dbName

		return nil
	default:
		return lazyerrors.Errorf("unexpected type %T", vt)
	}
}

func usersInfoQueryFilter(u []usersInfoPair) *types.Document {
	doc := must.NotFail(types.NewDocument())
	return doc
}
