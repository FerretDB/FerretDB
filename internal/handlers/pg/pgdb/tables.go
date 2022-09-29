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

package pgdb

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Tables returns a sorted list of PostgreSQL table names.
// Returns empty slice if schema does not exist.
// Tables with prefix "_ferretdb_" are filtered out.
func Tables(ctx context.Context, tx pgx.Tx, schema string) ([]string, error) {
	tables, err := tables(ctx, tx, schema)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	filtered := make([]string, 0, len(tables))
	for _, table := range tables {
		if strings.HasPrefix(table, reservedPrefix) {
			continue
		}

		filtered = append(filtered, table)
	}

	return filtered, nil
}

// tables returns a list of PostgreSQL table names.
func tables(ctx context.Context, tx pgx.Tx, schema string) ([]string, error) {
	sql := `SELECT table_name ` +
		`FROM information_schema.columns ` +
		`WHERE table_schema = $1 ` +
		`GROUP BY table_name ` +
		`ORDER BY table_name`
	rows, err := tx.Query(ctx, sql, schema)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	defer rows.Close()

	tables := make([]string, 0, 2)
	var name string
	for rows.Next() {
		if err = rows.Scan(&name); err != nil {
			return nil, lazyerrors.Error(err)
		}

		tables = append(tables, name)
	}
	if err = rows.Err(); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return tables, nil
}
