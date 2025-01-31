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
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/FerretDB/xfail"
	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// IsMongoDB returns true if the current test is running for MongoDB,
// and false if it's running for FerretDB/PostgreSQL/DocumentDB.
//
// This function should not be used lightly.
func IsMongoDB(tb testing.TB) bool {
	tb.Helper()

	return *targetBackendF == "mongodb"
}

// ensureIssueURL panics if URL is not a valid FerretDB issue URL.
func ensureIssueURL(url string) {
	ferretDB := strings.HasPrefix(url, "https://github.com/FerretDB/FerretDB/issues/")
	documentDB := strings.HasPrefix(url, "https://github.com/FerretDB/FerretDB-DocumentDB/issues/")
	must.BeTrue(ferretDB || documentDB)
}

// FailsForFerretDB return testing.TB that expects test to fail for FerretDB and pass for MongoDB.
// It returns original value if -no-xfail flag was passed.
//
// This function should not be used lightly and always with an issue URL.
func FailsForFerretDB(tb testing.TB, url string) testing.TB {
	tb.Helper()

	ensureIssueURL(url)

	if IsMongoDB(tb) {
		return tb
	}

	if *noXFailF {
		tb.Logf("Test should fail: %s", url)
		return tb
	}

	return xfail.XFail(tb, url)
}

// FailsForMongoDB return testing.TB that expects test to fail for MongoDB and pass for FerretDB.
// It returns original value if -no-xfail flag was passed.
//
// This function should not be used lightly.
func FailsForMongoDB(tb testing.TB, reason string) testing.TB {
	tb.Helper()

	if !IsMongoDB(tb) {
		return tb
	}

	if *noXFailF {
		tb.Logf("Test should fail: %s", reason)
		return tb
	}

	return xfail.XFail(tb, reason)
}

// SkipForMongoDB skips the current test for MongoDB.
//
// [FailsForMongoDB] should be used instead when possible.
func SkipForMongoDB(tb testing.TB, reason string) {
	tb.Helper()

	if IsMongoDB(tb) {
		require.NotEmpty(tb, reason, "reason must not be empty")

		tb.Skipf("Skipping for MongoDB: %s.", reason)
	}
}

// Dir returns the absolute directory of this package.
func Dir(tb testing.TB) string {
	tb.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(tb, ok)
	require.True(tb, filepath.IsAbs(file))

	return filepath.Dir(file)
}

// Main is the entry point for all integration test packages.
// It should be called from main_test.go in each package.
func Main(m *testing.M) {
	flag.Parse()

	var code int

	// ensure that Shutdown runs for any exit code or panic
	func() {
		// make `go test -list=.` work without side effects
		if flag.Lookup("test.list").Value.String() == "" {
			Startup()

			defer Shutdown()
		}

		code = m.Run()
	}()

	os.Exit(code)
}
