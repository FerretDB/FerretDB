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

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// recreateDir removes and creates a directory and returns its absolute path.
func recreateDir(dir string) string {
	dir, err := filepath.Abs(dir)
	must.NoError(err)

	must.NoError(os.RemoveAll(dir))
	must.NoError(os.MkdirAll(dir, 0o777))

	return dir
}

// parseSQLiteURL parses and checks SQLite URI.
func parseSQLiteURL(sqliteURL string) *url.URL {
	u, err := url.Parse(sqliteURL)
	must.NoError(err)

	must.BeTrue(u.Path == "")
	must.BeTrue(u.Opaque != "")

	return u
}

// sharedSQLiteURL returns SQLite URI for all tests.
func sharedSQLiteURL(sqliteURL string) string {
	u := parseSQLiteURL(sqliteURL)
	dir := recreateDir(u.Opaque)

	zap.S().Infof("Using shared SQLite URI: %s (%s).", u.String(), dir)

	return u.String()
}

// privateSQLiteURL returns test-specific SQLite URI.
// It is cleaned-up if test passes.
func privateSQLiteURL(tb testtb.TB, sqliteURL string) string {
	must.NotBeZero(tb)
	tb.Helper()

	u := parseSQLiteURL(sqliteURL)

	u.Opaque = path.Join(u.Opaque, testutil.DatabaseName(tb)) + "/"
	dir := recreateDir(u.Opaque)

	res := u.String()

	tb.Cleanup(func() {
		if tb.Failed() {
			tb.Logf("Keeping %s (%s) for debugging.", res, dir)

			return
		}

		require.NoError(tb, os.RemoveAll(dir))
	})

	return res
}

// sharedPostgreSQLURL returns PostgreSQL URL for all tests.
func sharedPostgreSQLURL(postgreSQLURL string) string {
	zap.S().Infof("Using shared PostgreSQL URL: %s.", postgreSQLURL)

	return postgreSQLURL
}

// privatePostgreSQLURL returns test-specific PostgreSQL URL.
// Currently, it is the same as shared URL.
func privatePostgreSQLURL(tb testtb.TB, postgreSQLURL string) string {
	must.NotBeZero(tb)
	tb.Helper()

	return postgreSQLURL
}
