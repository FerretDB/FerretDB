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

const (
	// PostgreSQL is a version of PostgreSQL this version of FerretDB is compatible with.
	PostgreSQL = "PostgreSQL 17.4 (Debian 17.4-1.pgdg120+2) on x86_64-pc-linux-gnu, " +
		"compiled by gcc (Debian 12.2.0-14) 12.2.0, 64-bit"

	// DocumentDB is a version of DocumentDB this version of FerretDB is compatible with.
	// DocumentDB = "0.103.0 gitref: ferretdb sha:c501656 buildId:0"

	// FIXME
	DocumentDB = "0.103.0 gitref: toggles sha:bd16dba buildId:0"

	// DocumentDBURL points to the release page of the DocumentDB version above.
	DocumentDBURL = "https://github.com/FerretDB/documentdb/releases/tag/v0.103.0-ferretdb-2.2.0"
)

// DocumentDBSafeToUpdate represents versions of DocumentDB that FerretDB can update.
var DocumentDBSafeToUpdate = []string{
	"0.102.0 gitref: HEAD sha:80462f5 buildId:0", // v2.1.0
}
