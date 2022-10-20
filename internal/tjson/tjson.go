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

// Package tjson provides converters from/to JSON with JSON Schema for built-in and `types` types.
//
// See contributing guidelines and documentation for package `types` for details.
// See https://docs.tigrisdata.com/overview/schema#data-types for more details about JSON types.
//
// # Mapping
//
// Composite types
//
//	*types.Document       {"$k": ["<key 1>", "<key 2>", ...], "<key 1>": <value 1>, "<key 2>": <value 2>, ...}
//	TODO *types.Array     JSON array
//
// Scalar types
//
//	float64               JSON number (double format)
//	string                JSON string
//	types.Binary          {"$b": "<base 64 string>", "s": <subtype number>}
//	types.ObjectID	      JSON string (byte format, length is 12 bytes)
//	bool                  JSON true|false values
//	time.Time        	  JSON string (date-time RFC3339 format)
//	TODO types.NullType   JSON null
//	types.Regex           {"$r": "<string without terminating 0x0>", "o": "<string without terminating 0x0>"}
//	int32                 JSON number (int32 format)
//	types.Timestamp  	  {"$t": "<number as string>"}
//	int64                 JSON number (int64 format)
//	TODO Decimal128       {"$n": "<number as string>"}
package tjson

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

// tjsontype is a type that can be marshaled from/to JSON with JSON Schema.
type tjsontype interface {
	tjsontype() // seal for go-sumtype

	json.Marshaler
}

//go-sumtype:decl tjsontype

// checkConsumed returns error if decoder or reader have buffered or unread data.
func checkConsumed(dec *json.Decoder, r *bytes.Reader) error {
	if dr := dec.Buffered().(*bytes.Reader); dr.Len() != 0 {
		b, _ := io.ReadAll(dr)

		// Tigris might add \n at the end of a valid document, we consider such situation as valid.
		b = bytes.TrimSpace(b)
		if l := len(b); l != 0 {
			return lazyerrors.Errorf("%[1]d bytes remains in the reader: `%[2]s` (%#02[2]x)", l, b)
		}
	}

	if l := r.Len(); l != 0 {
		b, _ := io.ReadAll(r)
		return lazyerrors.Errorf("%[1]d bytes remains in the reader: `%[2]s` (%#02[2]x)", l, b)
	}

	return nil
}

// fromTJSON converts tjsontype value to matching built-in or types' package value.
func fromTJSON(v tjsontype) any {
	switch v := v.(type) {
	case *documentType:
		return pointer.To(types.Document(*v))
	// case *arrayType:
	// 	return pointer.To(types.Array(*v))
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

// toTJSON converts built-in or types' package value to tjsontype value.
func toTJSON(v any) tjsontype {
	switch v := v.(type) {
	case *types.Document:
		return pointer.To(documentType(*v))
	// case *types.Array:
	// 	return pointer.To(arrayType(*v))
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

// Unmarshal decodes the given tjson-encoded data.
func Unmarshal(data []byte, schema *Schema) (any, error) {
	if bytes.Equal(data, []byte("null")) {
		return fromTJSON(new(nullType)), nil
	}

	var res tjsontype
	var err error
	switch t := schema.Type; t {
	case Integer:
		switch f := schema.Format; f {
		case EmptyFormat, Int64:
			var o int64Type
			err = o.UnmarshalJSON(data)
			res = &o
		case Int32:
			var o int32Type
			err = o.UnmarshalJSON(data)
			res = &o
		case Double, Float:
			fallthrough
		case Byte, UUID, DateTime:
			fallthrough
		default:
			err = lazyerrors.Errorf("tjson.Unmarshal: unhandled format %q for type %q", f, t)
		}
	case Number:
		switch f := schema.Format; f {
		case EmptyFormat, Double:
			var o doubleType
			err = o.UnmarshalJSON(data)
			res = &o
		case Float:
			fallthrough
		case Int64, Int32:
			fallthrough
		case Byte, UUID, DateTime:
			fallthrough
		default:
			err = lazyerrors.Errorf("tjson.Unmarshal: unhandled format %q for type %q", f, t)
		}
	case String:
		switch f := schema.Format; f {
		case EmptyFormat:
			var o stringType
			err = o.UnmarshalJSON(data)
			res = &o
		case Byte:
			var o objectIDType
			err = o.UnmarshalJSON(data)
			res = &o
		case DateTime:
			var o dateTimeType
			err = o.UnmarshalJSON(data)
			res = &o
		case UUID:
			fallthrough
		case Double, Float, Int64, Int32:
			fallthrough
		default:
			err = lazyerrors.Errorf("tjson.Unmarshal: unhandled format %q for type %q", f, t)
		}
	case Boolean:
		var o boolType
		err = o.UnmarshalJSON(data)
		res = &o
	case Array:
		err = lazyerrors.Errorf("tjson.Unmarshal: unhandled type %q", t)
	case Object:
		var v map[string]json.RawMessage
		r := bytes.NewReader(data)
		dec := json.NewDecoder(r)
		if err = dec.Decode(&v); err != nil {
			return nil, lazyerrors.Error(err)
		}
		if err = checkConsumed(dec, r); err != nil {
			return nil, lazyerrors.Error(err)
		}

		switch {
		case v["$k"] != nil:
			var o documentType
			err = o.UnmarshalJSONWithSchema(data, schema)
			res = &o
		case v["$t"] != nil:
			var o timestampType
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$b"] != nil:
			var o binaryType
			err = o.UnmarshalJSON(data)
			res = &o
		case v["$r"] != nil:
			var o regexType
			err = o.UnmarshalJSON(data)
			res = &o
		default:
			err = lazyerrors.Errorf("tjson.Unmarshal: unhandled map %v", v)
		}
	default:
		err = lazyerrors.Errorf("tjson.Unmarshal: unhandled type %q", t)
	}

	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return fromTJSON(res), nil
}

// Marshal encodes given built-in or types' package value into tjson.
func Marshal(v any) ([]byte, error) {
	if v == nil {
		panic("v is nil")
	}

	b, err := toTJSON(v).MarshalJSON()
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
}
