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

// Package postgresql provides backend for PostgreSQL and compatible databases.
//
// # Design principles
//
//  1. Metadata is heavily cached to avoid most queries and transactions.
package postgresql

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/FerretDB/FerretDB/internal/backends/postgresql/metadata"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// stats represents information about statistics of tables and indexes.
type stats struct {
	countDocuments  int64
	sizeIndexes     int64
	sizeTables      int64
	sizeFreeStorage int64
}

// collectionsStats returns statistics about tables and indexes for the given collections.
//
// If refresh is true, it calls ANALYZE on the tables of the given list of collections.
//
// If the list of collections is empty, then stats filled with zero values is returned.
func collectionsStats(ctx context.Context, p *pgxpool.Pool, dbName string, list []*metadata.Collection, refresh bool) (*stats, error) { //nolint:lll // for readability
	if len(list) == 0 {
		return new(stats), nil
	}

	if refresh {
		fields := make([]string, len(list))
		for i, c := range list {
			fields[i] = pgx.Identifier{dbName, c.TableName}.Sanitize()
		}

		q := fmt.Sprintf(`ANALYZE %s`, strings.Join(fields, ", "))
		if _, err := p.Exec(ctx, q); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var s stats
	var placeholder metadata.Placeholder
	placeholders := make([]string, len(list))
	args := []any{dbName}

	placeholder.Next()

	for i, c := range list {
		placeholders[i] = placeholder.Next()
		args = append(args, c.TableName)
	}

	// The table size is the size used by collection documents. It excludes visibility map,
	// initialization fork, free space map and TOAST. The `main` `pg_relation_size` is
	// used, however it is not updated immediately after operation such as DELETE
	// unless VACUUM is called, ANALYZE does not update pg_relation_size in this case.
	//
	// The free storage size is the size of free space map (fsm) of table relation.
	//
	// The smallest difference in size that `pg_relation_size` reports appears to be 8KB.
	// Because of that inserting or deleting a single small object may not change the size.
	//
	// See also https://www.postgresql.org/docs/current/functions-admin.html#FUNCTIONS-ADMIN-DBSIZE,
	// visibility map https://www.postgresql.org/docs/current/storage-vm.html,
	// initialization fork https://www.postgresql.org/docs/current/storage-init.html,
	// free space map https://www.postgresql.org/docs/current/storage-fsm.html and
	// TOAST https://www.postgresql.org/docs/current/storage-toast.html.
	q := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(c.reltuples), 0),
			COALESCE(SUM(pg_relation_size(c.oid,'main')), 0),
			COALESCE(SUM(pg_relation_size(c.oid,'fsm')), 0),
			COALESCE(SUM(pg_indexes_size(c.oid)), 0)
		FROM pg_tables AS t
			LEFT JOIN pg_class AS c ON c.relname = t.tablename AND c.relnamespace = quote_ident(t.schemaname)::regnamespace
		WHERE t.schemaname = $1 AND t.tablename IN (%s)`,
		strings.Join(placeholders, ", "),
	)

	row := p.QueryRow(ctx, q, args...)
	if err := row.Scan(&s.countDocuments, &s.sizeTables, &s.sizeFreeStorage, &s.sizeIndexes); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &s, nil
}
