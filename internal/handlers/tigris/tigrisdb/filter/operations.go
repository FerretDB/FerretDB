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

package filter

// List of supported operations.
const (
	and = "$and"
	or  = "$or"
)

// And composes 'and' operation.
// Result is equivalent to: (ops[0] && ... && ops[len(ops-1]).
func And(ops ...Expr) Expr {
	return Expr{and: ops}
}

// Or composes 'or' operation.
// Result is equivalent to: (ops[0] || ... || ops[len(ops-1]).
func Or(ops ...Expr) Expr {
	return Expr{or: ops}
}
