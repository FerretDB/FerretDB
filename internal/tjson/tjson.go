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
	"encoding/json"
	"time"

	"github.com/tigrisdata/tigris-client-go/driver"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

type timestampJSON struct {
	T uint64 `json:"$t"`
}

type objectIDJSON struct {
	O [12]byte `json:"$o"`
}

type regexJSON struct {
	R string `json:"$r"`
	O string `json:"o"`
}

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
	keys := make([]string, len(v))
	i := 0
	for k := range v {
		keys[i] = k
		i++
	}
	v["$k"] = keys

	docAny, err := fromTJSON(v)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}
	doc := docAny.(*types.Document)
	return doc, nil
}

// fromTJSON converts from Tigris data types to FerretDB data type.
func fromTJSON(v any) (any, error) {
	if v == nil {
		return types.Null, nil
	}
	switch v := v.(type) {
	case map[string]any:
		doc := new(types.Document)
		keys, ok := v["$k"].([]string)
		if !ok {
			return nil, lazyerrors.New("keys in $k absent")
		}
		for _, k := range keys {
			retVal, err := fromTJSON(v[k])
			if err != nil {
				return nil, lazyerrors.Errorf("cannot fromTJSON %s->%v", k, v[k])
			}
			if err = doc.Set(k, retVal); err != nil {
				return nil, lazyerrors.Errorf("cannot fromTJSON: %s", err)
			}
		}
		return doc, nil

	case []any:
		for i := range v {
			if _, err := fromTJSON(v[i]); err != nil {
				return nil, lazyerrors.Errorf("cannot fromTJSON %d->%v", i, v[i])
			}
		}
		arr, err := types.NewArray(v...)
		if err != nil {
			return nil, lazyerrors.Errorf("cannot fromTJSON: %s", err)
		}
		return arr, nil

	case float64:
		return v, nil

	case string:
		var val map[any]any
		err := json.Unmarshal([]byte(v), &val)
		if err != nil {
			return v, nil // it's not a map - return string as is
		}

		if _, ok := val["$o"]; ok {
			var objectIDVal types.ObjectID
			err := json.Unmarshal([]byte(v), &objectIDVal)
			return objectIDVal, err
		}

		if _, ok := val["$r"]; ok {
			var regexVal types.Regex
			err := json.Unmarshal([]byte(v), &regexVal)
			return regexVal, err
		}

		if _, ok := val["$t"]; ok {
			var ts types.Timestamp
			err := json.Unmarshal([]byte(v), &ts)
			return ts, err
		}

	case bool:
		return v, nil

	default:
		return nil, lazyerrors.Errorf("%T not supported", v)
	}

	panic("unreachable code")
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

// toTJSON converts FerretDB types to Tigris data type representation.
func toTJSON(v any) any {
	switch v := v.(type) {
	case *types.Document:
		keys := v.Keys()
		d := make(map[string]any, len(keys))
		for _, k := range keys {
			d[k] = toTJSON(must.NotFail(v.Get(k)))
		}
		d["$k"] = keys
		return d

	case *types.Array:
		a := make([]any, v.Len())
		for i := 0; i <= v.Len(); i++ {
			a[i] = toTJSON(must.NotFail(v.Get(i)))
		}
		return a

	case float64:
		return v

	case string:
		return v

	case types.Binary:
		return v

	case types.ObjectID:
		return objectIDJSON{O: v}

	case bool:
		return v

	case time.Time:
		return v

	case types.NullType:
		return nil

	case types.Regex:
		return regexJSON{R: v.Pattern, O: v.Options}

	case types.Timestamp:
		return timestampJSON{T: uint64(v)}

	default:
		return lazyerrors.Errorf("%T not supported", v)
	}
}

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
