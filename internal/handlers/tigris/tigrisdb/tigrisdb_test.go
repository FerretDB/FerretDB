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

package tigrisdb

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigris-client-go/config"
	"github.com/tigrisdata/tigris-client-go/driver"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/handlers/tigris/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCreateCollectionIfNotExist(t *testing.T) {
	t.Parallel()

	ctx := testutil.Ctx(t)
	cfg := &config.Driver{
		URL: testutil.TigrisURL(t),
	}

	logger := testutil.Logger(t, zap.NewAtomicLevelAt(zap.DebugLevel))
	tdb, err := New(ctx, cfg, logger)
	require.NoError(t, err)

	t.Run("DBCollectionDoNotExist", func(t *testing.T) {
		t.Parallel()

		dbName := testutil.DatabaseName(t)
		collName := testutil.CollectionName(t)

		_, _ = tdb.Driver.DeleteProject(ctx, dbName)

		t.Cleanup(func() {
			_, e := tdb.Driver.DeleteProject(ctx, dbName)
			require.NoError(t, e)
		})

		schema := must.NotFail(tjson.DocumentSchema(must.NotFail(types.NewDocument("_id", types.NewObjectID()))))
		schema.Title = collName
		created, err := tdb.CreateOrUpdateCollection(ctx, dbName, collName, schema)
		require.NoError(t, err)
		assert.True(t, created)

		require.NoError(t, err)
	})

	t.Run("OnlyDBExists", func(t *testing.T) {
		t.Parallel()

		dbName := testutil.DatabaseName(t)
		collName := testutil.CollectionName(t)

		_, _ = tdb.Driver.DeleteProject(ctx, dbName)

		_, err := tdb.Driver.CreateProject(ctx, dbName)
		require.NoError(t, err)

		t.Cleanup(func() {
			_, e := tdb.Driver.DeleteProject(ctx, dbName)
			require.NoError(t, e)
		})

		created, err := tdb.createDatabaseIfNotExists(ctx, dbName)
		require.NoError(t, err)
		assert.False(t, created)

		schema := must.NotFail(tjson.DocumentSchema(must.NotFail(types.NewDocument("_id", types.NewObjectID()))))
		schema.Title = collName
		created, err = tdb.CreateOrUpdateCollection(ctx, dbName, collName, schema)
		require.NoError(t, err)
		assert.True(t, created)
	})

	t.Run("DBCollectionExist", func(t *testing.T) {
		t.Parallel()

		dbName := testutil.DatabaseName(t)
		collName := testutil.CollectionName(t)

		_, _ = tdb.Driver.DeleteProject(ctx, dbName)

		_, err := tdb.Driver.CreateProject(ctx, dbName)
		require.NoError(t, err)

		schema := driver.Schema(strings.TrimSpace(fmt.Sprintf(
			`{"title": "%s","properties": {"_id": {"type": "string","format": "byte"}},"primary_key": ["_id"]}`,
			collName,
		)))
		require.NoError(t, tdb.Driver.UseDatabase(dbName).CreateOrUpdateCollection(ctx, collName, schema))

		t.Cleanup(func() {
			_, e := tdb.Driver.DeleteProject(ctx, dbName)
			require.NoError(t, e)
		})

		created, err := tdb.createDatabaseIfNotExists(ctx, dbName)
		require.NoError(t, err)
		assert.False(t, created)

		schema1 := must.NotFail(tjson.DocumentSchema(must.NotFail(types.NewDocument("_id", types.NewObjectID()))))
		schema1.Title = collName
		created, err = tdb.CreateOrUpdateCollection(ctx, dbName, collName, schema1)
		require.NoError(t, err)
		assert.False(t, created)
	})
}
