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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// documentType represents BSON Document type.
type documentType types.Document

// tjsontype implements tjsontype interface.
func (d *documentType) tjsontype() {}

// Marshal built-in to tigris.
func (d *documentType) Marshal(schema map[string]any) ([]byte, error) {
	doc := types.Document(*d)
	var buf bytes.Buffer
	buf.WriteString(`{"$k":["type","properties"]`)
	buf.WriteString(`,"type":"object"`)
	buf.WriteString(`,"properties":`)
	propertiesI, ok := schema["properties"]
	if !ok {
		return nil, lazyerrors.Errorf("tjson.Document.Marshal: missing properties %#v", schema)
	}
	properties, ok := propertiesI.(map[string]any)
	if !ok {
		return nil, lazyerrors.Errorf("tjson.Document.Marshal: wrong properties format %#v", propertiesI)
	}
	buf.WriteString(`{"$k":`)
	keys := doc.Keys()
	if keys == nil {
		keys = []string{}
	}
	b, err := json.Marshal(keys)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	buf.Write(b)

	for _, key := range doc.Keys() { // $k
		value, err := doc.Get(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		fieldSchema, ok := properties[key].(map[string]any)
		if !ok {
			return nil, lazyerrors.Errorf("tjson.Document.Marshal: missing schema for %q: %#v", key, properties)
		}

		if b, err = json.Marshal(key); err != nil {
			return nil, lazyerrors.Error(err)
		}
		buf.WriteByte(',')
		buf.Write(b)
		buf.WriteByte(':')
		b, err := Marshal(value, fieldSchema)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		buf.Write(b)
	}

	buf.WriteByte('}') // properties
	buf.WriteByte('}') // doc
	return buf.Bytes(), nil
}

type docFormat struct {
	Type       string                     `json:"type"`
	Properties map[string]json.RawMessage `json:"properties"`
}

// Unmarshal: tigris to built-in.
func (d *documentType) Unmarshal(data []byte, schema map[string]any) error {
	schemaTypeI, ok := schema["type"]
	if !ok {
		return lazyerrors.Errorf("schema type required")
	}
	if schemaType, ok := schemaTypeI.(string); !ok || schemaType != "object" {
		return lazyerrors.Errorf("wrong schema type for doc")
	}
	propertiesI, ok := schema["properties"]
	if !ok {
		return lazyerrors.Errorf("properties required")
	}
	properties, ok := propertiesI.(map[string]any)
	if !ok {
		return lazyerrors.Errorf("wrong properties format")
	}
	keyOrder := make([]string, len(properties))
	keyOrderI, ok := properties["$k"]
	if ok {
		keyOrder, ok = keyOrderI.([]string)
	}
	if !ok {
		var i int
		for k := range properties {
			keyOrder[i] = k
			i++
		}
	}
	var dataMap docFormat
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return lazyerrors.Error(err)
	}
	td := new(types.Document)
	for i := range keyOrder {
		key := keyOrder[i]
		dataVal, ok := dataMap.Properties[key]
		if !ok {
			continue
		}
		propertySchema, ok := properties[key].(map[string]any)
		if !ok {
			return lazyerrors.Errorf("%s: wrong schema format %[2]T %[2]v", key, propertySchema)
		}
		docVal, err := Unmarshal(dataVal, propertySchema)
		if err != nil {
			return lazyerrors.Error(err)
		}
		if err = td.Set(key, docVal); err != nil {
			return lazyerrors.Error(err)
		}
	}
	*d = documentType(*td)
	return nil
}

// check interfaces
var (
	_ tjsontype = (*documentType)(nil)
)
