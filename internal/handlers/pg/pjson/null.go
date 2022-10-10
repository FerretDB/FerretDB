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

package pjson

import (
	"fmt"

	"github.com/FerretDB/FerretDB/internal/types"
)

// nullType represents BSON Null type.
type nullType types.NullType

// pjsontype implements pjsontype interface.
func (*nullType) pjsontype() {}

// UnmarshalJSON implements pjsontype interface.
func (*nullType) UnmarshalJSON(data []byte) error {
	panic(fmt.Sprintf("must not be called, was called with %s", string(data)))
}

// MarshalJSON implements pjsontype interface.
func (*nullType) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

// check interfaces
var (
	_ pjsontype = (*nullType)(nil)
)
