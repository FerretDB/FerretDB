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
//  1. As they are used in handlers implementation.
//  2. As they are used in the wire protocol implementation.
//  3. As they are used to store data in PostgreSQL.
//
// The first representation is provided by the types package.
// The second and third representations are provided by this package (bson).
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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

type fjsontype interface {
	fjsontype() // seal

	json.Unmarshaler
	json.Marshaler
}

//go-sumtype:decl fjsontype

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

func Unmarshal(data []byte) (any, error) {
	var v any
	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	err := dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	if err := checkConsumed(dec, r); err != nil {
		return nil, lazyerrors.Error(err)
	}

	var res any
	switch v := v.(type) {
	case map[string]any:
		switch {
		case v["$f"] != nil:
			var o Double
			err = o.UnmarshalJSON(data)
			res = float64(o)
		case v["$k"] != nil:
			var o Document
			err = o.UnmarshalJSON(data)
			res = types.Document(o)
		case v["$b"] != nil:
			var o Binary
			err = o.UnmarshalJSON(data)
			res = types.Binary(o)
		case v["$o"] != nil:
			var o ObjectID
			err = o.UnmarshalJSON(data)
			res = types.ObjectID(o)
		case v["$d"] != nil:
			var o DateTime
			err = o.UnmarshalJSON(data)
			res = time.Time(o)
		case v["$r"] != nil:
			var o Regex
			err = o.UnmarshalJSON(data)
			res = types.Regex(o)
		case v["$t"] != nil:
			var o Timestamp
			err = o.UnmarshalJSON(data)
			res = types.Timestamp(o)
		case v["$l"] != nil:
			var o Int64
			err = o.UnmarshalJSON(data)
			res = int64(o)
		default:
			err = lazyerrors.Errorf("unmarshalJSONValue: unhandled map %v", v)
		}
	case string:
		res = v
	case []any:
		var o Array
		err = o.UnmarshalJSON(data)
		ta := types.Array(o)
		res = &ta
	case bool:
		res = v
	case nil:
		res = v
	case float64:
		res = int32(v)
	default:
		err = lazyerrors.Errorf("unmarshalJSONValue: unhandled element %[1]T (%[1]v)", v)
	}

	if err != nil {
		return nil, err
	}

	return res, nil
}

func Marshal(v any) ([]byte, error) {
	var o json.Marshaler
	switch v := v.(type) {
	case types.Document:
		o = Document(v)
	case *types.Array:
		o = Array(*v)
	case float64:
		o = Double(v)
	case string:
		o = String(v)
	case types.Binary:
		o = Binary(v)
	case types.ObjectID:
		o = ObjectID(v)
	case bool:
		o = Bool(v)
	case time.Time:
		o = DateTime(v)
	case nil:
		return []byte("null"), nil
	case types.Regex:
		o = Regex(v)
	case int32:
		o = Int32(v)
	case types.Timestamp:
		o = Timestamp(v)
	case int64:
		o = Int64(v)
	default:
		return nil, lazyerrors.Errorf("marshalJSONValue: unhandled type %T", v)
	}

	b, err := o.MarshalJSON()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
}
