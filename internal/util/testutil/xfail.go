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

package testutil

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// TB is a copy of testing.TB without a private method.
//
//nolint:interfacebloat // that's a copy of existing interface
type TB interface {
	Cleanup(func())
	Error(args ...any)
	Errorf(format string, args ...any)
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Helper()
	Log(args ...any)
	Logf(format string, args ...any)
	Name() string
	Setenv(key, value string)
	Skip(args ...any)
	SkipNow()
	Skipf(format string, args ...any)
	Skipped() bool
	TempDir() string
}

// XFail return a new TB instance that expects the test to fail.
//
// At the end of the test, if it was marked as failed, it will pass instead.
// If it passes, it will be failed, so that XFail call can be removed.
func XFail(t TB, reason string) TB {
	t.Helper()

	require.NotEmpty(t, reason, "reason must not be empty")

	x := &xfail{
		t: t,
	}

	x.t.Cleanup(func() {
		if x.failed.Load() {
			x.t.Logf("Test failed as expected: %s", reason)
			return
		}

		x.t.Fatalf("Test passed unexpectedly: %s", reason)
	})

	return x
}

// xfail wraps TB with expected failure logic.
type xfail struct {
	t      TB
	failed atomic.Bool
}

// Failed reports whether the function has failed.
func (x *xfail) Failed() bool {
	return x.failed.Load()
}

// Fail marks the function as having failed but continues execution.
func (x *xfail) Fail() {
	// we overload this method because we can't set testing.common.failed/skipped/etc fields
	x.failed.Store(true)
}

// Error is equivalent to Log followed by Fail.
func (x *xfail) Error(args ...any) {
	x.Log(args...)
	x.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (x *xfail) Errorf(format string, args ...any) {
	x.Logf(format, args...)
	x.Fail()
}

// FailNow marks the function as having failed and stops its execution.
func (x *xfail) FailNow() {
	x.Fail()

	// we can't use runtime.Goexit because we can't set testing.common.failed/skipped/etc fields
	x.SkipNow()
}

// Fatal is equivalent to Log followed by FailNow.
func (x *xfail) Fatal(args ...any) {
	x.Log(args...)
	x.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (x *xfail) Fatalf(format string, args ...any) {
	x.Logf(format, args...)
	x.FailNow()
}

// Below methods are delegated as-is.

// Cleanup registers a function to be called when the test (or subtest) and all its subtests complete.
func (x *xfail) Cleanup(f func()) { x.t.Cleanup(f) }

// Helper marks the calling function as a test helper function.
func (x *xfail) Helper() { x.t.Helper() }

// Log formats its arguments using default formatting, analogous to Println, and records the text in the error log.
func (x *xfail) Log(args ...any) { x.t.Log(args...) }

// Logf formats its arguments according to the format, analogous to Printf, and records the text in the error log.
func (x *xfail) Logf(format string, args ...any) { x.t.Logf(format, args...) }

// Name returns the name of the running (sub-) test or benchmark.
func (x *xfail) Name() string { return x.t.Name() }

// Setenv calls os.Setenv(key, value) and uses Cleanup to restore the environment variable to its original value after the test.
func (x *xfail) Setenv(key, value string) { x.t.Setenv(key, value) }

// Skip is equivalent to Log followed by SkipNow.
func (x *xfail) Skip(args ...any) { x.t.Skip(args...) }

// Skipf is equivalent to Logf followed by SkipNow.
func (x *xfail) Skipf(format string, args ...any) { x.t.Skipf(format, args...) }

// SkipNow marks the test as having been skipped and stops its execution.
func (x *xfail) SkipNow() { x.t.SkipNow() }

// Skipped reports whether the test was skipped.
func (x *xfail) Skipped() bool { return x.t.Skipped() }

// TempDir returns a temporary directory for the test to use.
func (x *xfail) TempDir() string { return x.t.TempDir() }

// check interfaces
var (
	_ TB = (*testing.T)(nil)
	_ TB = (*testing.B)(nil)
	_ TB = (*testing.F)(nil)
	_ TB = (testing.TB)(nil)
)
