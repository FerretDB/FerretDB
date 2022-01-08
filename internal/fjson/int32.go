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

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// Int32 represents BSON Int32 data type.
type Int32 int32

// fjsontype implements fjsontype interface.
func (i *Int32) fjsontype() {}

// UnmarshalJSON implements fjsontype interface.
func (i *Int32) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o int32
	if err := dec.Decode(&o); err != nil {
		return lazyerrors.Error(err)
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	*i = Int32(o)
	return nil
}

// MarshalJSON implements fjsontype interface.
func (i *Int32) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(int32(*i))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return res, nil
}

// check interfaces
var (
	_ fjsontype = (*Int32)(nil)
)
