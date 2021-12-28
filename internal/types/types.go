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

// Package types provides Go types matching BSON types without built-in Go equivalents.
//
// All BSON data types have three representations in FerretDB:
//
//  1. As they are used in handlers implementation (types package).
//  2. As they are used in the wire protocol implementation (bson package).
//  3. As they are used to store data in PostgreSQL (fjson package).
//
// The reason for that is a separation of concerns: to avoid method names clashes, to simplify type asserts, etc.
//
// Mapping
//
// Composite types
//  types.Document   *bson.Document  *fjson.Document
//  *types.Array     *bson.Array     *fjson.Array
// Value types
//  float64          *bson.Double     *fjson.Double
//  string           *bson.String     *fjson.String
//  types.Binary     *bson.Binary     *fjson.Binary
//  types.ObjectID   *bson.ObjectID   *fjson.ObjectID
//  bool             *bson.Bool       *fjson.Bool
//  time.Time        *bson.DateTime   *fjson.DateTime
//  any(nil)         any(nil)         any(nil)
//  types.Regex      *bson.Regex      *fjson.Regex
//  int32            *bson.Int32      *fjson.Int32
//  types.Timestamp  *bson.Timestamp  *fjson.Timestamp
//  int64            *bson.Int64      *fjson.Int64
//  TODO Decimal128
//  types.CString    *bson.CString    *fjson.CString
package types

import (
	"fmt"
	"time"
)

// CompositeType represents composite type - Document or *Array.
type CompositeType interface {
	compositeType()
}

//go-sumtype:decl CompositeType

type (
	CString string

	ObjectID [12]byte

	Regex struct {
		Pattern string
		Options string
	}

	Timestamp uint64
)

// validateValue validates value.
func validateValue(value any) error {
	switch value := value.(type) {
	case Document:
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
	case nil:
		return nil
	case Regex:
		return nil
	case int32:
		return nil
	case Timestamp:
		return nil
	case int64:
		return nil
	case CString:
		return nil
	default:
		return fmt.Errorf("types.validateValue: unsupported type: %[1]T (%[1]v)", value)
	}
}
