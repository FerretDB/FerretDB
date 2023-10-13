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

			dbName := testutil.DatabaseName(t)
			cleanupDatabase(t, ctx, b, dbName)

			db, err := b.Database(dbName)
			require.NoError(t, err)

			err = db.CreateCollection(ctx, &backends.CreateCollectionParams{
				Name: testutil.CollectionName(t),
			})
			require.NoError(t, err)

			s = b.sp.Get()
			require.NotEmpty(t, s.BackendName)

			switch s.BackendName {
			case "PostgreSQL":
				assert.True(t, strings.HasPrefix(s.BackendVersion, "16.0 ("), "%s", s.BackendName)
			case "SQLite":
				assert.Equal(t, "3.41.2", s.BackendVersion)
			default:
				t.Fatalf("unknown backend: %s", name)
			}
		})
	}
}
