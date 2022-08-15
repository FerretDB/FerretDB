// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tigrisdb

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCreateCollectionIfNotExist(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	cfg := &config.Driver{
		URL: testutil.TigrisURL(t),
	}
	tdb, err := New(cfg)
	require.NoError(t, err)

	t.Run("DBCollectionDoNotExist", func(t *testing.T) {
		t.Parallel()

		dbName := testutil.DatabaseName(t)
		collName := testutil.CollectionName(t)

		t.Cleanup(func() {
			require.NoError(t, tdb.Driver.DropDatabase(ctx, dbName))
		})

		created, err := tdb.CreateDatabaseIfNotExists(ctx, dbName)
		require.NoError(t, err)
		assert.True(t, created)
		err = tdb.InTransaction(ctx, dbName, func(tx driver.Tx) error {
			schema := driver.Schema(strings.TrimSpace(fmt.Sprintf(
				`{"title": "%s","properties": {"_id": {"type": "string","format": "byte"}},"primary_key": ["_id"]}`,
				collName)))
			created, err := CreateCollectionIfNotExist(ctx, tx, collName, schema)
			require.NoError(t, err)
			assert.True(t, created)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("DBCollectionExist", func(t *testing.T) {
		t.Parallel()

		dbName := testutil.DatabaseName(t)
		collName := testutil.CollectionName(t)

		require.NoError(t, tdb.Driver.CreateDatabase(ctx, dbName))
		schema := driver.Schema(strings.TrimSpace(fmt.Sprintf(
			`{"title": "%s","properties": {"_id": {"type": "string","format": "byte"}},"primary_key": ["_id"]}`,
			collName)))
		require.NoError(t, tdb.Driver.UseDatabase(dbName).CreateOrUpdateCollection(ctx, collName, schema))

		t.Cleanup(func() {
			require.NoError(t, tdb.Driver.DropDatabase(ctx, dbName))
		})

		created, err := tdb.CreateDatabaseIfNotExists(ctx, dbName)
		require.NoError(t, err)
		assert.False(t, created)

		err = tdb.InTransaction(ctx, dbName, func(tx driver.Tx) error {
			created, err := CreateCollectionIfNotExist(ctx, tx, collName, schema)
			require.NoError(t, err)
			assert.False(t, created)
			return nil
		})
		require.NoError(t, err)
	})
}
