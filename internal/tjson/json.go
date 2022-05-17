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

// tjsontype is a type that can be marshaled to/from TJSON.
type tjsontype interface {
	tjsontype() // seal for go-sumtype

	Unmarshal(tigrisDoc []byte, schema map[string]any) error // tigris to build-in
	Marshal(schema map[string]any) ([]byte, error)           // build-in to tigris.
}

//go-sumtype:decl tjsontype

// fromTJSON converts tjsontype value to matching built-in or types' package value.
func fromTJSON(v tjsontype) any {
	switch v := v.(type) {
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
	case *timestampType:
		return types.Timestamp(*v)
	}

	panic(fmt.Sprintf("not reached: %T", v))
}

// toTJSON converts built-in or types' package value to tjsontype value.
func toTJSON(v any) tjsontype {
	switch v := v.(type) {
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
	case types.Timestamp:
		return pointer.To(timestampType(v))
	}

	panic(fmt.Sprintf("not reached: %T", v))
}

// Marshal build-in to tigris.
func Marshal(v any, schema map[string]any) ([]byte, error) {
	tv := toTJSON(v)

	switch v := tv.(type) {
	case *stringType:
		s := stringType(*v)
		return s.Marshal(schema)
	case *binaryType:
		b := binaryType(*v)
		return b.Marshal(schema)
	case *objectIDType:
		o := objectIDType(*v)
		return o.Marshal(schema)
	case *boolType:
		b := boolType(*v)
		return b.Marshal(schema)
	case *dateTimeType:
		t := dateTimeType(*v)
		return t.Marshal(schema)
	case *nullType:
		n := nullType(*v)
		return n.Marshal(schema)
	case *regexType:
		r := regexType(*v)
		return r.Marshal(schema)
	case *timestampType:
		t := timestampType(*v)
		return t.Marshal(schema)
	}
	return nil, lazyerrors.Errorf("%T is not supported", v)
}

// Unmarshal tigris to build-in.
func Unmarshal(v []byte, schema map[string]any) (any, error) {
	fieldType, ok := schema["type"]
	if !ok {
		return nil, lazyerrors.Errorf("canont find field type")
	}

	var err error
	var res tjsontype
	switch fieldType {
	case "object":
		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			return nil, lazyerrors.Errorf("tjson.Document.Unmarshal: missing properties in schema")
		}
		if _, ok := properties["$b"]; ok {
			var o binaryType
			err = o.Unmarshal(v, schema)
			res = &o
			break
		}
		if _, ok := properties["$o"]; ok {
			var o objectIDType
			err = o.Unmarshal(v, schema)
			res = &o
			break
		}
		if _, ok := properties["$r"]; ok {
			var o regexType
			err = o.Unmarshal(v, schema)
			res = &o
			break
		}
		if _, ok := properties["$t"]; ok {
			var o timestampType
			err = o.Unmarshal(v, schema)
			res = &o
			break
		}
	case "array":
		err = lazyerrors.Errorf("arrays not supported yet")

	case "boolean":
		var o boolType
		err = o.Unmarshal(v, schema)
		res = &o

	case "string":
		if format, ok := schema["format"]; ok {
			if format == "date-time" {
				var o dateTimeType
				err = o.Unmarshal(v, schema)
				res = &o
				break
			}
		}

		var o stringType
		err = o.Unmarshal(v, schema)
		res = &o

	default:
		err = lazyerrors.Errorf("tjson.Marshal: unhandled map %#v", v)
	}
	return fromTJSON(res), err
}

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
