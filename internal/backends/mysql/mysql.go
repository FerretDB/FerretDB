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

// Package mysql provides backend for MySQL and compatible databases.
//
// # Design principles
//
//  1. Metadata is heavily cached to avoid most queries and transactions.
package mysql

import (
	"context"
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/backends/mysql/metadata"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

const (
	// ErrDuplicateEntry is the unique key violation error code for MySQL.
	ErrDuplicateEntry = 1062
)

// stats represents information about statistics of tables and indexes.
type stats struct {
	countDocuments  int64
	sizeIndexes     int64
	sizeTables      int64
	sizeFreeStorage int64
	totalSize       int64
}

// collectionStats returns statistics about tables and indexes for the given collections.
//
// If refresh is true, it calls ANALYZE on the tables of the given list of collections.
//
// If the list of collections is empty, then stats filled with zero values is returned.
func collectionsStats(ctx context.Context, p *fsql.DB, dbName string, list []*metadata.Collection, refresh bool) (*stats, error) { //nolint:lll // for readability
	if len(list) == 0 {
		return new(stats), nil
	}

	if refresh {
		fields := make([]string, len(list))
		for i, c := range list {
			fields[i] = fmt.Sprintf("%s.%s", dbName, c.TableName)
		}

		q := fmt.Sprintf(`ANALYZE TABLE %s`, strings.Join(fields, ", "))
		if _, err := p.ExecContext(ctx, q); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	var s stats
	placeholders := make([]string, len(list))
	args := []any{dbName}

	for i, c := range list {
		placeholders[i] = "?"
		args = append(args, c.TableName)
	}

	// The table size is the size used by collection documents. The `data_length` in addition
	// to the `index_length` is used since MySQL uses clustered indexes, however, these are
	// not updated immediately after operations such as DELETE unless OPTIMIZE TABLE is called.
	//
	// The free storage size of each relation is reported in `data_free`.
	//
	// The smallest difference in size that `data_length` reports appears to be 16KB.
	// Because of that inserting or deleting a single small object may not change the size.
	//
	// See also:
	// Clustered Index: https://dev.mysql.com/doc/refman/8.0/en/innodb-index-types.html
	q := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(table_rows), 0),
			COALESCE(SUM(data_length), 0),
			COALESCE(SUM(data_free)),
			COALESCE(SUM(index_length), 0),
			COALESCE(SUM(data_length) + SUM(index_length), 0)
		FROM information_schema.tables
		WHERE table_schema = ? AND table_name IN (%s)`,
		strings.Join(placeholders, ", "),
	)

	row := p.QueryRowContext(ctx, q, args...)
	if err := row.Scan(&s.countDocuments, &s.sizeTables, &s.sizeFreeStorage, &s.sizeIndexes, &s.totalSize); err != nil {
		return nil, lazyerrors.Error(err)
	}

	return &s, nil
}
