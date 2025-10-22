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
	"fmt"

	"github.com/AlekSi/lazyerrors"
	"github.com/FerretDB/wire/wirebson"

	"github.com/FerretDB/FerretDB/v2/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/v2/internal/handler/middleware"
	"github.com/FerretDB/FerretDB/v2/internal/mongoerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// msgKillCursors implements `killCursors` command.
//
// The passed context is canceled when the client connection is closed.
func (h *Handler) msgKillCursors(connCtx context.Context, req *middleware.Request) (*middleware.Response, error) {
	doc := req.Document()

	command := doc.Command()

	db, err := getRequiredParam[string](doc, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := getRequiredParam[string](doc, command)
	if err != nil {
		return nil, err
	}

	username := conninfo.Get(connCtx).Conv().Username()

	userID, _, err := h.s.CreateOrUpdateByLSID(connCtx, doc)
	if err != nil {
		return nil, err
	}

	cursorsV, err := getRequiredParamAny(doc, "cursors")
	if err != nil {
		return nil, err
	}

	curArr, ok := cursorsV.(wirebson.AnyArray)
	if !ok {
		msg := fmt.Sprintf(`required parameter "cursors" has type %T (expected array)`, cursorsV)
		return nil, lazyerrors.Error(mongoerrors.NewWithArgument(mongoerrors.ErrBadValue, msg, command))
	}

	cursors, err := curArr.Decode()
	if !ok {
		return nil, lazyerrors.Error(err)
	}

	var ids []int64
	cursorsKilled := wirebson.MustArray()
	cursorsNotFound := wirebson.MustArray()
	cursorsAlive := wirebson.MustArray()
	cursorsUnknown := wirebson.MustArray()

	for i := range cursors.Len() {
		v := cursors.Get(i)

		id, ok := v.(int64)
		if !ok {
			return nil, mongoerrors.NewWithArgument(
				mongoerrors.ErrTypeMismatch,
				fmt.Sprintf(
					"BSON field 'killCursors.cursors.%d' is the wrong type '%s', expected type 'long'",
					i,
					aliasFromType(v),
				),
				command,
			)
		}

		ids = append(ids, id)
	}

	for _, id := range ids {
		// Should we check database and collection names?
		// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/17
		_, _, _ = db, collection, username

		if err = h.s.DeleteCursor(userID, id, db); err != nil {
			return nil, err
		}

		if deleted := h.p.KillCursor(connCtx, id); !deleted {
			must.NoError(cursorsNotFound.Add(id))
			continue
		}

		must.NoError(cursorsKilled.Add(id))
	}

	return middleware.ResponseDoc(req, wirebson.MustDocument(
		"cursorsKilled", cursorsKilled,
		"cursorsNotFound", cursorsNotFound,
		"cursorsAlive", cursorsAlive,
		"cursorsUnknown", cursorsUnknown,
		"ok", float64(1),
	))
}
