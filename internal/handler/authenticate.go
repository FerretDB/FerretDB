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
	"net"
	"net/netip"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/password"
)

// authenticate validates users with stored credentials in admin.systems.user.
// If EnableNewAuth is false or bypass backend auth is set false, it succeeds
// authentication and let backend handle it.
//
// When admin.systems.user contains no user, and the client is connected from
// the localhost, it bypasses credentials check.
func (h *Handler) authenticate(ctx context.Context) error {
	if !h.EnableNewAuth {
		return nil
	}

	conninfo.Get(ctx).BypassBackendAuth()

	adminDB, err := h.b.Database("admin")
	if err != nil {
		return lazyerrors.Error(err)
	}

	usersCol, err := adminDB.Collection("system.users")
	if err != nil {
		return lazyerrors.Error(err)
	}

	username, userPassword := conninfo.Get(ctx).Auth()

	// For `PLAIN` mechanism $db field is always `$external` upon saslStart.
	// For `SCRAM-SHA-1` and `SCRAM-SHA-256` mechanisms $db field contains
	// authSource option of the client.
	// Let authorization handle the database access right.
	// TODO https://github.com/FerretDB/FerretDB/issues/174
	filter := must.NotFail(types.NewDocument("user", username))

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
		host, _, err := net.SplitHostPort(conninfo.Get(ctx).PeerAddr)
		if err != nil {
			return lazyerrors.Error(err)
		}

		if netip.MustParseAddr(host).IsLoopback() {
			// bypass credentials check when no user is found and the client is connected from the localhost
			return nil
		}
	}

	if storedUser == nil {
		return handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrAuthenticationFailed,
			"Authentication failed",
			"authenticate",
		)
	}

	credentials := must.NotFail(storedUser.Get("credentials")).(*types.Document)

	switch {
	case credentials.Has("SCRAM-SHA-256"), credentials.Has("SCRAM-SHA-1"):
		// SCRAM calls back scramCredentialLookup each time Step is called,
		// and that checks the authentication.
		return nil
	case credentials.Has("PLAIN"):
		break
	default:
		panic("credentials does not contain a known mechanism")
	}

	v := must.NotFail(credentials.Get("PLAIN"))

	doc, ok := v.(*types.Document)
	if !ok {
		return lazyerrors.Errorf("field 'PLAIN' has type %T, expected Document", v)
	}

	err = password.PlainVerify(userPassword, doc)
	if err != nil {
		return handlererrors.NewCommandErrorMsgWithArgument(
			handlererrors.ErrAuthenticationFailed,
			"Authentication failed",
			"authenticate",
		)
	}

	return nil
}
