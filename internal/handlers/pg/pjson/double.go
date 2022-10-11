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
	"math"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// doubleType represents BSON 64-bit binary floating point type.
type doubleType float64

// pjsontype implements pjsontype interface.
func (d *doubleType) pjsontype() {}

// doubleJSON is a JSON object representation of the doubleType.
type doubleJSON struct {
	F any `json:"$f"`
}

// UnmarshalJSON implements pjsontype interface.
func (d *doubleType) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var o doubleJSON
	if err := dec.Decode(&o); err != nil {
		return lazyerrors.Error(err)
	}

	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	switch f := o.F.(type) {
	case float64:
		*d = doubleType(f)
	case string:
		switch f {
		case "-0":
			*d = doubleType(math.Copysign(0, -1))
		case "Infinity":
			*d = doubleType(math.Inf(+1))
		case "-Infinity":
			*d = doubleType(math.Inf(-1))
		case "NaN":
			*d = doubleType(math.NaN())
		default:
			return lazyerrors.Errorf("pjson.Double.UnmarshalJSON: unexpected string %q", f)
		}
	default:
		return lazyerrors.Errorf("pjson.Double.UnmarshalJSON: unexpected type %[1]T: %[1]v", f)
	}

	return nil
}

// MarshalJSON implements pjsontype interface.
func (d *doubleType) MarshalJSON() ([]byte, error) {
	f := float64(*d)

	var o doubleJSON

	switch {
	case f == 0 && math.Signbit(f):
		o.F = "-0"
	case math.IsInf(f, 1):
		o.F = "Infinity"
	case math.IsInf(f, -1):
		o.F = "-Infinity"
	case math.IsNaN(f):
		o.F = "NaN"
	default:
		o.F = f
	}

	res, err := json.Marshal(o)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// check interfaces
var (
	_ pjsontype = (*doubleType)(nil)
)
