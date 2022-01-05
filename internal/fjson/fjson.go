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

// Package fjson provides converters from/to FJSON.
//
// All BSON data types have three representations in FerretDB:
//
//  1. As they are used in handlers implementation (types package).
//  2. As they are used in the wire protocol implementation (bson package).
//  3. As they are used to store data in PostgreSQL (fjson package).
//
// The reason for that is a separation of concerns: to avoid method names clashes, to simplify type asserts, etc.
//
// JSON mapping for storage
//
// Composite types
//  Document:   {"$k": ["<key 1>", "<key 2>", ...], "<key 1>": <value 1>, "<key 2>": <value 2>, ...}
//  Array:      JSON array
// Value types
//  Double:     {"$f": JSON number} or {"$f": "Infinity|-Infinity|NaN"}
//  String:     JSON string
//  Binary:     {"$b": "<base 64 string>", "s": <subtype number>}
//  ObjectID:   {"$o": "<ObjectID as 24 character hex string"}
//  Bool:       JSON true / false values
//  DateTime:   {"$d": milliseconds since epoch as JSON number}
//  nil:        JSON null
//  Regex:      {"$r": "<string without terminating 0x0>", "o": "<string without terminating 0x0>"}
//  Int32:      JSON number
//  Timestamp:  {"$t": "<number as string>"}
//  Int64:      {"$l": "<number as string>"}
//  Decimal128: {"$n": "<number as string>"}
//  CString:    {"$c": "<string without terminating 0x0>"}
package fjson

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/AlekSi/pointer"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

type fjsontype interface {
	fjsontype() // seal

	json.Unmarshaler
	json.Marshaler
}

//go-sumtype:decl fjsontype

// checkConsumed returns error if decoder or reader have buffered or unread data.
func checkConsumed(dec *json.Decoder, r *bytes.Reader) error {
	if dr := dec.Buffered().(*bytes.Reader); dr.Len() != 0 {
		b, _ := io.ReadAll(dr)
		return lazyerrors.Errorf("%d bytes remains in the decoded: %s", dr.Len(), b)
	}

	if l := r.Len(); l != 0 {
		b, _ := io.ReadAll(r)
		return lazyerrors.Errorf("%d bytes remains in the reader: %s", l, b)
	}

	return nil
}

func fromFJSON(v fjsontype) any {
	switch v := v.(type) {
	case *Document:
		return types.Document(*v)
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

func toFJSON(v any) fjsontype {
	switch v := v.(type) {
	case types.Document:
		return pointer.To(Document(v))
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

// Unmarshal decodes the given fjson-encoded data.
func Unmarshal(data []byte) (any, error) {
	var v any
	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	err := dec.Decode(&v)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	if err := checkConsumed(dec, r); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var res fjsontype
	switch v := v.(type) {
	case map[string]any:
		switch {
		case v["$f"] != nil:
			var o Double
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$k"] != nil:
			var o Document
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$b"] != nil:
			var o Binary
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$o"] != nil:
			var o ObjectID
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$d"] != nil:
			var o DateTime
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$r"] != nil:
			var o Regex
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$t"] != nil:
			var o Timestamp
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$l"] != nil:
			var o Int64
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$c"] != nil:
			var o CString
			err = o.UnmarshalJSON(data)
			res = &o
		default:
			err = lazyerrors.Errorf("fjson.Unmarshal: unhandled map %v", v)
		}
	case string:
		res = pointer.To(String(v))
	case []any:
		var o Array
		err = o.UnmarshalJSON(data)
		res = &o
	case bool:
		res = pointer.To(Bool(v))
	case nil:
		res = nil
	case float64:
		res = pointer.To(Int32(v))
	default:
		err = lazyerrors.Errorf("fjson.Unmarshal: unhandled element %[1]T (%[1]v)", v)
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return fromFJSON(res), nil
}

// Marshal encodes given value into fjson.
func Marshal(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	b, err := toFJSON(v).MarshalJSON()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
}
