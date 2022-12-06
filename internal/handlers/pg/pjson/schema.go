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
	Keys       []string         `json:"$k"`
	Properties map[string]*elem // each elem from $k
}

// elem describes an element of schema.
type elem struct {
	Type    schemaType // t, for each field
	Schema  *schema    // $s, only for objects
	Items   []*elem    // i, only for arrays
	Subtype byte       // s, only for binData
	Options string     // o, only for regex
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
	b, err := json.Marshal(s)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
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

// unmarshal parses the JSON-encoded schema.
func (s *schema) unmarshal(b []byte) error {
	r := bytes.NewReader(b)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	if err := dec.Decode(s); err != nil {
		return err
	}
	if err := checkConsumed(dec, r); err != nil {
		return err
	}

	// Add $k properties that are necessary for documents.
	s.addDocumentProperties()

	return nil
}

// addDocumentProperties adds missing $k properties to all the schema's documents (top-level and nested).
func (s *schema) addDocumentProperties() {
	for _, prop := range s.Properties {
		switch prop.Type {
		case schemaTypeObject:
			prop.Schema.addDocumentProperties()
		case schemaTypeArray:
			for _, item := range prop.Items {
				if item.Type == schemaTypeObject {
					item.Schema.addDocumentProperties()
				}
			}
		}
	}
}
