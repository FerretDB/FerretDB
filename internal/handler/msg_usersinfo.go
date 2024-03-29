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

	// TODO https://github.com/FerretDB/FerretDB/issues/4141
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
		"showCustomData", "showPrivileges",
		"showAuthenticationRestrictions", "comment", "filter",
	)

	var (
		users    []usersInfoPair
		allDBs   bool // allDBs set to true means we want users from all databases
		singleDB bool // singleDB set to true means we want users from a single database (when usersInfo: 1)
	)

	showCredentials, err := common.GetOptionalParam(document, "showCredentials", false)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	switch user := usersInfo.(type) {
	case *types.Document:
		if user.Has("forAllDBs") {
			allDBs = true
			break
		}

		var u usersInfoPair
		if err = u.extract(user, dbName); err != nil {
			return nil, lazyerrors.Error(err)
		}

		users = append(users, u)
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
	case string:
		var u usersInfoPair
		if err = u.extract(user, dbName); err != nil {
			return nil, lazyerrors.Error(err)
		}

		users = append(users, u)
	case int32, int64: // {usersInfo: 1 }
		singleDB = true
	default:
		return nil, handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrBadValue,
			"UserName must be either a string or an object",
			document.Command(),
		)
	}

	adminDB, err := h.b.Database("admin")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	usersCol, err := adminDB.Collection("system.users")
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	filter, err := usersInfoFilter(allDBs, singleDB, dbName, users)
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

	var res *types.Array
	res, err = types.NewArray()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

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

		if v.Has("credentials") {
			credentials := must.NotFail(v.Get("credentials")).(*types.Document)
			if credentialsKeys := credentials.Keys(); len(credentialsKeys) > 0 {
				mechanisms := must.NotFail(types.NewArray())
				for _, k := range credentialsKeys {
					mechanisms.Append(k)
				}

				v.Set("mechanisms", mechanisms)
			}
		}

		if !showCredentials {
			v.Remove("credentials")
		}

		if matches {
			res.Append(v)
		}
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.MakeOpMsgSection(
		must.NotFail(types.NewDocument(
			"users", res,
			"ok", float64(1),
		)),
	)))

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
			return handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				"UserName must contain a field named: user",
			)
		}

		var ok bool
		p.username, ok = ui.(string)

		if !ok {
			return handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				fmt.Sprintf("UserName must contain a string field named: user. But, has type %T", ui),
			)
		}

		db, err := vt.Get("db")
		if err != nil {
			return handlererrors.NewCommandErrorMsg(
				handlererrors.ErrBadValue,
				"UserName must contain a field named: db",
			)
		}

		if db != nil {
			p.db, ok = db.(string)
			if !ok {
				return handlererrors.NewCommandErrorMsg(
					handlererrors.ErrBadValue,
					fmt.Sprintf("UserName must contain a string field named: db. But, has type %T", db),
				)
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

// usersInfoFilter returns a filter for usersInfo command.
//
// When allDBs is true, it returns a filter for all databases.
// When singleDB is true, it returns a filter for a single database (case when usersInfo: 1 is invoked).
// Otherwise, it filters by any pair of user and database.
func usersInfoFilter(allDBs, singleDB bool, dbName string, pairs []usersInfoPair) (*types.Document, error) {
	filter := must.NotFail(types.NewDocument())

	if allDBs {
		return filter, nil
	}

	if singleDB {
		filter.Set("db", must.NotFail(types.NewDocument("$eq", dbName)))
		return filter, nil
	}

	ps := []any{}
	for _, p := range pairs {
		ps = append(ps, p.db+"."+p.username)
	}

	ids, err := types.NewArray(ps...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	filter.Set("_id", must.NotFail(types.NewDocument("$in", ids)))

	return filter, nil
}
