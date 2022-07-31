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

import (
	"encoding/json"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
)

// comparison represents comparison operators for Tigris, like comparison in tigrisdata/tigris-client-go/filter.
type comparison struct {
	Gt  any `json:"$gt,omitempty"`
	Gte any `json:"$gte,omitempty"`
	Lt  any `json:"$lt,omitempty"`
	Lte any `json:"$lte,omitempty"`
	Ne  any `json:"$ne,omitempty"`
	Eq  any `json:"$eq,omitempty"`
}

// Eq composes 'equal' operation for Tigris, like in tigrisdata/tigris-client-go/filter but without generics.
// It marshals the value to JSON, so it'll be used as RawJSON later.
// The value must be a tjson value. If it's not a tjson value, there will be a panic.
// Result is equivalent to: field == filterValue.
func Eq(field string, value any) Expr {
	var res any
	if objID, ok := value.(types.ObjectID); ok {
		r, err := tjson.Marshal(objID)
		if err != nil {
			panic(fmt.Sprintf("problem marshalling value as tjson: %v", err))
		}
		res = json.RawMessage(r)
	} else {
		res = value
	}

	return Expr{field: comparison{Eq: res}}
}
