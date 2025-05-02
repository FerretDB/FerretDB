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

package bsonhex

import (
	"encoding/hex"
	"github.com/FerretDB/wire/wirebson"
)

// Decode converts BSONHEX to the format expected by [wirebson.RawDocument].
// stripping the first 7 bytes (BSONHEX) and decoding the rest.
// Use this when `::bytea` usage in PostgreSQL is not possible such as procedure output.
func Decode(src []byte) (wirebson.RawDocument, error) {
	dst := make([]byte, hex.DecodedLen(len(src)-7))

	if _, err := hex.Decode(dst, src[7:]); err != nil {
		return nil, err
	}

	return dst, nil
}
