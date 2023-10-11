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

package setup

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/stretchr/testify/require"
)

// recreateDir removes and creates a directory and returns its absolute path.
func recreateDir(dir string) string {
	dir, err := filepath.Abs(dir)
	must.NoError(err)

	must.NoError(os.RemoveAll(dir))
	must.NoError(os.MkdirAll(dir, 0o777))

	return dir
}

func parseSQLiteURL(sqliteURL string) *url.URL {
	u, err := url.Parse(sqliteURL)
	must.NoError(err)

	must.BeTrue(u.Path == "")
	must.BeTrue(u.Opaque != "")

	return u
}

func sharedSQLiteURL(sqliteURL string) string {
	u := parseSQLiteURL(sqliteURL)
	recreateDir(u.Opaque)
	return u.String()
}

// use per-test directory to prevent backend's metadata registry
// read databases owned by concurrent tests
func privateSQLiteURL(tb testing.TB, sqliteURL string) string {
	tb.Helper()

	u := parseSQLiteURL(sqliteURL)

	u.Opaque = path.Join(u.Opaque, testutil.DatabaseName(tb)) + "/"
	dir := recreateDir(u.Opaque)

	res := u.String()

	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf("Keeping %s (%s) for debugging.", dir, res)

			return
		}

		require.NoError(tb, os.RemoveAll(dir))
	})

	return res
}

func sharedPostgreSQLURL(postgreSQLURL string) string {
	return postgreSQLURL
}

func privatePostgreSQLURL(tb testing.TB, postgreSQLURL string) string {
	tb.Helper()

	return postgreSQLURL
}
