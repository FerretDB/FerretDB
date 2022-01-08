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

// Timestamp represents BSON Timestamp data type.
type Timestamp types.Timestamp

// fjsontype implements fjsontype interface.
func (ts *Timestamp) fjsontype() {}

type timestampJSON struct {
	T uint64 `json:"$t,string"`
}

// UnmarshalJSON implements fjsontype interface.
func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o timestampJSON
	if err := dec.Decode(&o); err != nil {
		return lazyerrors.Error(err)
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	*ts = Timestamp(o.T)
	return nil
}

// MarshalJSON implements fjsontype interface.
func (ts *Timestamp) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(timestampJSON{
		T: uint64(*ts),
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return res, nil
}

// check interfaces
var (
	_ fjsontype = (*Timestamp)(nil)
)
