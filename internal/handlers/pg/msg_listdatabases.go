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

package pg

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListDatabases implements HandlerInterface.
func (h *Handler) MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	var filter *types.Document
	if filter, err = common.GetOptionalParam(document, "filter", filter); err != nil {
		return nil, err
	}

	common.Ignored(document, h.l, "comment", "authorizedDatabases")

	databaseNames, err := h.pgPool.Schemas(ctx)
	if err != nil {
		return nil, err
	}

	nameOnly, err := common.GetBoolOptionalParam(document, "nameOnly")
	if err != nil {
		return nil, err
	}

	databases := types.MakeArray(len(databaseNames))
	for _, databaseName := range databaseNames {
		tables, err := h.pgPool.Tables(ctx, databaseName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		// iterate over result to collect sizes
		var sizeOnDisk int64
		for _, name := range tables {
			var tableSize int64
			fullName := databaseName + "." + name
			err = h.pgPool.QueryRow(ctx, "SELECT pg_total_relation_size($1)", fullName).Scan(&tableSize)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}
			sizeOnDisk += tableSize
		}

		d := must.NotFail(types.NewDocument(
			"name", databaseName,
			"sizeOnDisk", sizeOnDisk,
			"empty", sizeOnDisk == 0,
		))

		matches, err := common.FilterDocument(d, filter)
		if err != nil {
			return nil, err
		}

		if matches {
			if nameOnly {
				d = must.NotFail(types.NewDocument(
					"name", databaseName,
				))
			}
			if err = databases.Append(d); err != nil {
				return nil, lazyerrors.Error(err)
			}
		}
	}

	if nameOnly {
		var reply wire.OpMsg
		err = reply.SetSections(wire.OpMsgSection{
			Documents: []*types.Document{must.NotFail(types.NewDocument(
				"databases", databases,
				"ok", float64(1),
			))},
		})
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return &reply, nil
	}

	var totalSize int64
	err = h.pgPool.QueryRow(ctx, "SELECT pg_database_size(current_database())").Scan(&totalSize)
	if err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{must.NotFail(types.NewDocument(
			"databases", databases,
			"totalSize", totalSize,
			"totalSizeMb", totalSize/1024/1024,
			"ok", float64(1),
		))},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
