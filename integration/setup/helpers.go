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
	"path/filepath"
	"runtime"
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testfail"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// IsPostgreSQL returns true if the current test is running for PostgreSQL backend.
//
// This function should not be used lightly.
func IsPostgreSQL(tb testtb.TB) bool {
	return *targetBackendF == "ferretdb-postgresql"
}

// IsSQLite returns true if the current test is running for SQLite backend.
//
// This function should not be used lightly.
func IsSQLite(tb testtb.TB) bool {
	return *targetBackendF == "ferretdb-sqlite"
}

// IsMongoDB returns true if the current test is running for MongoDB.
//
// This function should not be used lightly.
func IsMongoDB(tb testtb.TB) bool {
	return *targetBackendF == "mongodb"
}

// IsHana returns true if the current test is running for Hana backend.
//
// This function should not be used lightly.
func IsHana(tb testtb.TB) bool {
	return *targetBackendF == "ferretdb-hana"
}

// ensureIssueURL panics if URL is not a valid FerretDB issue URL.
func ensureIssueURL(url string) {
	must.BeTrue(strings.HasPrefix(url, "https://github.com/FerretDB/FerretDB/issues/"))
}

// FailsForFerretDB return testtb.TB that expects test to fail for FerretDB and pass for MongoDB.
//
// This function should not be used lightly and always with an issue URL.
func FailsForFerretDB(tb testtb.TB, url string) testtb.TB {
	tb.Helper()

	ensureIssueURL(url)

	if IsMongoDB(tb) {
		return tb
	}

	return testfail.Expected(tb, url)
}

// FailsForMongoDB return testtb.TB that expects test to fail for MongoDB and pass for FerretDB.
//
// This function should not be used lightly.
func FailsForMongoDB(tb testtb.TB, reason string) testtb.TB {
	tb.Helper()

	if IsMongoDB(tb) {
		return testfail.Expected(tb, reason)
	}

	return tb
}

// SkipForMongoDB skips the current test for MongoDB.
//
// Use [FailsForMongoDB] in new code.
func SkipForMongoDB(tb testtb.TB, reason string) {
	tb.Helper()

	if IsMongoDB(tb) {
		require.NotEmpty(tb, reason, "reason must not be empty")

		tb.Skipf("Skipping for MongoDB: %s.", reason)
	}
}

// PushdownDisabled returns true if FerretDB pushdown is disabled.
func PushdownDisabled() bool {
	return *disablePushdownF
}

// Dir returns the absolute directory of this package.
func Dir(tb testtb.TB) string {
	tb.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(tb, ok)

	return filepath.Dir(file)
}
