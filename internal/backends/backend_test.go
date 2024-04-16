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

package backends_test // to avoid import cycle

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	for name, b := range testBackends(t) {
		name, b := name, b
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := b.sp.Get()
			assert.Empty(t, s.BackendName)
			assert.Empty(t, s.BackendVersion)

			db, err := b.Database(testutil.DatabaseName(t))
			require.NoError(t, err)

			err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
				Name: testutil.CollectionName(t),
			})
			require.NoError(t, err)

			s = b.sp.Get()
			require.NotEmpty(t, s.BackendName)

			switch s.BackendName {
			case "PostgreSQL":
				assert.True(t, strings.HasPrefix(s.BackendVersion, "16.2 ("), "%s", s.BackendName)
			case "SQLite":
				assert.Equal(t, "3.45.2", s.BackendVersion)
			case "hana":
				assert.Equal(t, "4.00.000.00.1693911223", s.BackendVersion)
			default:
				t.Fatalf("unknown backend: %s", name)
			}
		})
	}
}

func TestListDatabases(t *testing.T) {
	t.Parallel()

	ctx := conninfo.Ctx(testutil.Ctx(t), conninfo.New())

	for name, b := range testBackends(t) {
		name, b := name, b
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// setup 3 DB with 1 collections each. random order of db name also ensure sorting test.
			dbNames := []string{"testDB2", "testDB1", "testDB3"}
			collectionName := "testCollection"

			testDB, err := b.Database(dbNames[0])
			require.NoError(t, err)
			err = testDB.CreateCollection(ctx, &backends.CreateCollectionParams{Name: collectionName})
			require.NoError(t, err)

			testDB, err = b.Database(dbNames[1])
			require.NoError(t, err)
			err = testDB.CreateCollection(ctx, &backends.CreateCollectionParams{Name: collectionName})
			require.NoError(t, err)

			testDB, err = b.Database(dbNames[2])
			require.NoError(t, err)
			err = testDB.CreateCollection(ctx, &backends.CreateCollectionParams{Name: collectionName})
			require.NoError(t, err)

			t.Run("ListDatabasesWithGivenName", func(t *testing.T) {
				t.Parallel()
				dbRes, err := b.ListDatabases(ctx, &backends.ListDatabasesParams{Name: dbNames[2]})
				require.NoError(t, err)
				require.Equal(t, 1, len(dbRes.Databases), "expected len 1 , since only 1 db with name testDB3")
				require.Equal(t, dbNames[2], dbRes.Databases[0].Name, "expected name testDB3")
			})

			t.Run("ListDatabaseWithDummyName", func(t *testing.T) {
				t.Parallel()
				dbRes, err := b.ListDatabases(ctx, &backends.ListDatabasesParams{Name: "dummy"})
				require.NoError(t, err)
				require.Equal(t, 0, len(dbRes.Databases), "expected len 0 since no db with name dummy")
			})

			t.Run("ListDatabasesWithNilParam", func(t *testing.T) {
				t.Parallel()
				dbRes, err := b.ListDatabases(ctx, nil)
				require.NoError(t, err)
				require.Equal(t, 3, len(dbRes.Databases), "expected full list len 3")
				require.Equal(t, dbNames[1], dbRes.Databases[0].Name, "expected name testDB1")
				require.Equal(t, dbNames[0], dbRes.Databases[1].Name, "expected name testDB2")
				require.Equal(t, dbNames[2], dbRes.Databases[2].Name, "expected name testDB3")
			})

			t.Run("ListDatabasesWithEMptyParam", func(t *testing.T) {
				t.Parallel()
				var param backends.ListDatabasesParams
				dbRes, err := b.ListDatabases(ctx, &param)
				require.NoError(t, err)
				require.Equal(t, 3, len(dbRes.Databases), "expected full list len 3")
				require.Equal(t, dbNames[1], dbRes.Databases[0].Name, "expected name testDB1")
				require.Equal(t, dbNames[0], dbRes.Databases[1].Name, "expected name testDB2")
				require.Equal(t, dbNames[2], dbRes.Databases[2].Name, "expected name testDB3")
			})
		})
	}
}
