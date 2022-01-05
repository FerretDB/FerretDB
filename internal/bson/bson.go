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

// Package bson provides converters from/to BSON.
//
// All BSON data types have three representations in FerretDB:
//
//  1. As they are used in handlers implementation (types package).
//  2. As they are used in the wire protocol implementation (bson package).
//  3. As they are used to store data in PostgreSQL (fjson package).
//
// The reason for that is a separation of concerns: to avoid method names clashes, to simplify type asserts, etc.
package bson

import (
	"bufio"
	"encoding"
	"encoding/json"
	"time"

	"github.com/AlekSi/pointer"

	"github.com/FerretDB/FerretDB/internal/types"
)

type bsontype interface {
	bsontype() // seal

	ReadFrom(*bufio.Reader) error
	WriteTo(*bufio.Writer) error
	encoding.BinaryMarshaler
	json.Unmarshaler
	json.Marshaler
}

//go-sumtype:decl bsontype

func fromBSON(v bsontype) any {
	switch v := v.(type) {
	case *Document:
		return types.MustConvertDocument(v)
	case *Array:
		return pointer.To(types.Array(*v))
	case *Double:
		return float64(*v)
	case *String:
		return string(*v)
	case *Binary:
		return types.Binary(*v)
	case *ObjectID:
		return types.ObjectID(*v)
	case *Bool:
		return bool(*v)
	case *DateTime:
		return time.Time(*v)
	case nil:
		return nil
	case *Regex:
		return types.Regex(*v)
	case *Int32:
		return int32(*v)
	case *Timestamp:
		return types.Timestamp(*v)
	case *Int64:
		return int64(*v)
	case *CString:
		return types.CString(*v)
	}

	panic("not reached") // for go-sumtype to work
}

//nolint:deadcode // remove later if it is not needed
func toBSON(v any) bsontype {
	switch v := v.(type) {
	case types.Document:
		return MustConvertDocument(v)
	case *types.Array:
		return pointer.To(Array(*v))
	case float64:
		return pointer.To(Double(v))
	case string:
		return pointer.To(String(v))
	case types.Binary:
		return pointer.To(Binary(v))
	case types.ObjectID:
		return pointer.To(ObjectID(v))
	case bool:
		return pointer.To(Bool(v))
	case time.Time:
		return pointer.To(DateTime(v))
	case nil:
		return nil
	case types.Regex:
		return pointer.To(Regex(v))
	case int32:
		return pointer.To(Int32(v))
	case types.Timestamp:
		return pointer.To(Timestamp(v))
	case int64:
		return pointer.To(Int64(v))
	case types.CString:
		return pointer.To(CString(v))
	}

	panic("not reached")
}
