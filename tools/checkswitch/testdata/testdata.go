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

// Package testdata provides vet tool test data.
package testdata

import (
	"time"

	"./types"
)

func test(v any) {
	switch v.(type) {
	case *types.Document:
	case *types.Array:
	case float64, int32, int64: // multiple types
	case int8: // unknown
	case string:
	case types.Binary:
	case types.ObjectID:
	case bool:
	case time.Time:
	case types.NullType:
	case types.Regex:
	case types.Timestamp:
	default:
	}

	switch v.(type) { // want "Document should go before Array in the switch"
	case *types.Array:
	case *types.Document:
	case float64:
	case string:
	case types.Binary:
	case types.ObjectID:
	case bool:
	case time.Time:
	case types.NullType:
	case types.Regex:
	case int32:
	case types.Timestamp:
	case int64:
	default:
	}

	switch v.(type) { // want "int32 should go before int64 in the switch"
	case *types.Document:
	case *types.Array:
	case float64, int64, int32:
	case string:
	case types.Binary:
	case types.ObjectID:
	case bool:
	case time.Time:
	case types.NullType:
	case types.Regex:
	case types.Timestamp:
	default:
	}
}
