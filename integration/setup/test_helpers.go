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
)

// IsTigris returns true if tests are running against FerretDB with `tigris` handler.
//
// This function should not be used lightly.
func IsTigris(tb testing.TB) bool {
	tb.Helper()

	return *targetBackendF == "ferretdb-tigris"
}

// SkipForTigris is deprecated.
//
// Deprecated: use SkipForTigrisWithReason instead if you must.
func SkipForTigris(tb testing.TB) {
	tb.Helper()

	SkipForTigrisWithReason(tb, "empty, please update this test")
}

// SkipForTigrisWithReason skips the current test for FerretDB with `tigris` handler.
//
// This function should not be used lightly.
func SkipForTigrisWithReason(tb testing.TB, reason string) {
	tb.Helper()

	require.NotEmpty(tb, reason, "reason must not be empty")

	if IsTigris(tb) {
		tb.Skipf("Skipping for Tigris: %s.", reason)
	}
}

// TigrisOnlyWithReason skips the current test except for FerretDB with `tigris` handler.
//
// This function should not be used lightly.
func TigrisOnlyWithReason(tb testing.TB, reason string) {
	tb.Helper()

	require.NotEmpty(tb, reason, "reason must not be empty")

	if !IsTigris(tb) {
		tb.Skipf("Skipping for non-tigris: %s", reason)
	}
}

// IsPushdownDisabled returns if FerretDB pushdowns are disabled.
func IsPushdownDisabled() bool {
	return *disablePushdownF
}
