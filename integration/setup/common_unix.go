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

//go:build unix

package setup

import (
	"testing"
)

// listenUnix returns temporary Unix domain socket path for that test.
func listenUnix(tb testing.TB) string {
	// The commented out code does not generate valid Unix domain socket path on macOS (at least).
	// Maybe the argument is too long?
	// TODO https://github.com/FerretDB/FerretDB/issues/1295
	// return filepath.Join(tb.TempDir(), "ferretdb.sock")

	return ""
}
