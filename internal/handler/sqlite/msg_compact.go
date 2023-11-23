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
	"github.com/FerretDB/FerretDB/internal/handler/commonparams"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgCompact implements `compact` command.
func (h *Handler) MsgCompact(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	common.Ignored(document, h.L, "comment")

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
		if backends.ErrorCodeIs(err, backends.ErrorCodeDatabaseNameIsInvalid) {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidNamespace,
				fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collection),
				command,
			)
		}

		return nil, lazyerrors.Error(err)
	}

	c, err := db.Collection(collection)
	if err != nil {
		if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionNameIsInvalid) {
			return nil, commonerrors.NewCommandErrorMsgWithArgument(
				commonerrors.ErrInvalidNamespace,
				fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collection),
				command,
			)
		}

		return nil, lazyerrors.Error(err)
	}

	var force bool

	if v, _ := document.Get("force"); v != nil {
		if force, err = commonparams.GetBoolOptionalParam("force", v); err != nil {
			return nil, err
		}
	}

	statsBefore, err := c.Stats(ctx, new(backends.CollectionStatsParams))
	if backends.ErrorCodeIs(err, backends.ErrorCodeCollectionDoesNotExist) {
		return nil, commonerrors.NewCommandErrorMsgWithArgument(
			commonerrors.ErrNamespaceNotFound,
			fmt.Sprintf("Invalid namespace specified '%s.%s'", dbName, collection),
			command,
		)
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if _, err = c.Compact(ctx, &backends.CompactParams{Full: force}); err != nil {
		return nil, lazyerrors.Error(err)
	}

	statsAfter, err := c.Stats(ctx, new(backends.CollectionStatsParams))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	bytesFreed := statsBefore.SizeTotal - statsAfter.SizeTotal

	// There's a possibility that the size of a collection might be greater at the
	// end of a compact operation if the collection is being actively written to at
	// the time of compaction.
	if bytesFreed < 0 {
		bytesFreed = 0
	}

	var reply wire.OpMsg
	must.NoError(reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"bytesFreed", float64(bytesFreed),
			"ok", float64(1),
		))},
	}))

	return &reply, nil
}
