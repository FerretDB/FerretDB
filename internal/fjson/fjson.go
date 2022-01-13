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
// Composite/pointer types
//  Document:   {"$k": ["<key 1>", "<key 2>", ...], "<key 1>": <value 1>, "<key 2>": <value 2>, ...}
//  Array:      JSON array
// Scalar/value types
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
	"fmt"
	"io"
	"time"

	"github.com/AlekSi/pointer"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// fjsontype is a type that can be marshaled to/from FJSON.
type fjsontype interface {
	fjsontype() // seal for go-sumtype

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

// fromFJSON converts fjsontype value to matching built-in or types' package value.
func fromFJSON(v fjsontype) any {
	switch v := v.(type) {
	case *document:
		return types.Document(*v)
	case *fjsonArray:
		return pointer.To(types.Array(*v))
	case *double:
		return float64(*v)
	case *fjsonString:
		return string(*v)
	case *fjsonBinary:
		return types.Binary(*v)
	case *fjsonObjectID:
		return types.ObjectID(*v)
	case *fjsonBool:
		return bool(*v)
	case *dateTime:
		return time.Time(*v)
	case nil:
		return nil
	case *fjsonRegex:
		return types.Regex(*v)
	case *fjsonInt32:
		return int32(*v)
	case *fjsonTimestamp:
		return types.Timestamp(*v)
	case *fjsonInt64:
		return int64(*v)
	case *fjsonCString:
		return types.CString(*v)
	}

	panic(fmt.Sprintf("not reached: %T", v)) // for go-sumtype to work
}

// toFJSON converts built-in or types' package value to fjsontype value.
func toFJSON(v any) fjsontype {
	switch v := v.(type) {
	case types.Document:
		return pointer.To(document(v))
	case *types.Array:
		return pointer.To(fjsonArray(*v))
	case float64:
		return pointer.To(double(v))
	case string:
		return pointer.To(fjsonString(v))
	case types.Binary:
		return pointer.To(fjsonBinary(v))
	case types.ObjectID:
		return pointer.To(fjsonObjectID(v))
	case bool:
		return pointer.To(fjsonBool(v))
	case time.Time:
		return pointer.To(dateTime(v))
	case nil:
		return nil
	case types.Regex:
		return pointer.To(fjsonRegex(v))
	case int32:
		return pointer.To(fjsonInt32(v))
	case types.Timestamp:
		return pointer.To(fjsonTimestamp(v))
	case int64:
		return pointer.To(fjsonInt64(v))
	case types.CString:
		return pointer.To(fjsonCString(v))
	}

	panic(fmt.Sprintf("not reached: %T", v)) // for go-sumtype to work
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
			var o double
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$k"] != nil:
			var o document
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$b"] != nil:
			var o fjsonBinary
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$o"] != nil:
			var o fjsonObjectID
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$d"] != nil:
			var o dateTime
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$r"] != nil:
			var o fjsonRegex
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$t"] != nil:
			var o fjsonTimestamp
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$l"] != nil:
			var o fjsonInt64
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$c"] != nil:
			var o fjsonCString
			err = o.UnmarshalJSON(data)
			res = &o
		default:
			err = lazyerrors.Errorf("fjson.Unmarshal: unhandled map %v", v)
		}
	case string:
		res = pointer.To(fjsonString(v))
	case []any:
		var o fjsonArray
		err = o.UnmarshalJSON(data)
		res = &o
	case bool:
		res = pointer.To(fjsonBool(v))
	case nil:
		res = nil
	case float64:
		res = pointer.To(fjsonInt32(v))
	default:
		err = lazyerrors.Errorf("fjson.Unmarshal: unhandled element %[1]T (%[1]v)", v)
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return fromFJSON(res), nil
}

// Marshal encodes given built-in or types' package value into fjson.
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
