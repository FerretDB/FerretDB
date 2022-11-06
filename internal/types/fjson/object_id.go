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
	"encoding/hex"
	"encoding/json"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// objectIDType represents BSON ObjectId type.
type objectIDType types.ObjectID

// fjsontype implements fjsontype interface.
func (obj *objectIDType) fjsontype() {}

// objectIDJSON is a JSON object representation of the objectIDType.
type objectIDJSON struct {
	O string `json:"$o"`
}

// MarshalJSON implements fjsontype interface.
func (obj *objectIDType) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(objectIDJSON{
		O: hex.EncodeToString(obj[:]),
	})
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	return res, nil
}

// check interfaces
var (
	_ fjsontype = (*objectIDType)(nil)
)
