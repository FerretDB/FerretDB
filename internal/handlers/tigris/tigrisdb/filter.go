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

package tigrisdb

import "github.com/tigrisdata/tigris-client-go/filter"

// filterValue represents a value in a filter expression, like value in tigrisdata/tigris-client-go/filter.
type filterValue any

// filterComparison represents comparison operators for Tigris, like comparison in tigrisdata/tigris-client-go/filter.
type filterComparison struct {
	Gt  filterValue `json:"$gt,omitempty"`
	Gte filterValue `json:"$gte,omitempty"`
	Lt  filterValue `json:"$lt,omitempty"`
	Lte filterValue `json:"$lte,omitempty"`
	Ne  filterValue `json:"$ne,omitempty"`
	Eq  filterValue `json:"$eq,omitempty"`
}

// FilterEq composes 'equal' operation for Tigris, like in tigrisdata/tigris-client-go/filter but without generics.
// Result is equivalent to: field == filterValue.
func FilterEq(field string, value any) filter.Expr {
	return filter.Expr{field: filterComparison{Eq: value}}
}
