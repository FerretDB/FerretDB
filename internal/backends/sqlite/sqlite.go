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

// Package sqlite provides SQLite backend.
package sqlite

// https://www.sqlite.org/limits.html#max_variable_number
const maxPlaceholders = 1000

func placeholders(n int) []string {
	if n > maxPlaceholders {
		panic("too many placeholders")
	}

	r := make([]string, n)
	for i := 0; i < n; i++ {
		r[i] = "?"
	}

	return r
}
