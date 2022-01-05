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

// CString represents BSON CString data type.
type CString string

func (cstr *CString) fjsontype() {}

type cstringJSON struct {
	CString string `json:"$c"`
}

// UnmarshalJSON implements fjsontype interface.
func (cstr *CString) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o cstringJSON
	if err := dec.Decode(&o); err != nil {
		return err
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Errorf("fjson.CString.UnmarshalJSON: %s", err)
	}

	*cstr = CString(o.CString)
	return nil
}

// MarshalJSON implements fjsontype interface.
func (cstr CString) MarshalJSON() ([]byte, error) {
	return json.Marshal(cstringJSON{
		CString: string(cstr),
	})
}

// check interfaces
var (
	_ fjsontype = (*CString)(nil)
)
