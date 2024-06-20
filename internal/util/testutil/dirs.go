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
	"path/filepath"
	"runtime"
	"testing"
)

var (
	RootDir       string // directory containing the main `go.mod` file
	BuildCertsDir string
)

func init() {
	if !testing.Testing() {
		panic("testutil package must be used only by tests")
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("testutil initialization failed")
	}

	RootDir = filepath.Join(filepath.Dir(file), "..", "..", "..")
	BuildCertsDir = filepath.Join(RootDir, "build", "certs")
}
