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
	"strconv"
	"strings"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// doubleType represents BSON 64-bit binary floating point type.
type doubleType float64

// tjsontype implements tjsontype interface.
func (d *doubleType) tjsontype() {}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *doubleType) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)

	var o float64
	if err := dec.Decode(&o); err != nil {
		return lazyerrors.Error(err)
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	*d = doubleType(o)
	return nil
}

// MarshalJSON implements tjsontype interface.
func (d *doubleType) MarshalJSON() ([]byte, error) {
	// This will default to precision of 6
	//res := []byte(fmt.Sprintf("%f", float64(*d)))

	// get precision of the value to preserve it
	dStr := strconv.FormatFloat(float64(*d), 'f', -1, 64)
	precision := len(dStr) - strings.Index(dStr, ".") - 1

	res := []byte(fmt.Sprintf("%."+strconv.Itoa(precision)+"f", float64(*d)))

	// This function still does not handle some corner cases like TestDouble/zero where we expect to return "0" and not "0.0"
	return res, nil
}

// check interfaces
var (
	_ tjsontype = (*doubleType)(nil)
)
