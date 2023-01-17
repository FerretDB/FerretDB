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

// int64Type represents BSON 64-bit integer type.
type int64Type int64

// pjsontype implements pjsontype interface.
func (i *int64Type) pjsontype() {}

// UnmarshalJSON implements json.Unmarshaler interface.
func (i *int64Type) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o int64
	if err := dec.Decode(&o); err != nil {
		return lazyerrors.Error(err)
	}

	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	*i = int64Type(o)

	return nil
}

// MarshalJSON implements pjsontype interface.
func (i *int64Type) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(int64(*i))
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// check interfaces
var (
	_ pjsontype = (*int64Type)(nil)
)
