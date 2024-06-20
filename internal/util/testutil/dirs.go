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
	"path/filepath"
	"testing"
)

var (
	RootDir        string // FerretDB root directory
	BinDir         string // <root>/bin directory
	BuildCertsDir  string // <root>/build/certs directory
	IntegrationDir string // <root>/integration directory
	TmpRecordsDir  string // <root>/tmp/records directory
)

func init() {
	if !testing.Testing() {
		panic("testutil package must be used only by tests")
	}

	// We can't use runtime.Caller because file path might be relative.
	// See also similar code in the tools module.

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		if _, err = os.Stat(filepath.Join(dir, ".git")); err == nil {
			break
		}

		dir = filepath.Dir(dir)
		if dir == "/" {
			panic("failed to locate .git directory")
		}
	}

	RootDir = dir

	BinDir = filepath.Join(RootDir, "bin")
	BuildCertsDir = filepath.Join(RootDir, "build", "certs")
	IntegrationDir = filepath.Join(RootDir, "integration")
	TmpRecordsDir = filepath.Join(RootDir, "tmp", "records")
}
