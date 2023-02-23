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

package pjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AlekSi/pointer"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// schema describes document/object schema needed to unmarshal pjson document.
type schema struct {
	Properties map[string]*elem `json:"p"`  // document's properties
	Keys       []string         `json:"$k"` // to preserve properties' order
}

// elem describes an element of schema.
type elem struct {
	Type    elemType             `json:"t"`            // for each field
	Schema  *schema              `json:"$s,omitempty"` // only for objects
	Options *string              `json:"o,omitempty"`  // only for regex
	Items   []*elem              `json:"i,omitempty"`  // only for arrays
	Subtype *types.BinarySubtype `json:"s,omitempty"`  // only for binData
}

// elemType represents possible types of schema elements.
type elemType string

// List of possible types in the schema elements.
const (
	elemTypeObject    elemType = "object"
	elemTypeArray     elemType = "array"
	elemTypeDouble    elemType = "double"
	elemTypeString    elemType = "string"
	elemTypeBinData   elemType = "binData"
	elemTypeObjectID  elemType = "objectId"
	elemTypeBool      elemType = "bool"
	elemTypeDate      elemType = "date"
	elemTypeNull      elemType = "null"
	elemTypeRegex     elemType = "regex"
	elemTypeInt       elemType = "int"
	elemTypeTimestamp elemType = "timestamp"
	elemTypeLong      elemType = "long"
)

func GetTypeOfValue(v any) string {
	switch v.(type) {
	case *types.Document:
		return string(elemTypeObject)
	case *types.Array:
		return string(elemTypeArray)
	case float64:
		return string(elemTypeDouble)
	case string:
		return string(elemTypeString)
	case types.Binary:
		return string(elemTypeBinData)
	case types.ObjectID:
		return string(elemTypeObjectID)
	case bool:
		return string(elemTypeBool)
	case time.Time:
		return string(elemTypeDate)
	case types.NullType:
		return string(elemTypeNull)
	case types.Regex:
		return string(elemTypeRegex)
	case int32:
		return string(elemTypeInt)
	case types.Timestamp:
		return string(elemTypeTimestamp)
	case int64:
		return string(elemTypeLong)
	}

	panic(fmt.Sprintf("Unexpected type: %T", v))
}

// Schemas for scalar types.
var (
	doubleSchema = &elem{
		Type: elemTypeDouble,
	}
	stringSchema = &elem{
		Type: elemTypeString,
	}
	binDataSchema = func(subtype types.BinarySubtype) *elem {
		return &elem{
			Type:    elemTypeBinData,
			Subtype: pointer.To(subtype),
		}
	}
	objectIDSchema = &elem{
		Type: elemTypeObjectID,
	}
	boolSchema = &elem{
		Type: elemTypeBool,
	}
	dateSchema = &elem{
		Type: elemTypeDate,
	}
	nullSchema = &elem{
		Type: elemTypeNull,
	}
	regexSchema = func(options string) *elem {
		return &elem{
			Type:    elemTypeRegex,
			Options: pointer.To(options),
		}
	}
	intSchema = &elem{
		Type: elemTypeInt,
	}
	timestampSchema = &elem{
		Type: elemTypeTimestamp,
	}
	longSchema = &elem{
		Type: elemTypeLong,
	}
)

// marshalSchemaForDoc makes schema for the given document based on its data.
// The result is json encoded schema that can be used to unmarshal the given document.
func marshalSchemaForDoc(td *types.Document) ([]byte, error) {
	var buf bytes.Buffer

	keys := td.Keys()
	if len(keys) == 0 {
		return []byte("{}"), nil
	}

	buf.WriteString(`{"p":{`)

	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}

		var b []byte
		var err error

		if b, err = json.Marshal(key); err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)
		buf.WriteByte(':')

		b, err = marshalElemForSingleValue(must.NotFail(td.Get(key)))
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)
	}

	buf.WriteString(`},"$k":`)

	b, err := json.Marshal(keys)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	buf.Write(b)
	buf.WriteByte('}')

	return buf.Bytes(), nil
}

// marshalElemForSingleValue makes elem for the given element based on its data.
// The result is json encoded elem that can be used to unmarshal the given value.
func marshalElemForSingleValue(value any) ([]byte, error) {
	var buf bytes.Buffer

	switch val := value.(type) {
	case *types.Document:
		buf.WriteString(`{"t":"object"`)

		buf.WriteString(`,"$s":`)

		b, err := marshalSchemaForDoc(val)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)

		buf.WriteByte('}')

	case *types.Array:
		buf.WriteString(`{"t":"array"`)

		buf.WriteString(`,"i":[`)

		for i := 0; i < val.Len(); i++ {
			if i > 0 {
				buf.WriteByte(',')
			}

			b, err := marshalElemForSingleValue(must.NotFail(val.Get(i)))
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			buf.Write(b)
		}

		buf.WriteByte(']')

		buf.WriteByte('}')

	case float64:
		buf.WriteString(`{"t":"double"}`)

	case string:
		buf.WriteString(`{"t":"string"}`)

	case types.Binary:
		subtype, err := json.Marshal(val.Subtype)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.WriteString(`{"t":"binData","s":`)
		buf.Write(subtype)
		buf.WriteString(`}`)

	case types.ObjectID:
		buf.WriteString(`{"t":"objectId"}`)

	case bool:
		buf.WriteString(`{"t":"bool"}`)

	case time.Time:
		buf.WriteString(`{"t":"date"}`)

	case types.NullType:
		buf.WriteString(`{"t":"null"}`)

	case types.Regex:
		options, err := json.Marshal(val.Options)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.WriteString(`{"t":"regex","o":`)
		buf.Write(options)
		buf.WriteString(`}`)

	case int32:
		buf.WriteString(`{"t":"int"}`)

	case types.Timestamp:
		buf.WriteString(`{"t":"timestamp"}`)

	case int64:
		buf.WriteString(`{"t":"long"}`)

	default:
		panic(fmt.Sprintf("pjson.marshalElemForSingleValue: unknown type %[1]T (value %[1]q)", val))
	}

	return buf.Bytes(), nil
}
