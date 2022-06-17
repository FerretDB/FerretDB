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

//go:build testcover
// +build testcover

package main

import (
	"os"
	"testing"
)

// TestCover allows us to run FerretDB with coverage enabled.
func TestCover(t *testing.T) {
	main()
}

// TestMain ensures that command-line flags are initialized correctly when FerretDB is run with coverage enabled.
func TestMain(m *testing.M) {
	initFlags()
	os.Exit(m.Run())
}
