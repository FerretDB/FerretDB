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

// Package sqlite provides SQLite backend.
//
// # Design principles
//
//  1. Transactions should be avoided when possible.
//     That's because there can be, at most, one write [transaction] at a given time for the whole database.
//     (Note that there is a separate branch of SQLite with [concurrent transactions], but it is not merged yet.)
//     FerretDB often could use more granular locks - for example, per collection.
//  2. Explicit transaction retries and [SQLITE_BUSY] handling should be avoided - see above.
//     Additionally, SQLite retries automatically with the [busy_timeout] parameter we set by default, which should be enough.
//  3. Metadata is heavily cached to avoid most queries and transactions.
//
// [transaction]: https://www.sqlite.org/lang_transaction.html
// [concurrent transactions]: https://www.sqlite.org/cgi/src/doc/begin-concurrent/doc/begin_concurrent.md
// [SQLITE_BUSY]: https://www.sqlite.org/rescode.html#busy
// [busy_timeout]: https://www.sqlite.org/pragma.html#pragma_busy_timeout
package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/backends/sqlite/metadata"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// stats represents information about statistics of tables and indexes.
type stats struct {
	countDocuments  int64
	sizeIndexes     int64
	sizeTables      int64
	sizeFreeStorage int64 // free storage for the entire database
}

// collectionsStats returns statistics about tables and indexes for the given collections.
//
// If refresh is true, it calls ANALYZE on the tables of the given list of collections.
//
// If the list of collections is empty, then stats filled with zero values is returned.
func collectionsStats(ctx context.Context, db *fsql.DB, list []*metadata.Collection, refresh bool) (*stats, error) {
	if len(list) == 0 {
		return new(stats), nil
	}

	var err error

	if refresh {
		queries := make([]string, len(list))
		for i, c := range list {
			queries[i] = fmt.Sprintf("ANALYZE %q;", c.TableName)
		}

		// SQLite ANALYZE does not allow multiple tables as arguments, hence
		// build multiple queries such as `ANALYZE 'table1'; ANALYZE 'table2';`.
		q := strings.Join(queries, "")
		if _, err = db.ExecContext(ctx, q); err != nil {
			return nil, lazyerrors.Error(err)
		}
	}

	placeholders := make([]string, len(list))
	args := make([]any, len(list))

	var indexes int64

	for i, c := range list {
		placeholders[i] = "?"
		args[i] = c.TableName

		indexes += int64(len(c.Settings.Indexes))
	}

	// The table size is the size used by collection documents. The `pgsize` of `dbstat`
	// table does not include freelist pages, pointer-map pages, and the lock page.
	//
	// If rows are deleted from a page but there are other rows on that same page,
	// the page won't be moved to freelist pages.
	//
	// The smallest difference in size that `pgsize` reports appears to be 4KB.
	// Because of that inserting or deleting a single small object may not change the size.
	//
	// See https://www.sqlite.org/dbstat.html and https://www.sqlite.org/fileformat.html.
	q := fmt.Sprintf(`
		SELECT SUM(pgsize)
		FROM dbstat
		WHERE name IN (%s) AND aggregate = TRUE`,
		strings.Join(placeholders, ", "),
	)

	stats := new(stats)
	if err = db.QueryRowContext(ctx, q, args...).Scan(&stats.sizeTables); err != nil {
		return nil, lazyerrors.Error(err)
	}

	// Use number of cells to approximate total document count,
	// excluding `internal` and `overflow` pagetype used by SQLite.
	// See https://www.sqlite.org/dbstat.html and https://www.sqlite.org/fileformat.html.
	q = fmt.Sprintf(`
		SELECT SUM(ncell)
		FROM dbstat
		WHERE name IN (%s) AND pagetype = 'leaf'`,
		strings.Join(placeholders, ", "),
	)

	if err = db.QueryRowContext(ctx, q, args...).Scan(&stats.countDocuments); err != nil {
		return nil, lazyerrors.Error(err)
	}

	placeholders = make([]string, 0, indexes)
	args = make([]any, 0, indexes)

	for _, c := range list {
		for _, index := range c.Settings.Indexes {
			placeholders = append(placeholders, "?")
			args = append(args, c.TableName+"_"+index.Name)
		}
	}

	q = fmt.Sprintf(`
		SELECT SUM(pgsize)
		FROM dbstat
		WHERE name IN (%s) AND aggregate = TRUE`,
		strings.Join(placeholders, ", "),
	)

	if err = db.QueryRowContext(ctx, q, args...).Scan(&stats.sizeIndexes); err != nil {
		return nil, lazyerrors.Error(err)
	}

	// https://www.sqlite.org/pragma.html#pragma_freelist_count
	q = `PRAGMA freelist_count`

	var freeListCount int64
	if err = db.QueryRowContext(ctx, q).Scan(&freeListCount); err != nil {
		return nil, lazyerrors.Error(err)
	}

	// https://www.sqlite.org/pragma.html#pragma_page_size
	q = `PRAGMA page_size`

	var pageSize int64
	if err = db.QueryRowContext(ctx, q).Scan(&pageSize); err != nil {
		return nil, lazyerrors.Error(err)
	}

	// SQLite is unable to provide free storage size per collection,
	// hence compute it for the entire database.
	// It reports free storage regardless of file or in memory database.
	//
	// https://www.sqlite.org/fileformat.html
	// > All pages within the same database are the same size.
	stats.sizeFreeStorage = freeListCount * pageSize

	return stats, nil
}
