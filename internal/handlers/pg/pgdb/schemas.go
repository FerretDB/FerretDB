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

	"github.com/jackc/pgx/v4"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// schemaExists returns true if given schema exists.
func schemaExists(ctx context.Context, tx pgx.Tx, db string) (bool, error) {
	sql := `SELECT nspname FROM pg_catalog.pg_namespace WHERE nspname = $1`
	rows, err := tx.Query(ctx, sql, db)
	if err != nil {
		return false, lazyerrors.Error(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		must.NoError(rows.Scan(&name))
		if name == db {
			return true, nil
		}
	}

	return false, nil
}
