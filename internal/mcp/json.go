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

package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/FerretDB/wire/wirebson"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/FerretDB/FerretDB/v2/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/v2/internal/util/must"
)

// prepareDocument creates a new bson document from JSON field values pairs,
// including json.RawMessage value.
//
// If any of pair values is nil it's ignored.
func prepareDocument(pairs ...any) (*wirebson.Document, error) {
	l := len(pairs)

	if l%2 != 0 {
		return nil, lazyerrors.Errorf("invalid number of arguments: %d", l)
	}

	docPairs := make([]any, 0, l)

	for i := 0; i < l; i += 2 {
		var err error

		key := pairs[i]
		v := pairs[i+1]

		switch val := v.(type) {
		// json.RawMessage is the non-pointer exception.
		// Other non-pointer types don't need special handling.
		case json.RawMessage:
			v, err = unmarshalExtJSON(&val)
			if err != nil {
				return nil, err
			}

		case *json.RawMessage:
			if val == nil {
				continue
			}

			v, err = unmarshalExtJSON(val)
			if err != nil {
				return nil, err
			}
		case *float32:
			if val == nil {
				continue
			}

			v = float64(*val)
		case *bool:
			if val == nil {
				continue
			}

			v = *val
		}

		if v == nil {
			continue
		}

		docPairs = append(docPairs, key, v)
	}

	return wirebson.NewDocument(docPairs...)
}

// unmarshalExtJSON takes extended JSON object and unmarshals it into the [*wirebson.Document].
// If provided json is nil it also returns nil.
func unmarshalExtJSON(json *json.RawMessage) (out any, err error) {
	if json == nil {
		return nil, nil
	}

	if len(*json) == 0 {
		return nil, fmt.Errorf("Invalid object: %v", *json)
	}

	var raw any

	err = bson.UnmarshalExtJSON(*json, false, &raw)
	if err != nil {
		return nil, err
	}

	t, b, err := bson.MarshalValue(raw)
	if err != nil {
		return nil, err
	}

	switch t {
	case bson.TypeEmbeddedDocument:
		out = wirebson.RawDocument(b)
	case bson.TypeArray:
		out = wirebson.RawArray(b)
	case bson.TypeDouble,
		bson.TypeString,
		bson.TypeBinary,
		bson.TypeUndefined,
		bson.TypeObjectID,
		bson.TypeBoolean,
		bson.TypeDateTime,
		bson.TypeNull,
		bson.TypeRegex,
		bson.TypeDBPointer,
		bson.TypeJavaScript,
		bson.TypeSymbol,
		bson.TypeCodeWithScope,
		bson.TypeInt32,
		bson.TypeTimestamp,
		bson.TypeInt64,
		bson.TypeDecimal128,
		bson.TypeMinKey,
		bson.TypeMaxKey:
		fallthrough
	default:
		must.NoError(fmt.Errorf("Unrecognized type: %T", t))
	}

	return
}
