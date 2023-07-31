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

// Package testfail provides testing helpers for expected tests failures.
package testfail

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// Expected return a new TB instance that expects the test to fail.
//
// At the end of the test, if it was marked as failed, it will pass instead.
// If it passes, it will be failed, so that Expected call can be removed.
func Expected(t testtb.TB, reason string) testtb.TB {
	t.Helper()

	require.NotEmpty(t, reason, "reason must not be empty")

	x := &expected{
		tb: t,
	}

	x.tb.Cleanup(func() {
		if x.Failed() {
			x.tb.Logf("Test failed as expected: %s", reason)
			return
		}

		x.tb.Fatalf("Test passed unexpectedly: %s", reason)
	})

	return x
}

// Expected return a new TB instance that expects the test to fail.
//
// At the end of the test, if it was marked as failed, it will pass instead.
// If it passes, it will be failed, so that Expected call can be removed.
func ExpectedTestingT(t testtb.T, reason string) testtb.T {
	t.Helper()

	require.NotEmpty(t, reason, "reason must not be empty")

	x := &expected{
		t:  t,
		tb: t,
	}

	x.tb.Cleanup(func() {
		if x.failed.Load() {
			x.tb.Logf("Test failed as expected: %s", reason)
			return
		}

		x.tb.Fatalf("Test passed unexpectedly: %s", reason)
	})

	return x
}

// expected wraps TB with expected failure logic.
type expected struct {
	tb     testtb.TB
	t      testtb.T
	failed atomic.Bool
}

// Failed reports whether the function has failed.
func (x *expected) Parallel() {
	x.t.Parallel()
}

// Failed reports whether the function has failed.
func (x *expected) Deadline() (time.Time, bool) {
	return x.t.Deadline()
}

// Failed reports whether the function has failed.
func (x *expected) Run(s string, f func(*testing.T)) bool {
	failed := x.t.Run(s, f)

	if failed {
		x.Fail()
	}

	return failed
}

// Failed reports whether the function has failed.
func (x *expected) Failed() bool {
	return x.failed.Load()
}

// Fail marks the function as having failed but continues execution.
func (x *expected) Fail() {
	// we overload this method because we can't set testing.common.failed/skipped/etc fields
	x.failed.Store(true)
}

// Error is equivalent to Log followed by Fail.
func (x *expected) Error(args ...any) {
	x.Log(args...)
	x.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (x *expected) Errorf(format string, args ...any) {
	x.Logf(format, args...)
	x.Fail()
}

// FailNow marks the function as having failed and stops its execution.
func (x *expected) FailNow() {
	x.Fail()

	// we can't use runtime.Goexit because we can't set testing.common.failed/skipped/etc fields
	x.SkipNow()
}

// Fatal is equivalent to Log followed by FailNow.
func (x *expected) Fatal(args ...any) {
	x.Log(args...)
	x.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (x *expected) Fatalf(format string, args ...any) {
	x.Logf(format, args...)
	x.FailNow()
}

// Below methods are delegated as-is.

// Cleanup registers a function to be called when the test (or subtest) and all its subtests complete.
func (x *expected) Cleanup(f func()) { x.tb.Cleanup(f) }

// Helper marks the calling function as a test helper function.
func (x *expected) Helper() { x.tb.Helper() }

// Log formats its arguments using default formatting, analogous to Println, and records the text in the error log.
func (x *expected) Log(args ...any) { x.tb.Log(args...) }

// Logf formats its arguments according to the format, analogous to Printf, and records the text in the error log.
func (x *expected) Logf(format string, args ...any) { x.tb.Logf(format, args...) }

// Name returns the name of the running (sub-) test or benchmark.
func (x *expected) Name() string { return x.tb.Name() }

// Setenv calls os.Setenv(key, value) and uses Cleanup to restore the environment variable to its original value after the test.
func (x *expected) Setenv(key, value string) { x.tb.Setenv(key, value) }

// Skip is equivalent to Log followed by SkipNow.
func (x *expected) Skip(args ...any) { x.tb.Skip(args...) }

// Skipf is equivalent to Logf followed by SkipNow.
func (x *expected) Skipf(format string, args ...any) { x.tb.Skipf(format, args...) }

// SkipNow marks the test as having been skipped and stops its execution.
func (x *expected) SkipNow() { x.tb.SkipNow() }

// Skipped reports whether the test was skipped.
func (x *expected) Skipped() bool { return x.tb.Skipped() }

// TempDir returns a temporary directory for the test to use.
func (x *expected) TempDir() string { return x.tb.TempDir() }
