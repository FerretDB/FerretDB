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
)

// binaryType represents BSON Binary data type.
type binaryType types.Binary

// tjsontype implements tjsontype interface.
func (bin *binaryType) tjsontype() {}

type binaryJSON struct {
	B []byte `json:"$b"`
	S byte   `json:"s"`
}

var binarySchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"$b": map[string]any{"type": "string", "format": "byte"},   // binary data
		"s":  map[string]any{"type": "integer", "format": "int32"}, // binary subtype
	},
}

// Unmarshal build-in to tigris.
func (bin *binaryType) Unmarshal(_ map[string]any) ([]byte, error) {
	res, err := json.Marshal(binaryJSON{
		B: bin.B,
		S: byte(bin.Subtype),
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Marshal tigris to build-in.
func (bin *binaryType) Marshal(data []byte, _ map[string]any) error {
	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o binaryJSON
	err := dec.Decode(&o)
	if err != nil {
		return err
	}
	if err = checkConsumed(dec, r); err != nil {
		return err
	}

	bin.B = o.B
	bin.Subtype = types.BinarySubtype(o.S)
	return nil
}

// check interfaces
var (
	_ tjsontype = (*binaryType)(nil)
)
