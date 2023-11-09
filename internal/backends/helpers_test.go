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
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/backends"
	"github.com/FerretDB/FerretDB/internal/backends/hana"
	"github.com/FerretDB/FerretDB/internal/backends/postgresql"
	"github.com/FerretDB/FerretDB/internal/backends/sqlite"
	"github.com/FerretDB/FerretDB/internal/util/state"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// testBackend contains information about backend under test.
type testBackend struct {
	backends.Backend
	sp *state.Provider
}

// testBackends returns all backends configured for testing contracts.
func testBackends(t *testing.T) map[string]*testBackend {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping in -short mode")
	}

	l := testutil.Logger(t)

	res := map[string]*testBackend{}

	{
		sp, err := state.NewProvider("")
		require.NoError(t, err)

		b, err := postgresql.NewBackend(&postgresql.NewBackendParams{
			URI: testutil.TestPostgreSQLURI(t, context.TODO(), ""),
			L:   l.Named("postgresql"),
			P:   sp,
		})
		require.NoError(t, err)
		t.Cleanup(b.Close)

		res["postgresql"] = &testBackend{
			Backend: b,
			sp:      sp,
		}
	}

	{
		sp, err := state.NewProvider("")
		require.NoError(t, err)

		b, err := sqlite.NewBackend(&sqlite.NewBackendParams{
			URI: testutil.TestSQLiteURI(t, ""),
			L:   l.Named("sqlite"),
			P:   sp,
		})
		require.NoError(t, err)
		t.Cleanup(b.Close)

		res["sqlite"] = &testBackend{
			Backend: b,
			sp:      sp,
		}
	}

	if hanaURL := os.Getenv("FERRETDB_HANA_URL"); hanaURL != "" {
		sp, err := state.NewProvider("")
		require.NoError(t, err)

		b, err := hana.NewBackend(&hana.NewBackendParams{
			URI: hanaURL,
			L:   l.Named("hana"),
			P:   sp,
		})
		require.NoError(t, err)
		t.Cleanup(b.Close)

		res["hana"] = &testBackend{
			Backend: b,
			sp:      sp,
		}
	}

	return res
}

// assertErrorCode asserts that err is *Error with one of the given error codes.
func assertErrorCode(t *testing.T, err error, code backends.ErrorCode, codes ...backends.ErrorCode) {
	assert.True(t, backends.ErrorCodeIs(err, code, codes...), "err = %v", err)
}
