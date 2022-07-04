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
package tjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

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
		return lazyerrors.Errorf("%d bytes remains in the decoded: %s", dr.Len(), b)
	}

	if l := r.Len(); l != 0 {
		b, _ := io.ReadAll(r)
		return lazyerrors.Errorf("%d bytes remains in the reader: %s", l, b)
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
	// case *dateTimeType:
	// 	return time.Time(*v)
	// case *nullType:
	// 	return types.Null
	// case *regexType:
	// 	return types.Regex(*v)
	case *int32Type:
		return int32(*v)
	// case *timestampType:
	// 	return types.Timestamp(*v)
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
	// case time.Time:
	// 	return pointer.To(dateTimeType(v))
	// case types.NullType:
	// 	return pointer.To(nullType(v))
	// case types.Regex:
	// 	return pointer.To(regexType(v))
	case int32:
		return pointer.To(int32Type(v))
	// case types.Timestamp:
	// 	return pointer.To(timestampType(v))
	case int64:
		return pointer.To(int64Type(v))
	}

	panic(fmt.Sprintf("not reached: %T", v)) // for go-sumtype to work
}

// Unmarshal decodes the given tjson-encoded data.
func Unmarshal(data []byte, schema *Schema) (any, error) {
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
		case UUID, DateTime:
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
		case v["$b"] != nil:
			var o binaryType
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

// ObjectID returns object ID as expected by Tigris filters.
//
// TODO Remove that function if possible. https://github.com/FerretDB/FerretDB/issues/683
func ObjectID(id types.ObjectID) []byte {
	return id[:]
}
