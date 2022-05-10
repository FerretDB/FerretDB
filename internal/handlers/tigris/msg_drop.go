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

package tigris

import (
	"context"

	api "github.com/tigrisdata/tigris-client-go/api/server/v1"
	"google.golang.org/grpc/codes"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgDrop removes a collection or view from the database.
func (h *Handler) MsgDrop(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.l, "writeConcern", "comment")

	command := document.Command()

	var db, collection string
	if db, err = common.GetRequiredParam[string](document, "$db"); err != nil {
		return nil, err
	}

	if collection, err = common.GetRequiredParam[string](document, command); err != nil {
		return nil, err
	}

	tigrisDB := h.client.conn.UseDatabase(db)
	if err = tigrisDB.DropCollection(ctx, collection); err != nil {
		switch err := err.(type) {
		case *api.TigrisError:
			// TODO: database not found DatabaseNotFound error
			// is hidden in codes.InvalidArgument due to same gRPC status codes
			if err.Code == codes.NotFound ||
				err.Code == codes.InvalidArgument { // DatabaseNotFound
				return nil, common.NewErrorMsg(common.ErrNamespaceNotFound, "ns not found")
			}
		default:
			return nil, lazyerrors.Error(err)
		}
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"nIndexesWas", int32(1), // TODO
			"ns", db+"."+collection,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
