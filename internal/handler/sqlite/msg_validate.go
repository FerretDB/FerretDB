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
	"fmt"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/handler/common"
	"github.com/FerretDB/FerretDB/internal/handler/commonerrors"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgValidate implements `validate` command.
func (h *Handler) MsgValidate(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "full", "repair", "metadata")

	command := document.Command()

	dbName, err := common.GetRequiredParam[string](document, "$db")
	if err != nil {
		return nil, err
	}

	collection, err := common.GetRequiredParam[string](document, command)
	if err != nil {
		return nil, err
	}

	db, err := h.b.Database(dbName)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(collection)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	_, err = c.Stats(ctx, &backends.CollectionStatsParams{Refresh: true})
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
			msg := fmt.Sprintf("Collection '%s.%s' does not exist to validate.", dbName, collection)
			return nil, commonerrors.NewCommandErrorMsgWithArgument(commonerrors.ErrNamespaceNotFound, msg, document.Command())
		}

		return nil, lazyerrors.Error(err)
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"ns", dbName+"."+collection,
			"nInvalidDocuments", int32(0),
			"nNonCompliantDocuments", int32(0),
			"nrecords", int32(-1), // TODO https://github.com/FerretDB/FerretDB/issues/419
			"nIndexes", int32(1), // TODO https://github.com/FerretDB/FerretDB/issues/419
			"valid", true,
			"repaired", false,
			"warnings", types.MakeArray(0),
			"errors", types.MakeArray(0),
			"extraIndexEntries", types.MakeArray(0),
			"missingIndexEntries", types.MakeArray(0),
			"corruptRecords", types.MakeArray(0),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
