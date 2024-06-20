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
	// RootDir is the FerretDB <root> directory.
	RootDir string
	// BinDir is the <root>/bin directory.
	BinDir string
	// BuildCertsDir is the <root>/build/certs directory.
	BuildCertsDir string
	// IntegrationDir is the <root>/integration directory.
	IntegrationDir string
	// TmpRecordsDir is the <root>/tmp/records directory.
	TmpRecordsDir string
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
