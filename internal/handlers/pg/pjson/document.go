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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// documentType represents BSON Document type.
type documentType types.Document

// pjsontype implements pjsontype interface.
func (doc *documentType) pjsontype() {}

// UnmarshalJSONWithSchema unmarshals the JSON data with the given schema.
func (doc *documentType) UnmarshalJSONWithSchema(data []byte, sch *schema) error {
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

	if len(rawMessages) == 0 {
		*doc = documentType{}
		return nil
	}

	if sch == nil {
		return lazyerrors.Errorf("document schema is nil for non-empty document")
	}

	if len(sch.Keys) != len(rawMessages) {
		return lazyerrors.Errorf("pjson.documentType.UnmarshalJSON: %d elements in $k in the schema, %d in the document",
			len(sch.Keys), len(rawMessages),
		)
	}

	td := must.NotFail(types.NewDocument())

	for _, key := range sch.Keys {
		b, ok := rawMessages[key]

		if !ok {
			return lazyerrors.Errorf("pjson.documentType.UnmarshalJSON: missing key %q", key)
		}

		v, err := unmarshalSingleValue(b, sch.Properties[key])
		if err != nil {
			return lazyerrors.Error(err)
		}

		td.Set(key, v)
	}

	*doc = documentType(*td)

	return nil
}

// MarshalJSON implements pjsontype interface.
func (doc *documentType) MarshalJSON() ([]byte, error) {
	td := types.Document(*doc)

	var buf bytes.Buffer

	buf.WriteString(`{`)

	keys := td.Keys()

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

		value, err := td.Get(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		b, err = MarshalSingleValue(value)
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
	_ pjsontype = (*documentType)(nil)
)
