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
//
// If the list of collections is empty, then stats filled with zero values is returned.
func collectionsStats(ctx context.Context, p *pgxpool.Pool, dbName string, list []*metadata.Collection) (*stats, error) {
	if len(list) == 0 {
		return new(stats), nil
	}

	var err error

	// TODO https://github.com/FerretDB/FerretDB/issues/3518
	q := `ANALYZE`
	if _, err = p.Exec(ctx, q); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var s stats
	var placeholder metadata.Placeholder
	placeholders := make([]string, len(list))
	args := []any{dbName}

	placeholder.Next()

	for i, c := range list {
		s.countIndexes += int64(len(c.Indexes))
		placeholders[i] = placeholder.Next()
		args = append(args, c.TableName)
	}

	q = fmt.Sprintf(`
		SELECT
			COALESCE(SUM(c.reltuples), 0),
			COALESCE(SUM(pg_table_size(c.oid)), 0),
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
