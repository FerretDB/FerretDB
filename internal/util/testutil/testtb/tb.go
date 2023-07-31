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

// Package testtb provides a common testing interface.
//
// It is a separate package to avoid dependency cycles.
package testtb

import (
	"testing"
	"time"
)

// TB is a copy of testing.TB without a private method.
// It allows us to implement it.
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

type T interface {
	TB

	Run(string, func(t *testing.T)) bool
	Parallel()
	Deadline() (time.Time, bool)
}

// check interfaces
var (
	_ TB = (*testing.T)(nil)
	_ TB = (*testing.B)(nil)
	_ TB = (*testing.F)(nil)
	_ TB = (testing.TB)(nil)

	_ T = (*testing.T)(nil)
)
