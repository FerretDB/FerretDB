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
)

// regexType represents BSON Regular expression type.
type regexType types.Regex

// pjsontype implements pjsontype interface.
func (regex *regexType) pjsontype() {}

// regexJSON is a JSON object representation of the regexType.
type regexJSON struct {
	R string `json:"$r"`
	O string `json:"o"`
}

// UnmarshalJSON implements pjsontype interface.
func (regex *regexType) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o regexJSON
	if err := dec.Decode(&o); err != nil {
		return lazyerrors.Error(err)
	}

	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	*regex = regexType{
		Pattern: o.R,
		Options: o.O,
	}

	return nil
}

// MarshalJSON implements pjsontype interface.
func (regex *regexType) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(regexJSON{
		R: regex.Pattern,
		O: regex.Options,
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// check interfaces
var (
	_ pjsontype = (*regexType)(nil)
)
