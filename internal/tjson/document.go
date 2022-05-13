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
	jsoniter "github.com/json-iterator/go"
)

// documentType represents BSON Document type.
type documentType types.Document

// tjsontype implements tjsontype interface.
func (dt *documentType) tjsontype() {}

// Unmarshal tigris document to build-in
func (dt *documentType) Unmarshal(doc []byte, schema map[string]any) error {
	var obj map[string][]byte
	if err := jsoniter.Unmarshal(doc, &obj); err != nil {
		return err
	}
	td := new(types.Document)
	for key, val := range obj {
		fieldSchema, ok := schema[key].(map[string]any)
		if !ok {
			return lazyerrors.Errorf("tjson.Document.Unmarshal: missing schema for %q", key)
		}
		v, err := Unmarshal(val, fieldSchema)
		if err != nil {
			return lazyerrors.Error(err)
		}
		if err = td.Set(key, v); err != nil {
			return lazyerrors.Error(err)
		}
	}
	*dt = documentType(*td)
	return nil
}

// Marshal: build-in and schema to tigris doc
func (d *documentType) Marshal(schema map[string]any) ([]byte, error) {
	doc := types.Document(*d)

	var buf bytes.Buffer
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

	for _, key := range doc.Keys() {
		value, err := doc.Get(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		fieldSchema, ok := schema[key].(map[string]any)
		if !ok {
			return nil, lazyerrors.Errorf("tjson.Document.Marshal: missing schema for %q", key)
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

	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// check interfaces
var (
	_ tjsontype = (*documentType)(nil)
)
