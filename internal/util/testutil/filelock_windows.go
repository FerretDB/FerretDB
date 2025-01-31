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
	"os"
	"testing"
)

// flock is not yet implemented on Windows.
func flock(tb testing.TB, f *os.File, op string) {
	// If we need to implement it, we probably should remove our own code and use
	// https://pkg.go.dev/github.com/rogpeppe/go-internal/lockedfile
	// (that is already used by tools).
	panic("flock is not yet implemented on Windows")
}
