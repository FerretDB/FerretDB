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
//  int32            integer
//  float64          number - 8 bytes (64-bit IEEE 754-2008 binary floating point)
//  Decimal128       not implemented - number as string
//  int64            not implemented - number as string
//  types.Binary     not implemented - binary field
//  types.ObjectID   not implemented - string
//  time.Time        not implemented - date-time string
//  types.Timestamp  not implemented - <number as string>
//  types.CString    not implemented - <string without terminating 0x0>
//  types.Regex      not implemented - <string without terminating 0x0>
package tjson

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tigrisdata/tigrisdb-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// Unmarshal decodes the given tigris data type *driver.Document to a *types.Document.
func Unmarshal(data *driver.Document) (*types.Document, error) {
	var v map[string]any
	err := json.Unmarshal(*data, &v)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	if err = checkUnmarshalSupported(v); err != nil {
		return nil, err
	}

	pairs := make([]any, 2*len(v))
	var i int
	for k, v := range v {
		pairs[i] = k
		pairs[i+1] = v
		i += 2
	}

	doc, err := types.NewDocument(pairs...)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return doc, nil
}

// Marshal encodes the given *types.Document to a tigris driver.Document.
func Marshal(v *types.Document) (*driver.Document, error) {
	if v == nil {
		panic("v is nil")
	}

	if err := checkMarshalSupported(v); err != nil {
		return nil, err
	}

	doc := toTJSON(v)

	b, err := json.Marshal(doc)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	d := driver.Document(b)
	return &d, nil
}

type timestampJSON struct {
	D uint64 `json:"$d"`
}

type objectIDJSON struct {
	O [12]byte `json:"$o"`
}

type regexJSON struct {
	R string `json:"$r"`
	O string `json:"o"`
}

// toTJSON converts FerretDB types to Tigris data type representation
func toTJSON(v any) any {
	if v == nil {
		return types.Null
	}
	switch v := v.(type) {
	case string:
		return v

	case float64:
		return v

	case types.Binary:
		return v

	case types.ObjectID:
		return objectIDJSON{O: v}

	case time.Time:
		return v

	case types.NullType:
		return nil

	case types.Timestamp:
		return timestampJSON{D: uint64(v)}

	case types.Regex:
		return regexJSON{R: v.Pattern, O: v.Options}

	case *types.Document:
		keys := v.Keys()
		d := make(map[string]any, len(keys))
		for _, k := range keys {
			d[k] = toTJSON(must.NotFail(v.Get(k)))
		}
		return d

	case *types.Array:
		a := make([]any, v.Len())
		for i := 0; i <= v.Len(); i++ {
			a[i] = toTJSON(must.NotFail(v.Get(i)))
		}
		return a

	default:
		panic(fmt.Sprintf("not reached: %T", v))
	}
}

// checkUnmarshalSupported returns nil if the conversion from Tigris document to Ferret document is supported.
func checkUnmarshalSupported(v any) error {
	if v == nil {
		return nil
	}

	switch v := v.(type) {
	case bool, string, float64:
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
	case bool, string, float64:
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
