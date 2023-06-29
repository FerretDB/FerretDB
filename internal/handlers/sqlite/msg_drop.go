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

package sqlite

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/handlers/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDrop implements HandlerInterface.
func (h *Handler) MsgDrop(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "writeConcern", "comment")

	command := document.Command()

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collectionName, err := common.GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	// This is mainly needed for SQLite and our cursor tests to avoid SQLITE_BUSY errors,
	// but fine for other backends too.
	//
	// There is a race condition: another client could create a new cursor for that collection
	// after we closed all of them, but before we drop the collection itself.
	// In that case, we expect the client to retry the operation.
	for _, c := range h.cursors.All() {
		if c.DB == dbName && c.Collection == collectionName {
			c.Close()
		}
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	defer db.Close()

	err = db.DropCollection(ctx, &backends.DropCollectionParams{
		Name: collectionName,
	})

	switch {
	case err == nil:
		var reply wire.OpMsg
		must.NoError(reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"nIndexesWas", int32(1), // TODO
				"ns", dbName+"."+collectionName,
				"ok", float64(1),
			))},
		}))

		return &reply, nil

	case backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist):
		return nil, commonerrors.NewCommandErrorMsg(commonerrors.ErrNamespaceNotFound, "ns not found")

	default:
		return nil, lazyerrors.Error(err)
	}
}
