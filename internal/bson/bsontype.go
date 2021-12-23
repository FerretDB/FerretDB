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
// The first representation is provided by the types package.
// The second and third representations are provided by this package (bson).
// The reason for that is a separation of concerns: to avoid method names clashes, to simplify type asserts, etc.
//
// JSON mapping for storage
//
// Composite types
//  Document:   {"$k": ["<key 1>", "<key 2>", ...], "<key 1>": <value 1>, "<key 2>": <value 2>, ...}
//  Array:      JSON array
// Value types
//  Double:     {"$f": JSON number} or {"$f": "Infinity|-Infinity|NaN"}
//  String:     JSON string
//  Binary:     {"$b": "<base 64 string>", "s": <subtype number>}
//  ObjectID:   {"$o": "<ObjectID as 24 character hex string"}
//  Bool:       JSON true / false values
//  DateTime:   {"$d": milliseconds since epoch as JSON number}
//  nil:        JSON null
//  Regex:      {"$r": "<string without terminating 0x0>", "o": "<string without terminating 0x0>"}
//  Int32:      JSON number
//  Timestamp:  {"$t": "<number as string>"}
//  Int64:      {"$l": "<number as string>"}
//  Decimal128: {"$n": "<number as string>"}
//  CString:    {"$c": "<string without terminating 0x0>"}
package bson

import (
	"bufio"
	"bytes"
	"encoding"
	"encoding/json"
	"io"

	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
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

func checkConsumed(dec *json.Decoder, r *bytes.Reader) error {
	if dr := dec.Buffered().(*bytes.Reader); dr.Len() != 0 {
		b, _ := io.ReadAll(dr)
		return lazyerrors.Errorf("%d bytes remains in the decoded: %s", dr.Len(), b)
	}

	if l := r.Len(); l != 0 {
		b, _ := io.ReadAll(r)
		return lazyerrors.Errorf("%d bytes remains in the reader: %s", l, b)
	}

	return nil
}
