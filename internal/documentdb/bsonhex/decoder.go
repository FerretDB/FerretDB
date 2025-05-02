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

// Package bsonhex provides functionality to decode BSONHEX type.
package bsonhex

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/FerretDB/wire/wirebson"
)

// bsonHexPrefix is the prefix for BSONHEX type.
var bsonHexPrefix = []byte{'B', 'S', 'O', 'N', 'H', 'E', 'X'}

// Decode converts BSONHEX to the format expected by [wirebson.RawDocument].
// Usage of `::bytea` in PostgreSQL is the preferred approach.
func Decode(src []byte) (wirebson.RawDocument, error) {
	if !bytes.HasPrefix(src, bsonHexPrefix) {
		return nil, fmt.Errorf("expected 'BSONHEX' prefix, got %q", src[:7])
	}

	b := bytes.TrimPrefix(src, bsonHexPrefix)

	dst := make([]byte, hex.DecodedLen(len(b)))

	if _, err := hex.Decode(dst, b); err != nil {
		return nil, err
	}

	return dst, nil
}
