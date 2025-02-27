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
	PostgreSQL = "PostgreSQL 16.8 (Debian 16.8-1.pgdg120+1) on x86_64-pc-linux-gnu, " +
		"compiled by gcc (Debian 12.2.0-14) 12.2.0, 64-bit"

		// DocumentDB is a version of DocumentDB this version of FerretDB is compatible with.
	DocumentDB = "0.102.0 gitref: HEAD sha:f7539ee buildId:0"
)
