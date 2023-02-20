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
)

// indexParams describes the parameters for creating an index.
type indexParams struct {
	schema   string // pg schema name
	table    string // pg table name
	isUnique bool   // whether the index is unique
}

// createIndexIfNotExists creates a new index for the given params if it does not exist.
func createIndexIfNotExists(ctx context.Context, tx pgx.Tx, p *indexParams) error {
	var err error

	unique := ""
	if p.isUnique {
		unique = " UNIQUE"
	}

	sql := `CREATE` + unique + ` INDEX IF NOT EXISTS ` + pgx.Identifier{p.table + "_id"}.Sanitize() +
		` ON ` + pgx.Identifier{p.schema, p.table}.Sanitize() +
		` ((_jsonb->>'_id'))`

	if _, err = tx.Exec(ctx, sql); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}
