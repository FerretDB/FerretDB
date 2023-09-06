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

package pool

import "net/url"

// setDefaultValue sets default query parameters.
//
// Keep it in sync with docs.
//
//nolint:unused // for now
func setDefaultValues(values url.Values) {
	if !values.Has("pool_max_conns") {
		// the default is too low
		values.Set("pool_max_conns", "50")
	}

	values.Set("application_name", "FerretDB")

	// That only affects text protocol; pgx mostly uses a binary one.
	// See:
	//   - https://github.com/jackc/pgx/issues/520
	//   - https://github.com/jackc/pgx/issues/789
	//   - https://github.com/jackc/pgx/issues/863
	//
	// TODO https://github.com/FerretDB/FerretDB/issues/43
	values.Set("timezone", "UTC")
}
