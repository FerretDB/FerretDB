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

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/handlererrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDrop implements `drop` command.
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

	// Most backends would block on `DropCollection` below otherwise.
	//
	// There is a race condition: another client could create a new cursor for that collection
	// after we closed all of them, but before we drop the collection itself.
	// In that case, we expect the client to wait or to retry the operation.
	for _, c := range h.cursors.All() {
		if c.DB == dbName && c.Collection == collectionName {
			h.cursors.CloseAndRemove(c)
		}
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			msg := fmt.Sprintf("Invalid namespace specified '%s'", dbName)
			return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "drop")
		}

		return nil, lazyerrors.Error(err)
	}

	err = db.DropCollection(ctx, &backends.DropCollectionParams{
		Name: collectionName,
	})

	switch {
	case err == nil, backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist):
		var reply wire.OpMsg
		must.NoError(reply.SetSections(wire.MakeOpMsgSection(
			must.NotFail(types.NewDocument(
				"nIndexesWas", int32(1), // TODO https://github.com/FerretDB/FerretDB/issues/2337
				"ns", dbName+"."+collectionName,
				"ok", float64(1),
			)),
		)))

		return &reply, nil

	case backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid):
		msg := fmt.Sprintf("Invalid collection name: %s", collectionName)
		return nil, handlererrors.NewCommandErrorMsgWithArgument(handlererrors.ErrInvalidNamespace, msg, "drop")

	default:
		return nil, lazyerrors.Error(err)
	}
}
