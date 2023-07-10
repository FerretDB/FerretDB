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

	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// FailsForFerretDB return testutil.TB that expects test to fail for FerretDB and pass for MongoDB.
//
// This function should not be used lightly and always with an issue URL.
func FailsForFerretDB(tb testutil.TB, reason string) testutil.TB {
	tb.Helper()

	if *targetBackendF == "mongodb" {
		return tb
	}

	return testutil.XFail(tb, reason)
}

// SkipForMongoDB skips the current test for MongoDB.
//
// This function should not be used lightly.
func SkipForMongoDB(tb testutil.TB, reason string) {
	tb.Helper()

	if *targetBackendF == "mongodb" {
		require.NotEmpty(tb, reason, "reason must not be empty")

		tb.Skipf("Skipping for MongoDB: %s.", reason)
	}
}

// IsPushdownDisabled returns if FerretDB pushdowns are disabled.
func IsPushdownDisabled() bool {
	return *disableFilterPushdownF
}

// IsSortPushdownEnabled returns true if sort pushdown is enabled.
func IsSortPushdownEnabled() bool {
	return *enableSortPushdownF
}
