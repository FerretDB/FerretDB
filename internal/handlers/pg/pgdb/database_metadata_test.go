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
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestDatabaseMetadata(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)

	pool := getPool(ctx, t)
	databaseName := testutil.DatabaseName(t)
	collectionName := testutil.CollectionName(t)
	setupDatabase(ctx, t, pool, databaseName)

	err := pool.InTransactionRetry(ctx, func(tx pgx.Tx) error {
		m := newMetadataStorage(tx, databaseName, collectionName)
		nameCreated, _, err := m.store(ctx)
		// In this case error is possible: if this test is run in parallel with other tests,
		// ensureMetadata may fail to create the index or insert data due to concurrent requests to PostgreSQL.
		// In such case, we expect InTransactionRetry to handle the error and retry the transaction if neede.
		if err != nil {
			return err
		}

		var nameFound string

		nameFound, err = m.getTableName(ctx)
		require.NoError(t, err)

		assert.Equal(t, nameCreated, nameFound)

		// adding metadata that already exist should not fail
		_, _, err = m.store(ctx)
		require.NoError(t, err)

		err = m.remove(ctx)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)
}
