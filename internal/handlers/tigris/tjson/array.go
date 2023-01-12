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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// arrayType represents BSON Array type.
type arrayType types.Array

// tjsontype implements tjsontype interface.
func (a *arrayType) tjsontype() {}

// UnmarshalJSONWithSchema unmarshals the JSON data with the given schema.
func (a *arrayType) UnmarshalJSONWithSchema(data []byte, schema *Schema) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	if schema.Type != Array {
		panic(fmt.Sprintf("unexpected type %q", schema.Type))
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)

	var rawMessages []json.RawMessage
	if err := dec.Decode(&rawMessages); err != nil {
		return lazyerrors.Error(err)
	}

	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	if len(rawMessages) > 0 && schema.Items == nil {
		return lazyerrors.Errorf("tjson.arrayType.UnmarshalJSON: array schema is nil for non-empty array")
	}

	ta := types.MakeArray(len(rawMessages))

	for _, el := range rawMessages {
		v, err := Unmarshal(el, schema.Items)
		if err != nil {
			return lazyerrors.Error(err)
		}

		if err = ta.Append(v); err != nil {
			return lazyerrors.Error(err)
		}
	}

	*a = arrayType(*ta)

	return nil
}

// MarshalJSON implements tjsontype interface.
func (a *arrayType) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')

	ta := types.Array(*a)

	for i := 0; i < ta.Len(); i++ {
		if i != 0 {
			buf.WriteByte(',')
		}

		el, err := ta.Get(i)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		b, err := Marshal(el)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)
	}

	buf.WriteByte(']')

	return buf.Bytes(), nil
}

// check interfaces
var (
	_ tjsontype = (*arrayType)(nil)
)
