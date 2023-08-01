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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil/testfail"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// FailsForFerretDB return testtb.TB that expects test to fail for FerretDB and pass for MongoDB.
//
// This function should not be used lightly and always with an issue URL.
func FailsForFerretDB(tb testtb.TB, reason string) testtb.TB {
	tb.Helper()

	if *targetBackendF == "mongodb" {
		return tb
	}

	return testfail.Expected(tb, reason)
}

// FailsForSQLite return testtb.TB that expects test to fail for FerretDB with SQLite backend and pass otherwise.
//
// This function should not be used lightly and always with an issue URL.
func FailsForSQLite(tb testtb.TB, reason string) testtb.TB {
	tb.Helper()

	if *targetBackendF == "ferretdb-sqlite" {
		return testfail.Expected(tb, reason)
	}

	return tb
}

// SkipForMongoDB skips the current test for MongoDB.
//
// This function should not be used lightly and always with an issue URL.
func SkipForMongoDB(tb testtb.TB, reason string) {
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

type failCatcher struct {
	parent        testtb.TB   // the parent test
	tcs           []testtb.TB // subtests which we expect to have some failures
	targetBackend string      // target backend that we expect failure for
	reason        string      // reason of expected failure
}

// NewFailCatcher returns failCatcher that should be used to handle expected failures
// for subtests of t TB.
func NewFailCatcher(t testing.TB, backend, reason string) *failCatcher {
	if backend == "" {
		return &failCatcher{}
	}

	require.Contains(t, allBackends, backend)

	return &failCatcher{
		parent:        t,
		targetBackend: backend,
		reason:        reason,
	}
}

// Wrap wraps provided tb test into another test, that handles Fail() without reporting failure.
// It adds the test to the FailCatcher to catch any failures, and returns the test to be used in a subtest.
//
// If the targetBackend is not equal to the one set in failCatcher it returns nonwrapped test, so the
// Fail() calls are treated in the standard fashion.
func (e *failCatcher) Wrap(tb testtb.TB) testtb.TB {
	if *targetBackendF != e.targetBackend {
		return tb
	}

	t := testfail.New(tb)

	e.tcs = append(e.tcs, t)

	return t
}

// Catch goes through all of the subtests and checks if at least one failed.
// If not, it fails the parent test.
//
// If it's called for different backend than provided in failCatcher, it won't
// do anything.
func (e *failCatcher) Catch() bool {
	if *targetBackendF != e.targetBackend {
		return false
	}

	for _, tc := range e.tcs {
		if tc.Failed() {
			e.parent.Logf("Test failed as expected: %s", e.reason)
			return true
		}
	}

	e.parent.Fatalf("Test passed unexpectedly: %s", e.reason)
	return true
}
