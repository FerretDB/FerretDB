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

package handlers

import (
	"context"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/wire"
)

// MsgListDatabases command provides a list of all existing databases along with basic statistics about them.
func (h *Handler) MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = common.UnimplementedNonDefault(document, "filter", func(v any) bool {
		d, ok := v.(*types.Document)
		return ok && d.Len() == 0
	}); err != nil {
		return nil, err
	}

	// TODO https://github.com/FerretDB/FerretDB/issues/301
	// if err = common.UnimplementedNonDefault(document, "nameOnly", func(v any) bool {
	// 	nameOnly, ok := v.(bool)
	// 	return ok && !nameOnly
	// }); err != nil {
	// 	return nil, err
	// }

	common.Ignored(document, h.l, "comment", "authorizedDatabases")

	databaseNames, err := h.pgPool.Schemas(ctx)
	if err != nil {
		return nil, err
	}

	databases := types.MakeArray(len(databaseNames))
	for _, databaseName := range databaseNames {
		tables, _, err := h.pgPool.Tables(ctx, databaseName)
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

		d := types.MustNewDocument(
			"name", databaseName,
			"sizeOnDisk", sizeOnDisk,
			"empty", sizeOnDisk == 0,
		)
		if err = databases.Append(d); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var totalSize int64
	err = h.pgPool.QueryRow(ctx, "SELECT pg_database_size(current_database())").Scan(&totalSize)
	if err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []*types.Document{types.MustNewDocument(
			"databases", databases,
			"totalSize", totalSize,
			"totalSizeMb", totalSize/1024/1024,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &reply, nil
}
