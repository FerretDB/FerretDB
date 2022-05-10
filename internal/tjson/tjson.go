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

// Package tjson provides converters from/to tigris data format for built-in and `types` types.
//
// Tigris Type Mapping
//
// Composite types
//  object { "<key 1>": <value 1>, "<key 2>": <value 2>, ...}
//  array  {<value 1>, <value 2>, <value 3>, ...}
//
// Scalar types
//  types.NullType   null
//  bool             true / false values
//  string           string
//  int32            not implemented
//  float64          number
//  Decimal128       not implemented
//  int64            not implemented
//  types.Binary     binary field
//  types.ObjectID   {$o: string}
//  time.Time        not implemented
//  types.Timestamp  {$t: uint64}
//  types.Regex      {$r: <pattern>, o: <options>}
package tjson

import (
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// checkUnmarshalSupported returns nil if the conversion from Tigris document to Ferret document is supported.
func checkUnmarshalSupported(v any) error {
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case float64, string, bool:
		return nil

	case []any:
		for i := 0; i < len(v); i++ {
			if err := checkUnmarshalSupported(v[i]); err != nil {
				return err
			}
		}
		return nil

	case map[string]any:
		for _, val := range v {
			if err := checkUnmarshalSupported(val); err != nil {
				return err
			}
		}
		return nil

	default:
		return lazyerrors.Errorf("%T not supported", v)
	}
}

// checkMarshalSupported returns nil if the conversion from Ferret document to Tigris document is supported.
func checkMarshalSupported(v any) error {
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case float64, string, bool:
		return nil

	case *types.Array:
		for i := 0; i < v.Len(); i++ {
			if err := checkMarshalSupported(must.NotFail(v.Get(i))); err != nil {
				return err
			}
		}
		return nil

	case *types.Document:
		m := v.Map()
		for _, val := range m {
			if err := checkMarshalSupported(val); err != nil {
				return err
			}
		}
		return nil

	default:
		return lazyerrors.Errorf("%T not supported", v)
	}
}
