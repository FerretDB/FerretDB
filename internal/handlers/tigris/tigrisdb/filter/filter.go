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
package filter

import (
	"encoding/json"

	"github.com/tigrisdata/tigris-client-go/driver"
)

// All represents filter which includes all the documents of the collection.
var All = driver.Filter(`{}`)

// Expr represents a filter expression for Tigris.
type Expr map[string]any

// Build "encodes" expression to be used with Tigris driver.
func (e Expr) Build() (driver.Filter, error) {
	if e == nil {
		return nil, nil
	}

	return json.Marshal(e)
}
