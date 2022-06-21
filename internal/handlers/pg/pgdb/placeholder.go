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

package pgdb

import "strconv"

// Placeholder stores the number of the relevant placeholder of the query.
type Placeholder int

// Next increases the identifier value for the next variable in the PostgreSQL query.
func (p *Placeholder) Next() string {
	*p++
	return "$" + strconv.Itoa(int(*p))
}
