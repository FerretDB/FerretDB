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

// Package pjson provides converters from/to pjson (JSON with some extensions) for built-in and `types` types.
//
// See contributing guidelines and documentation for package `types` for details.
//
// # Mapping
//
// Composite types
//
//	*types.Document  {"$k": ["<key 1>", "<key 2>", ...], "<key 1>": <value 1>, "<key 2>": <value 2>, ...}
//	*types.Array     JSON array
//
// Scalar types
//
//	float64          {"$f": JSON number}
//	string           JSON string
//	types.Binary     {"$b": "<base 64 string>", "s": <subtype number>}
//	types.ObjectID   {"$o": "<ObjectID as 24 character hex string"}
//	bool             JSON true / false values
//	time.Time        {"$d": milliseconds since epoch as JSON number}
//	types.NullType   JSON null
//	types.Regex      {"$r": "<string without terminating 0x0>", "o": "<string without terminating 0x0>"}
//	int32            JSON number
//	types.Timestamp  {"$t": "<number as string>"}
//	int64            {"$l": "<number as string>"}
//	TODO Decimal128  {"$n": "<number as string>"}
package pjson

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

// pjsontype is a type that can be marshaled from/to pjson.
type pjsontype interface {
	pjsontype() // seal for go-sumtype

	json.Unmarshaler
	json.Marshaler
}

//go-sumtype:decl pjsontype

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

// fromPJSON converts pjsontype value to matching built-in or types' package value.
func fromPJSON(v pjsontype) any {
	switch v := v.(type) {
	case *documentType:
		return pointer.To(types.Document(*v))
	case *arrayType:
		return pointer.To(types.Array(*v))
	case *doubleType:
		return float64(*v)
	case *stringType:
		return string(*v)
	case *binaryType:
		return types.Binary(*v)
	case *objectIDType:
		return types.ObjectID(*v)
	case *boolType:
		return bool(*v)
	case *dateTimeType:
		return time.Time(*v)
	case *nullType:
		return types.Null
	case *regexType:
		return types.Regex(*v)
	case *int32Type:
		return int32(*v)
	case *timestampType:
		return types.Timestamp(*v)
	case *int64Type:
		return int64(*v)
	}

	panic(fmt.Sprintf("not reached: %T", v)) // for go-sumtype to work
}

// toPJSON converts built-in or types' package value to pjsontype value.
func toPJSON(v any) pjsontype {
	switch v := v.(type) {
	case *types.Document:
		return pointer.To(documentType(*v))
	case *types.Array:
		return pointer.To(arrayType(*v))
	case float64:
		return pointer.To(doubleType(v))
	case string:
		return pointer.To(stringType(v))
	case types.Binary:
		return pointer.To(binaryType(v))
	case types.ObjectID:
		return pointer.To(objectIDType(v))
	case bool:
		return pointer.To(boolType(v))
	case time.Time:
		return pointer.To(dateTimeType(v))
	case types.NullType:
		return pointer.To(nullType(v))
	case types.Regex:
		return pointer.To(regexType(v))
	case int32:
		return pointer.To(int32Type(v))
	case types.Timestamp:
		return pointer.To(timestampType(v))
	case int64:
		return pointer.To(int64Type(v))
	}

	panic(fmt.Sprintf("not reached: %T", v)) // for go-sumtype to work
}

// Unmarshal decodes the given pjson-encoded data.
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

	var res pjsontype

	switch v := v.(type) {
	case map[string]any:
		switch {
		case v["$k"] != nil:
			var o documentType
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$f"] != nil:
			var o doubleType
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$b"] != nil:
			var o binaryType
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$o"] != nil:
			var o objectIDType
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$d"] != nil:
			var o dateTimeType
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$r"] != nil:
			var o regexType
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$t"] != nil:
			var o timestampType
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$l"] != nil:
			var o int64Type
			err = o.UnmarshalJSON(data)
			res = &o
		default:
			err = lazyerrors.Errorf("pjson.Unmarshal: unhandled map %v", v)
		}
	case []any:
		var o arrayType
		err = o.UnmarshalJSON(data)
		res = &o
	case string:
		res = pointer.To(stringType(v))
	case bool:
		res = pointer.To(boolType(v))
	case nil:
		res = new(nullType)
	case float64:
		res = pointer.To(int32Type(v))
	default:
		err = lazyerrors.Errorf("pjson.Unmarshal: unhandled element %[1]T (%[1]v)", v)
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return fromPJSON(res), nil
}

// Marshal encodes given built-in or types' package value into pjson.
func Marshal(v any) ([]byte, error) {
	if v == nil {
		panic("v is nil")
	}

	b, err := toPJSON(v).MarshalJSON()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
}
