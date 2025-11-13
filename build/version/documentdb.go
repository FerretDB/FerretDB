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

package version

import "runtime"

const (
	// DocumentDB is a version of DocumentDB this version of FerretDB is compatible with.
	DocumentDB = "0.107.0 gitref: ferretdb sha:e63835403d buildId:0"

	// DocumentDBURL points to the release page of the DocumentDB version above.
	DocumentDBURL = "https://github.com/FerretDB/documentdb/releases/tag/v0.108.0-ferretdb-2.8.0"
)

// DocumentDBSafeToUpdate represents versions of DocumentDB that FerretDB can update.
var DocumentDBSafeToUpdate = []string{
	"0.102.0 gitref: HEAD sha:80462f5 buildId:0",    // v2.1.0
	"0.103.0 gitref: HEAD sha:7514232 buildId:0",    // v2.2.0
	"0.104.0 gitref: HEAD sha:2045d0e buildId:0",    // v2.3.0, v2.3.1
	"0.105.0 gitref: HEAD sha:8453d93b buildId:0",   // v2.4.0
	"0.106.0 gitref: HEAD sha:beb9d25d98 buildId:0", // v2.5.0
	// FerretDB v2.6.0 wasn't released
	"0.107.0 gitref: HEAD sha:e63835403d buildId:0", // v2.7.0
}

// PostgreSQLTest is a version of PostgreSQL used by tests.
var PostgreSQLTest string

func init() {
	arch := "x86_64-pc-linux-gnu"
	if runtime.GOARCH == "arm64" {
		arch = "aarch64-unknown-linux-gnu"
	}

	PostgreSQLTest = "PostgreSQL 17.6 (Debian 17.6-2.pgdg12+1) on " + arch + ", " +
		"compiled by gcc (Debian 12.2.0-14+deb12u1) 12.2.0, 64-bit"
}
