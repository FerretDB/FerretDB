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
	countRows    int64
	countIndexes int64
	sizeIndexes  int64
	sizeTables   int64
}

// collectionsStats returns statistics about tables and indexes for the given collections.
func collectionsStats(ctx context.Context, p *pgxpool.Pool, dbName string, list []*metadata.Collection, refresh bool) (*stats, error) { //nolint:lll // for readability
	var err error

	if refresh {
		// Calling VACUUM marks dead rows for deletion, then calls ANALYZE.
		// However, the purpose is to update statistics of tables from recent
		// delete operations instead of reclaiming space.
		// It does not lock read and write operations.
		// See https://www.postgresql.org/docs/current/sql-vacuum.html.
		var q string
		for _, c := range list {
			q += fmt.Sprintf(`VACUUM ANALYZE %s;`, pgx.Identifier{dbName, c.TableName}.Sanitize())
		}

		if _, err = p.Exec(ctx, q); err != nil {
			return nil, lazyerrors.Error(err)
		}
	} else {
		// Call ANALYZE to update statistics of tables and indexes,
		// see https://wiki.postgresql.org/wiki/Count_estimate.
		q := `ANALYZE`
		if _, err = p.Exec(ctx, q); err != nil {
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

	// get index count from metadata
	// TODO https://github.com/FerretDB/FerretDB/issues/3394
	s.countIndexes = 0

	// The sizeTables is the size used by collection objects and excludes visibility map,
	// initialization fork, free space map and TOAST. The main pg_relation_size is used,
	// however it may not be immediately updated after operation such as DELETE.
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
		    COALESCE(SUM(pg_indexes_size(c.oid)), 0)
		FROM pg_tables AS t
		    LEFT JOIN pg_class AS c ON c.relname = t.tablename AND c.relnamespace = quote_ident(t.schemaname)::regnamespace
		WHERE t.schemaname = $1 AND t.tablename IN (%s)`,
		strings.Join(placeholders, ", "),
	)

	row := p.QueryRow(ctx, q, args...)
	if err := row.Scan(&s.countRows, &s.sizeTables, &s.sizeIndexes); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &s, nil
}
