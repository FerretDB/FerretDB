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
// All BSON types have three representations in FerretDB:
//
//  1. As they are used in "business logic" / handlers - `types` package.
//  2. As they are used in the wire protocol implementation - `bson` package.
//  3. As they are used to store data in PostgreSQL - `fjson` package.
//
// The reason for that is a separation of concerns: to avoid method names clashes, to simplify type asserts,
// to make refactorings and optimizations easier, etc.
//
// Mapping
//
// Composite types (passed by pointers)
//  *types.Document  *bson.Document       *fjson.documentType   Document
//  *types.Array     *bson.arrayType      *fjson.arrayType      Array
//
// Scalar types (passed by values)
//  float64          *bson.doubleType     *fjson.doubleType     64-bit binary floating point
//  string           *bson.stringType     *fjson.stringType     UTF-8 string
//  types.Binary     *bson.binaryType     *fjson.binaryType     Binary data
//  types.ObjectID   *bson.objectIDType   *fjson.objectIDType   ObjectId
//  bool             *bson.boolType       *fjson.boolType       Boolean
//  time.Time        *bson.dateTimeType   *fjson.dateTimeType   UTC datetime
//  types.NullType   *bson.nullType       *fjson.nullType       Null
//  types.Regex      *bson.regexType      *fjson.regexType      Regular expression
//  int32            *bson.int32Type      *fjson.int32Type      32-bit integer
//  types.Timestamp  *bson.timestampType  *fjson.timestampType  Timestamp
//  int64            *bson.int64Type      *fjson.int64Type      64-bit integer
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

// validateValue validates value.
//
// TODO https://github.com/FerretDB/FerretDB/issues/260
func validateValue(value any) error {
	switch value := value.(type) {
	case *Document:
		return value.validate()
	case *Array:
		// It is impossible to construct invalid Array using exported function, methods, or type conversions,
		// so no need to revalidate it.
		return nil
	case float64:
		return nil
	case string:
		return nil
	case Binary:
		return nil
	case ObjectID:
		return nil
	case bool:
		return nil
	case time.Time:
		return nil
	case NullType:
		return nil
	case Regex:
		return nil
	case int32:
		return nil
	case Timestamp:
		return nil
	case int64:
		return nil
	default:
		return fmt.Errorf("types.validateValue: unsupported type: %[1]T (%[1]v)", value)
	}
}

// deepCopy returns a deep copy of the given value.
func deepCopy(value any) any {
	if value == nil {
		panic("types.deepCopy: nil value")
	}

	switch value := value.(type) {
	case *Document:
		keys := make([]string, len(value.keys))
		copy(keys, value.keys)

		m := make(map[string]any, len(value.m))
		for k, v := range value.m {
			m[k] = deepCopy(v)
		}

		return &Document{
			keys: keys,
			m:    m,
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
		panic(fmt.Sprintf("types.deepCopy: unsupported type: %[1]T (%[1]v)", value))
	}
}
