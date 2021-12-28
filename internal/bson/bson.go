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

// Package bson provides converters from/to BSON.
//
// All BSON data types have three representations in FerretDB:
//
//  1. As they are used in handlers implementation.
//  2. As they are used in the wire protocol implementation.
//  3. As they are used to store data in PostgreSQL.
//
// The first representation is provided by types package.
// The second is provided by this package (bson).
// The third is provided by fjson package.
// The reason for that is a separation of concerns: to avoid method names clashes, to simplify type asserts, etc.
package bson

import (
	"bufio"
	"encoding"
	"encoding/json"
)

type bsontype interface {
	bsontype() // seal

	ReadFrom(*bufio.Reader) error
	WriteTo(*bufio.Writer) error
	encoding.BinaryMarshaler
	json.Unmarshaler
	json.Marshaler
}

//go-sumtype:decl bsontype
