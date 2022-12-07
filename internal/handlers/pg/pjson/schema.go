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

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// schema is a document's schema to unmarshal the document correctly.
type schema struct {
	Keys       []string
	Properties map[string]*elem // each elem from $k
}

// elem describes an element of schema.
type elem struct {
	Type    schemaType `json:"t"`            // for each field
	Schema  *schema    `json:"$s,omitempty"` // only for objects
	Items   []*elem    `json:"i,omitempty"`  // only for arrays
	Subtype byte       `json:"s,omitempty"`  // only for binData
	Options string     `json:"o,omitempty"`  // only for regex
}

// schemaType represents possible types in the schema.
type schemaType string

// List of possible types in the schema.
const (
	schemaTypeObject    schemaType = "object"
	schemaTypeArray     schemaType = "array"
	schemaTypeDouble    schemaType = "double"
	schemaTypeString    schemaType = "string"
	schemaTypeBinData   schemaType = "binData"
	schemaTypeObjectID  schemaType = "objectId"
	schemaTypeBool      schemaType = "bool"
	schemaTypeDate      schemaType = "date"
	schemaTypeNull      schemaType = "null"
	schemaTypeRegex     schemaType = "regex"
	schemaTypeInt       schemaType = "int"
	schemaTypeTimestamp schemaType = "timestamp"
	schemaTypeLong      schemaType = "long"
)

// Schemas for scalar types.
var (
	doubleSchema = &elem{
		Type: schemaTypeDouble,
	}
	stringSchema = &elem{
		Type: schemaTypeString,
	}
	binDataSchema = func(subtype byte) *elem {
		return &elem{
			Type:    schemaTypeBinData,
			Subtype: subtype,
		}
	}
	objectIDSchema = &elem{
		Type: schemaTypeObjectID,
	}
	boolSchema = &elem{
		Type: schemaTypeBool,
	}
	dateSchema = &elem{
		Type: schemaTypeDate,
	}
	nullSchema = &elem{
		Type: schemaTypeNull,
	}
	regexSchema = func(options string) *elem {
		return &elem{
			Type:    schemaTypeRegex,
			Options: options,
		}
	}
	intSchema = &elem{
		Type: schemaTypeInt,
	}
	timestampSchema = &elem{
		Type: schemaTypeTimestamp,
	}
	longSchema = &elem{
		Type: schemaTypeLong,
	}
)

func (s *schema) Marshal() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(`{"$k":`)

	keys := s.Keys
	if keys == nil {
		keys = []string{}
	}

	b, err := json.Marshal(keys)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	buf.Write(b)

	for _, key := range keys {
		buf.WriteByte(',')

		if b, err = json.Marshal(key); err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)
		buf.WriteByte(':')

		if b, err = s.Properties[key].Marshal(); err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}

// marshalSchema marshals document's schema.
/* func marshalSchema(td *types.Document) (json.RawMessage, error) {
	var buf bytes.Buffer

	buf.WriteString(`{"$k":`)

	keys := td.Keys()
	if keys == nil {
		keys = []string{}
	}

	b, err := json.Marshal(keys)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	buf.Write(b)

	for _, key := range keys {
		buf.WriteByte(',')

		if b, err = json.Marshal(key); err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)
		buf.WriteByte(':')

		value := must.NotFail(td.Get(key))

		switch val := value.(type) {
		case *types.Document:
			buf.WriteString(`{"t": "object", "$s":`)

			b, err := marshalSchema(val)
			if err != nil {
				return nil, lazyerrors.Error(err)
			}

			buf.Write(b)

			buf.WriteByte('}')

		case *types.Array:
			buf.WriteString(`{"t": "array", "$i":`)

			// todo recursive schema for each element

			buf.WriteByte('}')

		case float64:
			buf.WriteString(`{"t": "double"}`)

		case string:
			buf.WriteString(`{"t": "string"}`)

		case types.Binary:
			buf.WriteString(`{"t": "binData", "s": 0}`) // todo

		case types.ObjectID:
			buf.WriteString(`{"t": "objectId"}`)

		case bool:
			buf.WriteString(`{"t": "bool"}`)

		case time.Time:
			buf.WriteString(`{"t": "date"}`)

		case types.NullType:
			buf.WriteString(`{"t": "null"}`)

		case types.Regex:
			buf.WriteString(`{"t": "regex", "o": ""}`) // todo

		case int32:
			buf.WriteString(`{"t": "int"}`)

		case types.Timestamp:
			buf.WriteString(`{"t": "timestamp"}`)

		case int64:
			buf.WriteString(`{"t": "long"}`)

		default:
			panic(fmt.Sprintf("pjson.marshalSchema: unknown type %[1]T (value %[1]q)", val))
		}

		b, err := Marshal(value)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)
	}

	buf.WriteByte('}')

	return buf.Bytes(), nil
}
*/

// Unmarshal parses the JSON-encoded schema.
func (s *schema) Unmarshal(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)

	var rawMessages map[string]json.RawMessage
	if err := dec.Decode(&rawMessages); err != nil {
		return lazyerrors.Error(err)
	}

	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	b, ok := rawMessages["$k"]
	if !ok {
		return lazyerrors.Errorf("pjson.schema.Unmarshal: missing $k")
	}

	var keys []string
	if err := json.Unmarshal(b, &keys); err != nil {
		return lazyerrors.Error(err)
	}

	s.Keys = keys
	delete(rawMessages, "$k")

	if len(keys) != len(rawMessages) {
		return lazyerrors.Errorf("pjson.schema.Unmarshal: %d elements in $k, %d in total", len(keys), len(rawMessages))
	}

	s.Properties = make(map[string]*elem, len(keys))

	for _, key := range keys {
		b, ok = rawMessages[key]

		if !ok {
			return lazyerrors.Errorf("pjson.schema.Unmarshal: missing key %q", key)
		}

		var e elem
		if err := json.Unmarshal(b, &e); err != nil {
			return lazyerrors.Error(err)
		}

		s.Properties[key] = &e
	}

	return nil
}

func (el *elem) Marshal() ([]byte, error) {
	var b []byte
	var err error

	switch el.Type {
	case schemaTypeObject:
		var buf bytes.Buffer
		buf.WriteString(`{"t": "object", "$s":`)

		if b, err = el.Schema.Marshal(); err != nil {
			return nil, err
		}

		buf.Write(b)

		buf.WriteString(`}`)
		return buf.Bytes(), nil

	case schemaTypeArray:
		var buf bytes.Buffer
		buf.WriteString(`{"t": "array", "i": [`)

		for i, e := range el.Items {
			if b, err = e.Marshal(); err != nil {
				return nil, err
			}

			buf.Write(b)

			if i != len(el.Items)-1 {
				buf.WriteByte(',')
			}
		}

		buf.WriteString(`]}`)
		return buf.Bytes(), nil

	default:
		return json.Marshal(el)
	}
}

func (el *elem) Unmarshal(data []byte) error {
}
