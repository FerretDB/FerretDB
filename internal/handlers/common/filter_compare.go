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

package common

import (
	"bytes"
	"fmt"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
)

// filterScalarEqual returns true if given scalar values are equal as used by filters.
func filterScalarEqual(a, b any) bool {
	if a == nil {
		panic("a is nil")
	}
	if b == nil {
		panic("b is nil")
	}

	switch a := a.(type) {
	case float64:
		b := b.(float64)
		return a == b
	case string:
		b := b.(string)
		return a == b
	case types.Binary:
		b := b.(types.Binary)
		return a.Subtype == b.Subtype && bytes.Equal(a.B, b.B)
	case types.ObjectID:
		b := b.(types.ObjectID)
		return a == b
	case bool:
		b := b.(bool)
		return a == b
	case time.Time:
		b := b.(time.Time)
		return a.Equal(b)
	case types.NullType:
		_ = b.(types.NullType)
		return true
	case types.Regex:
		b := b.(types.Regex)
		return a == b
	case int32:
		b := b.(int32)
		return a == b
	case types.Timestamp:
		b := b.(types.Timestamp)
		return a == b
	case int64:
		b := b.(int64)
		return a == b
	default:
		panic(fmt.Sprintf("unhandled type %T", a))
	}
}
