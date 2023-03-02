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

//go:build ferretdb_testcover

package main

import (
	"flag"
	"os"
	"testing"

	"github.com/alecthomas/kong"
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/internal/util/must"
)

// TestCover allows us to run FerretDB with coverage enabled.
func TestCover(t *testing.T) {
	run()
}

// TestMain ensures that command-line flags are initialized correctly when FerretDB is run with coverage enabled.
func TestMain(m *testing.M) {
	// Split flags for kong and for `go test` by "--". For example:
	// bin/ferretdb-local -test.coverprofile=cover.txt -- --test-records-dir=records --mode=diff-normal --listen-addr=:27017
	// forKong: --test-records-dir=records --mode=diff-normal --listen-addr=:27017
	// forTest: -test.coverprofile=cover.txt
	forKong := os.Args[1:]
	forTest := []string{}
	i := slices.Index(os.Args, "--")
	if i != -1 {
		forKong = os.Args[i+1:]
		forTest = os.Args[1:i]
	}

	parser := must.NotFail(kong.New(&cli, kongOptions...))

	_, err := parser.Parse(forKong)
	parser.FatalIfErrorf(err)

	flag.CommandLine.Parse(forTest)

	os.Exit(m.Run())
}
