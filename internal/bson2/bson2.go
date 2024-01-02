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
	"fmt"
	"time"

	"github.com/cristalhq/bson/bsonproto"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

type (
	ScalarType = bsonproto.ScalarType

	Binary        = bsonproto.Binary
	BinarySubtype = bsonproto.BinarySubtype
	NullType      = bsonproto.NullType
	ObjectID      = bsonproto.ObjectID
	Regex         = bsonproto.Regex
	Timestamp     = bsonproto.Timestamp
)

var (
	ErrDecodeShortInput   = bsonproto.ErrDecodeShortInput
	ErrDecodeInvalidInput = bsonproto.ErrDecodeInvalidInput

	Null = bsonproto.Null
)

// Type represents a BSON type.
type Type interface {
	ScalarType | CompositeType
}

// CompositeType represents a BSON composite type (including raw types).
type CompositeType interface {
	*Document | *Array | RawDocument | RawArray
}

// validBSONType checks if v is a valid BSON type (including raw types).
func validBSONType(v any) bool {
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

func convertToTypes(v any) (any, error) {
	switch v := v.(type) {
	case *Document:
		doc, err := v.Convert()
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		return doc, nil

	case RawDocument:
		d, err := DecodeDocument(v)
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
		a, err := DecodeArray(v)
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
