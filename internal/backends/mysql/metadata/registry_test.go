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

package metadata

import (
	"context"
	"testing"

	"github.com/FerretDB/FerretDB/internal/clientconn/conninfo"
	"github.com/FerretDB/FerretDB/internal/util/fsql"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/stretchr/testify/require"
)

// createDatabse creates a new provider and registry required for creating a database and
// returns registry, db pool and created database name.
func createDatabase(t *testing.T, ctx context.Context) (r *Registry, db *fsql.DB, dbName string) {
	t.Helper()

	sp, err := state.NewProvider("")
	require.NoError(t, err)

	u := "mysql://username:password@127.0.0.1:3306/ferretdb"
	r, err = NewRegistry(u, testutil.Logger(t), sp)
	require.NoError(t, err)
	t.Cleanup(r.Close)

	dbName = testutil.DatabaseName(t)
	db, err = r.DatabaseGetOrCreate(ctx, dbName)
	require.NoError(t, err)
	require.NotNil(t, db)

	t.Cleanup(func() {
		_, err = r.DatabaseDrop(ctx, dbName)
		require.NoError(t, err)
	})

	return r, db, dbName
}

func TestCheckAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	connInfo := conninfo.New()
	ctx := conninfo.Ctx(testutil.Ctx(t), connInfo)

	for name, tc := range map[string]struct {
		uri string
		err string
	}{
		"Auth": {
			uri: "mysql://username:password@127.0.0.1:3306/ferretdb",
			err: "",
		},
		"NoAuth": {
			uri: "mysql://127.0.0.1:3306/ferretdb",
			err: "failed to connect to `host=127.0.0.1 user=`",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {

			sp, err := state.NewProvider("")
			require.NoError(t, err)

			r, err := NewRegistry(tc.uri, testutil.Logger(t), sp)
			require.NoError(t, err)
			t.Cleanup(r.Close)

			_, err = r.getPool(ctx)
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}

			require.NoError(t, err)
		})
	}
}
