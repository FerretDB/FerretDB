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

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// authenticate validates the user's credentials if new auth is enabled.
func authenticate(ctx context.Context, msg *wire.OpMsg, h *Handler) error {
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

	username, _ := conninfo.Get(ctx).Auth()

	// NOTE: do we need to handle user who have access to all databases?
	filter := must.NotFail(types.NewDocument("_id", dbName+"."+username))

	// Filter isn't being passed to the query as we are filtering after retrieving all data
	// from the database due to limitations of the internal/backends filters.
	qr, err := usersCol.Query(ctx, nil)
	if err != nil {
		return lazyerrors.Error(err)
	}

	defer qr.Iter.Close()

	var storedUser *types.Document

	for {
		var v *types.Document
		_, v, err = qr.Iter.Next()

		if errors.Is(err, iterator.ErrIteratorDone) {
			break
		}

		if err != nil {
			return lazyerrors.Error(err)
		}

		var matches bool

		if matches, err = common.FilterDocument(v, filter); err != nil {
			return lazyerrors.Error(err)
		}

		if matches {
			storedUser = v
			break
		}
	}

	if storedUser == nil {
		return handlererrors.NewCommandErrorMsg(
			handlererrors.ErrUserNotFound,
			fmt.Sprintf("User %s@%s not found", username, dbName),
		)
	}

	// authenticate user

	return nil
}
