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
// All BSON types have five representations in FerretDB:
//
//  1. As they are used in "business logic" / handlers - `types` package.
//  2. As they are used for logging - `fjson` package.
//  3. As they are used in the wire protocol implementation - `bson` package.
//  4. As they are used to store data in PostgreSQL - `pjson` package.
//  5. As they are used to store data in Tigris - `tjson` package.
//
// The reason for that is a separation of concerns: to avoid method names clashes, to simplify type asserts,
// to make refactorings and optimizations easier, etc.
//
// # Mapping
//
// Composite types (passed by pointers)
//
//	*types.Document  *bson.Document       *pjson.documentType   Document
//	*types.Array     *bson.arrayType      *pjson.arrayType      Array
//
// Scalar types (passed by values)
//
//	float64          *bson.doubleType     *pjson.doubleType     64-bit binary floating point
//	string           *bson.stringType     *pjson.stringType     UTF-8 string
//	types.Binary     *bson.binaryType     *pjson.binaryType     Binary data
//	types.ObjectID   *bson.objectIDType   *pjson.objectIDType   ObjectId
//	bool             *bson.boolType       *pjson.boolType       Boolean
//	time.Time        *bson.dateTimeType   *pjson.dateTimeType   UTC datetime
//	types.NullType   *bson.nullType       *pjson.nullType       Null
//	types.Regex      *bson.regexType      *pjson.regexType      Regular expression
//	int32            *bson.int32Type      *pjson.int32Type      32-bit integer
//	types.Timestamp  *bson.timestampType  *pjson.timestampType  Timestamp
//	int64            *bson.int64Type      *pjson.int64Type      64-bit integer
package types

import (
	"fmt"
	"time"
)

// MaxDocumentLen is the maximum BSON object size.
const MaxDocumentLen = 16777216

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
// TODO remove once we have go-sumtype equivalent?
type CompositeTypeInterface interface {
	CompositeType

	GetByPath(path Path) (any, error)
	RemoveByPath(path Path)

	compositeType() // seal for go-sumtype
}

//go-sumtype:decl CompositeTypeInterface

type (
	// NullType represents BSON type Null.
	//
	// Most callers should use types.Null value instead.
	NullType struct{}
)

// Null represents BSON value Null.
var Null = NullType{}

// deepCopy returns a deep copy of the given value.
func deepCopy(value any) any {
	if value == nil {
		panic("types.deepCopy: nil value")
	}

	switch value := value.(type) {
	case *Document:
		fields := make([]field, len(value.fields))
		for i, f := range value.fields {
			fields[i] = field{
				key:   f.key,
				value: deepCopy(f.value),
			}
		}

		return &Document{fields}

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
		panic(fmt.Sprintf("types.deepCopy: unsupported type: %[1]T (%[1]v)", value))
	}
}
