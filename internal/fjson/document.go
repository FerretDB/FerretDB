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

package fjson

import (
	"bytes"
	"encoding/json"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// documentType represents BSON Document type.
type documentType types.Document

// fjsontype implements fjsontype interface.
func (doc *documentType) fjsontype() {}

// UnmarshalJSON implements fjsontype interface.
func (doc *documentType) UnmarshalJSON(data []byte) error {
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
		return lazyerrors.Errorf("fjson.Document.UnmarshalJSON: missing $k")
	}

	var keys []string
	if err := json.Unmarshal(b, &keys); err != nil {
		return lazyerrors.Error(err)
	}
	if len(keys)+1 != len(rawMessages) {
		return lazyerrors.Errorf("fjson.Document.UnmarshalJSON: %d elements in $k, %d in total", len(keys), len(rawMessages))
	}

	td := types.MustNewDocument()
	for _, key := range keys {
		b, ok = rawMessages[key]
		if !ok {
			return lazyerrors.Errorf("fjson.Document.UnmarshalJSON: missing key %q", key)
		}
		v, err := Unmarshal(b)
		if err != nil {
			return lazyerrors.Error(err)
		}
		if err = td.Set(key, v); err != nil {
			return lazyerrors.Error(err)
		}
	}

	*doc = documentType(*td)
	return nil
}

// MarshalJSON implements fjsontype interface.
func (doc *documentType) MarshalJSON() ([]byte, error) {
	td := types.Document(*doc)

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

		value, err := td.Get(key)
		if err != nil {
			return nil, lazyerrors.Error(err)
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

// check interfaces
var (
	_ fjsontype = (*documentType)(nil)
)
