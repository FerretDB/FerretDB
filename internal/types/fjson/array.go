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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
)

// arrayType represents BSON Array type.
type arrayType types.Array

// fjsontype implements fjsontype interface.
func (a *arrayType) fjsontype() {}

// MarshalJSON implements fjsontype interface.
func (a *arrayType) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('[')

	ta := types.Array(*a)
	l := ta.Len()
	for i := 0; i < l; i++ {
		if i != 0 {
			buf.WriteByte(',')
		}

		el, err := ta.Get(i)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		b, err := Marshal(el)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}

		buf.Write(b)
	}

	buf.WriteByte(']')
	return buf.Bytes(), nil
}

// check interfaces
var (
	_ fjsontype = (*arrayType)(nil)
)
