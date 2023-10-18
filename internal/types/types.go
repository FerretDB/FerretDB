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

// Package types provides Go types matching BSON types that don't have built-in Go equivalents.
//
// All BSON types have four representations in FerretDB:
//
//  1. As they are used in "business logic" / handlers - `types` package.
//  2. As they are used for logging - `fjson` package.
//  3. As they are used in the wire protocol implementation - `bson` package.
//  4. As they are used to store data in SQL based databases - `sjson` package.
//
// The reason for that is a separation of concerns: to avoid method names clashes, to simplify type asserts,
// to make refactorings and optimizations easier, etc.
//
// # Mapping
//
// Composite types (passed by pointers)
//
//	Alias      types package    Description
//
//	object     *types.Document  Document
//	array      *types.Array     Array
//
// Scalar types (passed by values)
//
//	Alias      types package    Description
//
//	double     float64          64-bit binary floating point
//	string     string           UTF-8 string
//	binData    types.Binary     Binary data
//	objectId   types.ObjectID   Object ID
//	bool       bool             Boolean
//	date       time.Time        UTC datetime
//	null       types.NullType   Null
//	regex      types.Regex      Regular expression
//	int        int32            32-bit integer
//	timestamp  types.Timestamp  Timestamp
//	long       int64            64-bit integer
//
//nolint:dupword // false positive
package types

import (
	"fmt"
	"time"
)

// MaxDocumentLen is the maximum BSON object size.
const MaxDocumentLen = 16 * 1024 * 1024 // 16 MiB = 16777216 bytes

// MaxSafeDouble is the maximum double value that can be represented precisely.
const MaxSafeDouble = float64(1<<53 - 1) // 52bit mantissa max value = 9007199254740991

// ScalarType represents scalar type.
type ScalarType interface {
	float64 | string | Binary | ObjectID | bool | time.Time | NullType | Regex | int32 | Timestamp | int64
}

// CompositeType represents composite type - *Document or *Array.
type CompositeType interface {
	*Document | *Array
}

// Type represents any BSON type (scalar or composite).
type Type interface {
	ScalarType | CompositeType
}

// CompositeTypeInterface consists of Document and Array.
type CompositeTypeInterface interface {
	CompositeType

	GetByPath(path Path) (any, error)
	RemoveByPath(path Path)

	compositeType() // seal for go-sumtype
}

//go-sumtype:decl CompositeTypeInterface

// assertType panics if value is not a BSON type (scalar or composite).
//
// It should be used in places where we can't use compile-time check via type parameters (generics).
// For example, `NewArray(42)` should panic, because `int` is not a BSON type.
//
// It should check only type, not value.
func assertType(value any) {
	switch value := value.(type) {
	case *Document, *Array:
		return
	case float64, string, Binary, ObjectID, bool, time.Time, NullType, Regex, int32, Timestamp, int64:
		return
	case nil:
		panic("types: unexpected nil type")
	default:
		panic(fmt.Sprintf("types: unexpected type %[1]T (%#[1]v)", value))
	}
}

// isScalar checks if v is a BSON scalar value.
//
// It panics if value is not a BSON type.
func isScalar(value any) bool {
	assertType(value)

	switch value.(type) {
	case float64, string, Binary, ObjectID, bool, time.Time, NullType, Regex, int32, Timestamp, int64:
		return true
	}

	return false
}

// deepCopy returns a deep copy of the given value.
func deepCopy(value any) any {
	switch value := value.(type) {
	case *Document:
		fields := make([]field, len(value.fields))
		for i, f := range value.fields {
			fields[i] = field{
				key:   f.key,
				value: deepCopy(f.value),
			}
		}

		return &Document{
			fields:   fields,
			recordID: value.recordID,
		}

	case *Array:
		s := make([]any, len(value.s))
		for i, v := range value.s {
			s[i] = deepCopy(v)
		}
		return &Array{
			s: s,
		}

	case float64:
		return value
	case string:
		return value
	case Binary:
		b := make([]byte, len(value.B))
		copy(b, value.B)
		return Binary{
			Subtype: value.Subtype,
			B:       b,
		}
	case ObjectID:
		return value
	case bool:
		return value
	case time.Time:
		return value
	case NullType:
		return value
	case Regex:
		return value
	case int32:
		return value
	case Timestamp:
		return value
	case int64:
		return value

	default:
		panic(fmt.Sprintf("types.deepCopy: unexpected type %[1]T (%#[1]v)", value))
	}
}
