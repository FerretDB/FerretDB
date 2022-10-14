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
	"encoding/json"
	"time"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// dateTimeType represents BSON UTC datetime type.
type dateTimeType time.Time

// fjsontype implements fjsontype interface.
func (dt *dateTimeType) fjsontype() {}

// String returns formatted time for debugging.
func (dt *dateTimeType) String() string {
	return time.Time(*dt).Format(time.RFC3339Nano)
}

// dateTimeJSON is a JSON object representation of the dateTimeType.
type dateTimeJSON struct {
	D int64 `json:"$d"`
}

// MarshalJSON implements fjsontype interface.
func (dt *dateTimeType) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(dateTimeJSON{
		D: time.Time(*dt).UnixMilli(),
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return res, nil
}

// check interfaces
var (
	_ fjsontype = (*dateTimeType)(nil)
)
