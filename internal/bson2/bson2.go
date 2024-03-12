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

// Package bson2 implements encoding and decoding of BSON as defined by https://bsonspec.org/spec.html.
//
// # Types
//
// The following BSON types are supported:
//
//	BSON                Go
//
//	Document/Object     *bson2.Document or bson2.RawDocument
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
	"fmt"
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

// convertToTypes converts valid BSON value of that package to types package type.
//
// Conversions of composite types (including raw types) may cause errors.
// Invalid types cause panics.
func convertToTypes(v any) (any, error) {
	switch v := v.(type) {
	case *Document:
		doc, err := v.Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case RawDocument:
		d, err := v.Decode()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		doc, err := d.Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case *Array:
		arr, err := v.Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return arr, nil

	case RawArray:
		a, err := v.Decode()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		arr, err := a.Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return arr, nil

	case float64:
		return v, nil
	case string:
		return v, nil
	case Binary:
		// Special case to prevent it from being stored as null in sjson.
		// TODO https://github.com/FerretDB/FerretDB/issues/260
		if v.B == nil {
			v.B = []byte{}
		}

		return types.Binary{
			B:       v.B,
			Subtype: types.BinarySubtype(v.Subtype),
		}, nil
	case ObjectID:
		return types.ObjectID(v), nil
	case bool:
		return v, nil
	case time.Time:
		return v, nil
	case NullType:
		return types.Null, nil
	case Regex:
		return types.Regex{
			Pattern: v.Pattern,
			Options: v.Options,
		}, nil
	case int32:
		return v, nil
	case Timestamp:
		return types.Timestamp(v), nil
	case int64:
		return v, nil

	default:
		panic(fmt.Sprintf("invalid BSON type %T", v))
	}
}

// convertFromTypes converts valid types package values to BSON values of that package.
//
// Conversions of composite types may cause errors.
// Invalid types cause panics.
func convertFromTypes(v any) (any, error) {
	switch v := v.(type) {
	case *types.Document:
		doc, err := ConvertDocument(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return doc, nil

	case *types.Array:
		arr, err := ConvertArray(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		return arr, nil

	case float64:
		return v, nil
	case string:
		return v, nil
	case types.Binary:
		return Binary{
			B:       v.B,
			Subtype: BinarySubtype(v.Subtype),
		}, nil
	case types.ObjectID:
		return ObjectID(v), nil
	case bool:
		return v, nil
	case time.Time:
		return v, nil
	case types.NullType:
		return Null, nil
	case types.Regex:
		return Regex{
			Pattern: v.Pattern,
			Options: v.Options,
		}, nil
	case int32:
		return v, nil
	case types.Timestamp:
		return Timestamp(v), nil
	case int64:
		return v, nil

	default:
		panic(fmt.Sprintf("invalid type %T", v))
	}
}
