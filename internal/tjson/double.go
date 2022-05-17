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
	"errors"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// doubleType represents BSON 64-bit binary floating point type.
type doubleType float64

// tjsontype implements tjsontype interface.
func (d *doubleType) tjsontype() {}

var doubleSchema = map[string]any{"type": "number"}

// Marshal build-in to tigris.
func (d *doubleType) Marshal(_ map[string]any) ([]byte, error) {
	res, err := json.Marshal(*d)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Unmarshal tigris to build-in.
func (d *doubleType) Unmarshal(data []byte, _ map[string]any) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}
	switch string(data) {
	case "Inf", "Infinity":
		return errors.New("json: unsupported value: +Inf")
	case "-Inf", "-Infinity":
		return errors.New("json: unsupported value: -Inf")
	case "NaN":
		return errors.New("json: unsupported value: NaN")
	default:
	}
	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o any
	if err := dec.Decode(&o); err != nil {
		return lazyerrors.Error(err)
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}
	switch f := o.(type) {
	case float64:
		*d = doubleType(f)
	default:
		return lazyerrors.Errorf("fjson.Double.Marshal: unexpected type %[1]T: %[1]v", f)
	}
	return nil
}

// check interfaces
var (
	_ tjsontype = (*doubleType)(nil)
)
