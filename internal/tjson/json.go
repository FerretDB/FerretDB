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
	jsoniter "github.com/json-iterator/go"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// tjsontype is a type that can be marshaled to/from FJSON.
type tjsontype interface {
	tjsontype() // seal for go-sumtype

	json.Unmarshaler
	json.Marshaler
}

//go-sumtype:decl tjsontype

// fromTJSON converts tjsontype value to matching built-in or types' package value.
func fromTJSON(v tjsontype) (any, error) {
	switch v := v.(type) {
	case *documentType:
		return pointer.To(types.Document(*v)), nil
	// case *arrayType:
	case *doubleType:
		return float64(*v), nil
	case *stringType:
		return string(*v), nil
	case *binaryType:
		return types.Binary(*v), nil
	case *objectIDType:
		return types.ObjectID(*v), nil
	case *boolType:
		return bool(*v), nil
	case *dateTimeType:
		return time.Time(*v), nil
	case *nullType:
		return types.Null, nil
	case *regexType:
		return types.Regex(*v), nil
	// case *int32Type:
	case *timestampType:
		return types.Timestamp(*v), nil
	default:
		return nil, common.NewErrorMsg(
			common.ErrNotImplemented,
			"int64 not supported yet",
		)
	}

	panic(fmt.Sprintf("not reached: %T", v)) // for go-sumtype to work
}

// toTJSON converts built-in or types' package value to tjsontype value.
func toTJSON(v any) tjsontype {
	switch v := v.(type) {
	case *types.Document:
		return pointer.To(documentType(*v))

	// case *types.Array:
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
		// case int32:

	case types.Timestamp:
		return pointer.To(timestampType(v))
		// case int64:

	}

	panic(fmt.Sprintf("not reached: %T", v)) // for go-sumtype to work
}

// Unmarshal decodes the tjson to build-in.
func Unmarshal(data []byte) (any, error) {
	var err error
	var mp map[string][]byte
	if err = jsoniter.Unmarshal(data, &mp); err != nil {
		return nil, err
	}
	var res tjsontype
	if len(mp) == 0 {
		return new(types.Document), nil
	}
	for _, v := range mp {
		res, err := unmarshalField(v)
		if err != nil {
			return nil, err
		}
	}
}

func unmarshalField(v []byte, schema map[string]any) (tjsontype, error) {

	fieldType, ok := schema["type"]
	if !ok {
		return nil, lazyerrors.Errorf("canont find field type")
	}
	var err error
	var res tjsontype
	switch fieldType {
	case "object":
		var obj map[string]any
		if err = jsoniter.Unmarshal(v, &obj); err != nil {
			return nil, err
		}
		if _, ok := obj["$b"]; ok {
			var o binaryType
			err = o.UnmarshalJSON(v)
			res = &o
		}
		if _, ok := obj["$o"]; ok {
			var o objectIDType
			err = o.UnmarshalJSON(v)
			res = &o
		}
		if _, ok := obj["$r"]; ok {
			var o regexType
			err = o.UnmarshalJSON(v)
			res = &o
		}
		if _, ok := obj["$t"]; ok {
			var o timestampType
			err = o.UnmarshalJSON(v)
			res = &o
		}

		var o documentType
		err = o.UnmarshalJSON(v)
		res = &o
	case "array":
		err = common.NewErrorMsg(common.ErrNotImplemented, "arrays not supported yet")

	case "boolean":
		var o boolType
		err = o.UnmarshalJSON(v)
		res = &o

	case "string":
		var obj map[string]any
		if err = jsoniter.Unmarshal(v, &obj); err != nil {
			return nil, err
		}
		if format, ok := obj["format"]; ok {
			if format == "date-time" {
				var o dateTimeType
				err = o.UnmarshalJSON(v)
				res = &o
			}
		}

	case "$k":

	default:
		err = lazyerrors.Errorf("tjson.Unmarshal: unhandled map %#v", v)
	}
	return res, nil
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
