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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/password"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// authenticate validates the user's credentials in the connection with the
// credentials in the database. If EnableNewAuth is false, it does nothing.
//
// When admin.systems.user contains no user, the authentication succeeds until
// the first user is created.
func (h *Handler) authenticate(ctx context.Context, msg *wire.OpMsg) error {
	if !h.EnableNewAuth {
		return nil
	}

	adminDB, err := h.b.Database("admin")
	if err != nil {
		return lazyerrors.Error(err)
	}

	usersCol, err := adminDB.Collection("system.users")
	if err != nil {
		return lazyerrors.Error(err)
	}

	document, err := msg.Document()
	if err != nil {
		return lazyerrors.Error(err)
	}

	var dbName string

	if dbName, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return err
	}

	username, pwd := conninfo.Get(ctx).Auth()

	// NOTE: how does a user with access to all database look like?
	filter := must.NotFail(types.NewDocument("_id", dbName+"."+username))

	qr, err := usersCol.Query(ctx, nil)
	if err != nil {
		return lazyerrors.Error(err)
	}

	defer qr.Iter.Close()

	var storedUser *types.Document

	var hasUser bool

	for {
		var v *types.Document
		_, v, err = qr.Iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return lazyerrors.Error(err)
		}

		hasUser = true

		var matches bool

		if matches, err = common.FilterDocument(v, filter); err != nil {
			return lazyerrors.Error(err)
		}

		if matches {
			storedUser = v
			break
		}
	}

	if !hasUser {
		// an exception where authentication is skipped until the first user is created.
		return nil
	}

	credentials := must.NotFail(storedUser.Get("credentials")).(*types.Document)
	if !credentials.Has("PLAIN") {
		return handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrAuthenticationFailed,
			"TODO: wrong authentication mechanism",
			"PLAIN",
		)
	}

	err = password.PlainVerify(pwd, credentials)
	if err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
