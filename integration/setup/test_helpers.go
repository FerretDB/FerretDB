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
	"github.com/stretchr/testify/require"

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

// FailsForFerretDB return testtb.TB that expects test to fail for FerretDB and pass for MongoDB.
//
// This function should not be used lightly and always with an issue URL.
func FailsForFerretDB(tb testtb.TB, reason string) testtb.TB {
	tb.Helper()

	if IsMongoDB(tb) {
		return tb
	}

	return testfail.Expected(tb, reason)
}

// FailsForSQLite return testtb.TB that expects test to fail for FerretDB with SQLite backend and pass otherwise.
//
// This function should not be used lightly and always with an issue URL.
func FailsForSQLite(tb testtb.TB, reason string) testtb.TB {
	tb.Helper()

	if IsSQLite(tb) {
		return testfail.Expected(tb, reason)
	}

	return tb
}

// SkipForPostgreSQL skips the current test for PostgreSQL backend.
//
// This function should not be used lightly and always with an reason provided.
// It should be removed when a new PG handler is fully supported.
// TODO https://github.com/FerretDB/FerretDB/issues/3435
func SkipForPostgreSQL(tb testtb.TB, reason string) {
	tb.Helper()

	if IsPostgreSQL(tb) {
		require.NotEmpty(tb, reason, "reason must not be empty")

		tb.Skipf("Skipping for PostgreSQL: %s.", reason)
	}
}

// SkipForMongoDB skips the current test for MongoDB.
//
// This function should not be used lightly and always with an issue URL.
func SkipForMongoDB(tb testtb.TB, reason string) {
	tb.Helper()

	if IsMongoDB(tb) {
		require.NotEmpty(tb, reason, "reason must not be empty")

		tb.Skipf("Skipping for MongoDB: %s.", reason)
	}
}

// FilterPushdownDisabled returns true if FerretDB filter pushdown is disabled.
func FilterPushdownDisabled() bool {
	return *disableFilterPushdownF
}

// SortPushdownEnabled returns true if sort pushdown is enabled.
func SortPushdownEnabled() bool {
	return *enableSortPushdownF
}

// UnsafePushdwon returns true if unsafe pushdown is enabled.
func UnsafePushdown() bool {
	return *enableUnsafePushdown
}
