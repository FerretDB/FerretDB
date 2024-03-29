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

package integration

import (
	"flag"
	"os"
	"testing"

	"github.com/FerretDB/FerretDB/integration/setup"
)

// TestMain is the entry point for all integration tests.
func TestMain(m *testing.M) {
	flag.Parse()

	var code int

	// ensure that Shutdown runs for any exit code or panic
	func() {
		// make `go test -list=.` work without side effects
		if flag.Lookup("test.list").Value.String() == "" {
			setup.Startup()
			defer setup.Shutdown()
		}

		code = m.Run()
	}()

	os.Exit(code)
}
