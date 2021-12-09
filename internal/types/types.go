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

// Package types provides Go types matching BSON types without built-in Go equivalents.
//
// All BSON data types have three representations in FerretDB:
//
//  1. As they are used in handlers implementation.
//  2. As they are used in the wire protocol implementation.
//  3. As they are used to store data in PostgreSQL.
//
// The first representation is provided by this package (types).
// The second and third representations are provided by the bson package.
// The reason for that is a separation of concerns: to avoid method names clashes, to simplify type asserts, etc.
//
// Mapping
//
//  float64:         bson.Double
//  string:          bson.String
//  types.Document:  bson.Document
//  types.Array:     bson.Array
//  types.Binary:    bson.Binary
//  types.ObjectID:  bson.ObjectID
//  bool:            bson.Bool
//  time.Time:       bson.DateTime
//  nil              nil
//  types.Regex:     bson.Regex
//  int32:           bson.Int32
//  types.Timestamp: bson.Timestamp
//  int64:           bson.Int64
//  TODO:            bson.Decimal128
//  (does not exist) bson.CString
package types

type (
	ObjectID [12]byte

	Regex struct {
		Pattern string
		Options string
	}

	Timestamp uint64
)
