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
func fromTJSON(v tjsontype) any {
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

// toTJSON converts built-in or types' package value to tjsontype value.
func toTJSON(v any) tjsontype {
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

// Unmarshal decodes the given tjson-encoded data.
func Unmarshal(data []byte) (any, error) {
	var err error
	var mp map[string]jsoniter.RawMessage
	if err = jsoniter.Unmarshal(data, &mp); err != nil {
		return nil, err
	}
	var res tjsontype
	if len(mp) == 0 {
		return new(types.Document), nil
	}

	for _, keyVal := range mp {
		switch string(keyVal[:2]) {
		case "$f":
			var o doubleType
			err = o.UnmarshalJSON(data)
			res = &o
		case "$k":
			var o documentType
			err = o.UnmarshalJSON(data)
			res = &o
		case "$b":
			var o binaryType
			err = o.UnmarshalJSON(data)
			res = &o
		case "$o":
			var o objectIDType
			err = o.UnmarshalJSON(data)
			res = &o
		case "$d":
			var o dateTimeType
			err = o.UnmarshalJSON(data)
			res = &o
		case "$r":
			var o regexType
			err = o.UnmarshalJSON(data)
			res = &o
		case "$t":
			var o timestampType
			err = o.UnmarshalJSON(data)
			res = &o
		case "$l":
			var o int64Type
			err = o.UnmarshalJSON(data)
			res = &o
		default:
			err = lazyerrors.Errorf("tjson.Unmarshal: unhandled map %#v", keyVal)
		}
	}
	return fromTJSON(res), err
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
