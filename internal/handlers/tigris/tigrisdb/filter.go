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

// Package filter provides helpers to simplify generation of filters for Tigris driver.
//
// It's similar to tigrisdata/tigris-client-go/filter but without generics as they don't fit us - we need to call
// tjson.Marshal to encode tjson data types correctly.
package tigrisdb

import (
	"encoding/json"
	"fmt"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/tjson"
)

// All represents filter which includes all the documents of the collection.
var All = driver.Filter(`{}`)

// Eq composes 'equal' operation for Tigris, like in tigrisdata/tigris-client-go/filter but without generics.
// It marshals the value (which is always of one of the tjson types) to JSON, and it'll be used as RawMessage later.
func Eq(field string, value any) map[string]any {
	res, err := tjson.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("problem marshalling value as tjson: %v", err))
	}

	return map[string]any{field: map[string]json.RawMessage{"$eq": res}}
}

// And composes 'and' operation.
// Result is equivalent to: (ops[0] && ... && ops[len(ops-1]).
func And(ops ...map[string]any) map[string]any {
	return map[string]any{"$and": ops}
}

// Or composes 'or' operation.
// Result is equivalent to: (ops[0] || ... || ops[len(ops-1]).
func Or(ops ...map[string]any) map[string]any {
	return map[string]any{"$or": ops}
}
