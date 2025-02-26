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
	// DocumentDB is a version of DocumentDB this version of FerretDB is compatible with,
	// as reporter by `documentdb_api.binary_extended_version()`.
	DocumentDB = "0.102.0 gitref: HEAD sha:39ec23d buildId:0"

	// DocumentDBURL points to the release page of the DocumentDB version above.
	DocumentDBURL = "https://github.com/FerretDB/documentdb/releases/tag/v0.102.0-ferretdb-2.0.0-rc.2"
)
