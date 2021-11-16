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
	"fmt"
	"sort"
	"strings"

	"github.com/MangoDB-io/MangoDB/internal/handlers/common"
	"github.com/MangoDB-io/MangoDB/internal/types"
	"github.com/MangoDB-io/MangoDB/internal/wire"
)

func (h *Handler) MsgListDatabases(ctx context.Context, msg *wire.OpMsg) (*wire.OpMsg, error) {
	var names []string
	rows, err := h.pgPool.Query(ctx, "SELECT schema_name FROM information_schema.schemata")
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

		names = append(names, name)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	sort.Strings(names)

	dbs := make(types.Array, len(names))
	for i, n := range names {
		totalRelationSize, err := h.pgPool.Query(ctx, fmt.Sprintf("SELECT pg_total_relation_size(%s)", n))
		if err != nil {
			return nil, err
		}

		sizeOnDisk, err := totalRelationSize.Values()
		if err != nil {
			return nil, err
		}

		// TODO return true if there are not collections
		var empty bool

		dbs[i] = types.MustMakeDocument(
			"name", n,
			"sizeOnDisk", sizeOnDisk[0].(int64),
			"empty", empty,
		)
	}

	databaseSize, err := h.pgPool.Query(ctx, "SELECT pg_database_size(current_database())")
	if err != nil {
		return nil, err
	}

	totalSize, err := databaseSize.Values()
	if err != nil {
		return nil, err
	}

	var reply wire.OpMsg
	err = reply.SetSections(wire.OpMsgSection{
		Documents: []types.Document{types.MustMakeDocument(
			"databases", dbs,
			"totalSize", totalSize[0].(int64),
			"totalSizeMb", totalSize[0].(int64)>>20,
			"ok", float64(1),
		)},
	})
	if err != nil {
		return nil, common.NewError(common.ErrInternalError, err)
	}

	return &reply, nil
}
