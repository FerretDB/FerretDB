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

	"github.com/tigrisdata/tigris-client-go/driver"
)

var (
	//All represents filter which includes all the documents of the collection
	All = driver.Filter(`{}`)
)

// Expr represents a filter expression for Tigris.
type Expr map[string]any

// Build "encodes" expression to be used with Tigris driver.
func (e Expr) Build() (driver.Filter, error) {
	if e == nil {
		return nil, nil
	}
	b, err := json.Marshal(e)
	return b, err
}
