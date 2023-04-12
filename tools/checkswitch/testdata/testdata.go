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
	"./types"
	"time"
)

func testCorrect(v any) {
	switch v.(type) {
	case *types.Document:
	case *types.Array:
	case float64, int32, int64: // multiple types
	case int8: // unexpected type
	case string:
	case types.Binary:
	case types.ObjectID:
	case bool:
	case time.Time:
	case types.NullType:
	case types.Regex:
	case types.Timestamp:
	}
}

func testIncorrectSimple(v any) {
	switch v.(type) { // want "Document should go before Array in the switch"
	case *types.Array:
	case *types.Document:
	}
}

func testIncorrectMixed(v any) {
	switch v.(type) { // want "Document should go before Time in the switch"
	case time.Time:
	case *types.Document:
	}
}

func testIncorrectMultiple(v any) {
	switch v.(type) { // want "int32 should go before int64 in the switch"
	case float64, int64, int32:
	}
}
