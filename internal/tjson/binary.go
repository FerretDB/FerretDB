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

// binaryType represents BSON Binary data type.
type binaryType types.Binary

// tjsontype implements tjsontype interface.
func (bin *binaryType) tjsontype() {}

// binaryJSON is a JSON object representation of the binaryType.
type binaryJSON struct {
	B []byte `json:"$b"`
	S byte   `json:"s"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (bin *binaryType) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o binaryJSON
	err := dec.Decode(&o)
	if err != nil {
		return lazyerrors.Error(err)
	}
	if err = checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	bin.B = o.B
	bin.Subtype = types.BinarySubtype(o.S)
	return nil
}

// MarshalJSON implements tjsontype interface.
func (bin *binaryType) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(binaryJSON{
		B: bin.B,
		S: byte(bin.Subtype),
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return res, nil
}

// check interfaces
var (
	_ tjsontype = (*binaryType)(nil)
)
