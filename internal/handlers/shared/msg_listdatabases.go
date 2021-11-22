// Copyright 2021 Baltoro OÜ.
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

package shared

import (
	"context"
	"strings"

	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/util/lazyerrors"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

func (h *Handler) MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	document, err := msg.Document()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	m := document.Map()

	var databaseNames []string

	db, ok := m["$db"].(string)
	if !ok {
		// collect MangoDB databases / PostgreSQL schema names
		rows, err := h.pgPool.Query(ctx, "SELECT schema_name FROM information_schema.schemata ORDER BY schema_name")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			if err = rows.Scan(&name); err != nil {
				return nil, err
			}

			if strings.HasPrefix(name, "pg_") || name == "information_schema" {
				continue
			}

			databaseNames = append(databaseNames, name)
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
	} else {
		databaseNames = append(databaseNames, db)
	}

	databases := make(types.Array, len(databaseNames))
	for i, databaseName := range databaseNames {
		// get database collections / schema tables
		rows, err := h.pgPool.Query(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = $1", databaseName)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		defer rows.Close()

		// iterate over result to collect sizes
		var sizeOnDisk int64
		for rows.Next() {
			var name string
			if err = rows.Scan(&name); err != nil {
				return nil, lazyerrors.Error(err)
			}

			var tableSize int64
			fullName := databaseName + "." + name
			err = h.pgPool.QueryRow(ctx, "SELECT pg_total_relation_size($1)", fullName).Scan(&tableSize)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			sizeOnDisk += tableSize
		}
		if err = rows.Err(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		databases[i] = types.MustMakeDocument(
			"name", databaseName,
			"sizeOnDisk", sizeOnDisk,
			"empty", sizeOnDisk == 0,
		)
	}

	var totalSize int64
	err = h.pgPool.QueryRow(ctx, "SELECT pg_database_size(current_database())").Scan(&totalSize)
	if err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
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
