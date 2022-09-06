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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// timestampType represents BSON Timestamp type.
type timestampType types.Timestamp

// tjsontype implements tjsontype interface.
func (t *timestampType) tjsontype() {}

// String returns formatted time for debugging.
func (t *timestampType) String() string {
	return fmt.Sprint(*t)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *timestampType) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		panic("null data")
	}

	r := bytes.NewReader(data)
	dec := json.NewDecoder(r)

	var o types.Timestamp
	if err := dec.Decode(&o); err != nil {
		return lazyerrors.Error(err)
	}

	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	*t = timestampType(o)

	return nil
}

// MarshalJSON implements tjsontype interface.
func (t *timestampType) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(*t)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return res, nil
}

// check interfaces
var (
	_ tjsontype = (*timestampType)(nil)
)
