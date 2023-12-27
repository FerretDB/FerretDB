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

const (
	TagFloat64   = byte(0x01)
	TagString    = byte(0x02)
	TagDocument  = byte(0x03)
	TagArray     = byte(0x04)
	TagBinary    = byte(0x05)
	TagObjectID  = byte(0x07)
	TagBool      = byte(0x08)
	TagTime      = byte(0x09)
	TagNullType  = byte(0x0a)
	TagRegex     = byte(0x0b)
	TagInt32     = byte(0x10)
	TagTimestamp = byte(0x11)
	TagInt64     = byte(0x12)
)
