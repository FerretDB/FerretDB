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
	"fmt"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCreateIndexIfNotExists(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	pool := getPool(ctx, t)

	databaseName := testutil.DatabaseName(t)
	collectionName := testutil.CollectionName(t)
	setupDatabase(ctx, t, pool, databaseName)

	indexName := "test"
	err := pool.InTransaction(ctx, func(tx pgx.Tx) error {
		idx := Index{
			Name: indexName,
			Key:  []IndexKeyPair{{Field: "foo", Order: types.Ascending}, {Field: "bar", Order: types.Descending}},
		}
		return CreateIndexIfNotExists(ctx, tx, databaseName, collectionName, &idx)
	})
	require.NoError(t, err)

	tableName := collectionNameToTableName(collectionName)
	pgIndexName := indexNameToPgIndexName(collectionName, indexName)

	var indexdef string
	err = pool.QueryRow(
		ctx,
		"SELECT indexdef FROM pg_indexes WHERE schemaname = $1 AND tablename = $2 AND indexname = $3",
		databaseName, tableName, pgIndexName,
	).Scan(&indexdef)
	require.NoError(t, err)

	expectedIndexdef := fmt.Sprintf(
		"CREATE UNIQUE INDEX %s ON %s.%s USING btree (((_jsonb -> 'foo'::text)), ((_jsonb -> 'bar'::text)) DESC)",
		pgIndexName, databaseName, tableName,
	)
	assert.Equal(t, expectedIndexdef, indexdef)
}
