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

// Package bson implements encoding and decoding of BSON as defined by https://bsonspec.org/spec.html.
//
// # Types
//
// The following BSON types are supported:
//
//	BSON                Go
//
//	Document/Object     *bson.Document or bson.RawDocument
//	Array               *bson.Array    or bson.RawArray
//
//	Double              float64
//	String              string
//	Binary data         bson.Binary
//	ObjectId            bson.ObjectID
//	Boolean             bool
//	Date                time.Time
//	Null                bson.NullType
//	Regular Expression  bson.Regex
//	32-bit integer      int32
//	Timestamp           bson.Timestamp
//	64-bit integer      int64
//
// Composite types (Document and Array) are passed by pointers.
// Raw composite type and scalars are passed by values.
package bson

import (
	"time"

	"github.com/cristalhq/bson/bsonproto"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

type (
	// ScalarType represents a BSON scalar type.
	//
	// CString is not included as it is not a real BSON type.
	ScalarType = bsonproto.ScalarType

	// Binary represents BSON scalar type binary.
	Binary = bsonproto.Binary

	// BinarySubtype represents BSON Binary's subtype.
	BinarySubtype = bsonproto.BinarySubtype

	// NullType represents BSON scalar type null.
	NullType = bsonproto.NullType

	// ObjectID represents BSON scalar type ObjectID.
	ObjectID = bsonproto.ObjectID

	// Regex represents BSON scalar type regular expression.
	Regex = bsonproto.Regex

	// Timestamp represents BSON scalar type timestamp.
	Timestamp = bsonproto.Timestamp
)

const (
	// BinaryGeneric represents a BSON Binary generic subtype.
	BinaryGeneric = bsonproto.BinaryGeneric

	// BinaryFunction represents a BSON Binary function subtype.
	BinaryFunction = bsonproto.BinaryFunction

	// BinaryGenericOld represents a BSON Binary generic-old subtype.
	BinaryGenericOld = bsonproto.BinaryGenericOld

	// BinaryUUIDOld represents a BSON Binary UUID old subtype.
	BinaryUUIDOld = bsonproto.BinaryUUIDOld

	// BinaryUUID represents a BSON Binary UUID subtype.
	BinaryUUID = bsonproto.BinaryUUID

	// BinaryMD5 represents a BSON Binary MD5 subtype.
	BinaryMD5 = bsonproto.BinaryMD5

	// BinaryEncrypted represents a BSON Binary encrypted subtype.
	BinaryEncrypted = bsonproto.BinaryEncrypted

	// BinaryUser represents a BSON Binary user-defined subtype.
	BinaryUser = bsonproto.BinaryUser
)

// Null represents BSON scalar value null.
var Null = bsonproto.Null

//go:generate ../../bin/stringer -linecomment -type decodeMode

// decodeMode represents a mode for decoding BSON.
type decodeMode int

const (
	_ decodeMode = iota

	// DecodeShallow represents a mode in which only top-level fields/elements are decoded;
	// nested documents and arrays are converted to RawDocument and RawArray respectively,
	// using raw's subslices without copying.
	decodeShallow

	// DecodeDeep represents a mode in which nested documents and arrays are decoded recursively;
	// RawDocuments and RawArrays are never returned.
	decodeDeep
)

var (
	// ErrDecodeShortInput is returned wrapped by Decode functions if the input bytes slice is too short.
	ErrDecodeShortInput = bsonproto.ErrDecodeShortInput

	// ErrDecodeInvalidInput is returned wrapped by Decode functions if the input bytes slice is invalid.
	ErrDecodeInvalidInput = bsonproto.ErrDecodeInvalidInput
)

// SizeCString returns a size of the encoding of v cstring in bytes.
func SizeCString(s string) int {
	return bsonproto.SizeCString(s)
}

// EncodeCString encodes cstring value v into b.
//
// Slice must be at least len(v)+1 ([SizeCString]) bytes long; otherwise, EncodeString will panic.
// Only b[0:len(v)+1] bytes are modified.
func EncodeCString(b []byte, v string) {
	bsonproto.EncodeCString(b, v)
}

// DecodeCString decodes cstring value from b.
//
// If there is not enough bytes, DecodeCString will return a wrapped [ErrDecodeShortInput].
// If the input is otherwise invalid, a wrapped [ErrDecodeInvalidInput] is returned.
func DecodeCString(b []byte) (string, error) {
	return bsonproto.DecodeCString(b)
}

// Type represents a BSON type.
type Type interface {
	ScalarType | CompositeType
}

// CompositeType represents a BSON composite type (including raw types).
type CompositeType interface {
	*Document | *Array | RawDocument | RawArray
}

// AnyDocument represents a BSON document type (both [*Document] and [RawDocument]).
//
// Note that the Encode and Decode methods could return the receiver itself,
// so care must be taken when results are modified.
type AnyDocument interface {
	Encode() (RawDocument, error)
	Decode() (*Document, error)
	Convert() (*types.Document, error)
}

// AnyArray represents a BSON array type (both [*Array] and [RawArray]).
//
// Note that the Encode and Decode methods could return the receiver itself,
// so care must be taken when results are modified.
type AnyArray interface {
	Encode() (RawArray, error)
	Decode() (*Array, error)
}

// validBSONType checks if v is a valid BSON type (including raw types).
func validBSONType(v any) error {
	switch v := v.(type) {
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
		return lazyerrors.Errorf("invalid BSON type %T", v)
	}

	return nil
}
