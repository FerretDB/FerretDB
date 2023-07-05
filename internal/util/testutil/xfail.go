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

func XFail(t TB, reason string) TB {
	t.Helper()

	require.NotEmpty(t, reason, "reason must not be empty")

	x := &xfail{
		t: t,
	}

	x.t.Cleanup(func() {
		if x.failed.Load() {
			x.t.Skipf("Test failed as expected: %s", reason)
		}

		x.t.Fatalf("Test passed unexpectedly. See %s", reason)
	})

	return x
}

type xfail struct {
	t      TB
	failed atomic.Bool
}

func (x *xfail) Error(args ...any) {
	x.Log(args...)
	x.Fail()
}

func (x *xfail) Errorf(format string, args ...any) {
	x.Logf(format, args...)
	x.Fail()
}

func (x *xfail) Fail() {
	x.failed.Store(true)
}

func (x *xfail) FailNow() {
	x.Fail()

	// TODO
	x.SkipNow()
}

func (x *xfail) Failed() bool {
	return x.failed.Load()
}

func (x *xfail) Fatal(args ...any) {
	x.Log(args...)
	x.FailNow()
}

func (x *xfail) Fatalf(format string, args ...any) {
	x.Logf(format, args...)
	x.FailNow()
}

func (x *xfail) Cleanup(f func())                 { x.t.Cleanup(f) }
func (x *xfail) Helper()                          { x.t.Helper() }
func (x *xfail) Log(args ...any)                  { x.t.Log(args...) }
func (x *xfail) Logf(format string, args ...any)  { x.t.Logf(format, args...) }
func (x *xfail) Name() string                     { return x.t.Name() }
func (x *xfail) Setenv(key, value string)         { x.t.Setenv(key, value) }
func (x *xfail) Skip(args ...any)                 { x.t.Skip(args...) }
func (x *xfail) Skipf(format string, args ...any) { x.t.Skipf(format, args...) }
func (x *xfail) SkipNow()                         { x.t.SkipNow() }
func (x *xfail) Skipped() bool                    { return x.t.Skipped() }
func (x *xfail) TempDir() string                  { return x.t.TempDir() }

// check interfaces
var (
	_ TB = (*testing.T)(nil)
	_ TB = (*testing.B)(nil)
	_ TB = (*testing.F)(nil)
	_ TB = (testing.TB)(nil)
)
