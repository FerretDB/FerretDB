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

// Package bson2 implements encoding and decoding of BSON as defined by https://bsonspec.org/spec.html
// and https://www.mongodb.com/docs/manual/reference/bson-types/.
//
// # Types
//
// The following BSON types are supported:
//
//	BSON                Go
//
//	Document            *bson2.Document or bson2.RawDocument
//	Array               *bson2.Array    or bson2.RawArray
//
//	Double              float64
//	String              string
//	Binary data         bson2.Binary
//	ObjectId            bson2.ObjectID
//	Boolean             bool
//	Date                time.Time
//	Null                bson2.NullType
//	Regular Expression  bson2.Regex
//	32-bit integer      int32
//	Timestamp           bson2.Timestamp
//	64-bit integer      int64
//
// Composite types (Document and Array) are passed by pointers.
// Raw composite type and scalars are passed by values.
package bson2

import (
	"time"

	"github.com/cristalhq/bson/bsonproto"
)

type (
	ScalarType = bsonproto.ScalarType
	Binary     = bsonproto.Binary
	ObjectID   = bsonproto.ObjectID
	NullType   = bsonproto.NullType
	Regex      = bsonproto.Regex
	Timestamp  = bsonproto.Timestamp
)

var (
	ErrDecodeShortInput   = bsonproto.ErrDecodeShortInput
	ErrDecodeInvalidInput = bsonproto.ErrDecodeInvalidInput
)

// Type represents a BSON type.
type Type interface {
	ScalarType | CompositeType
}

// CompositeType represents a BSON composite type (including raw types).
type CompositeType interface {
	*Document | *Array | RawDocument | RawArray
}

// validType checks if v is a valid BSON type (including raw types).
func validType(v any) bool {
	switch v.(type) {
	case *Document:
	case RawDocument:
	case *Array:
	case RawArray:
	case float64:
	case string:
	case Binary:
	case ObjectID:
	case bool:
	case time.Time:
	case NullType:
	case Regex:
	case int32:
	case Timestamp:
	case int64:

	default:
		return false
	}

	return true
}

//go:generate ../../bin/stringer -linecomment -type tag

type tag byte

const (
	tagFloat64         = tag(0x01) // Float64
	tagString          = tag(0x02) // String
	tagDocument        = tag(0x03) // Document
	tagArray           = tag(0x04) // Array
	tagBinary          = tag(0x05) // Binary
	tagUndefined       = tag(0x06) // Undefined
	tagObjectID        = tag(0x07) // ObjectID
	tagBool            = tag(0x08) // Bool
	tagTime            = tag(0x09) // Time
	tagNull            = tag(0x0a) // Null
	tagRegex           = tag(0x0b) // Regex
	tagDBPointer       = tag(0x0c) // DBPointer
	tagJavaScript      = tag(0x0d) // JavaScript
	tagSymbol          = tag(0x0e) // Symbol
	tagJavaScriptScope = tag(0x0f) // JavaScriptScope
	tagInt32           = tag(0x10) // Int32
	tagTimestamp       = tag(0x11) // Timestamp
	tagInt64           = tag(0x12) // Int64
	tagDecimal         = tag(0x13) // Decimal
	tagMinKey          = tag(0xff) // MinKey
	tagMaxKey          = tag(0x7f) // MaxKey
)
