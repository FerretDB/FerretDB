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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// ParseSchema returns build-in description based on tigris schema description.
func ParseSchema(data []byte) (map[string]any, error) {
	var a map[string]any
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, lazyerrors.Error(err)
	}
	properties, ok := a["properties"]
	if !ok {
		return nil, lazyerrors.Errorf("cannot find properties")
	}
	return parseSchema(properties)
}

// parseSchema recursively parses tigris schema description.
func parseSchema(in any) (map[string]any, error) {
	switch in := in.(type) {
	case map[string]any:
		schema := make(map[string]any, len(in))
		for k, v := range in {
			var valSchema map[string]any
			var err error
			switch v := v.(type) {
			case map[string]any:
				key, ok := v["type"]
				if !ok {
					continue
				}
				switch key {
				case "object":
					properties, ok := in["properties"]
					if !ok {
						return nil, lazyerrors.Errorf("cannot find properties")
					}
					valSchema, err = parseSchema(properties)

				case "array":
					return nil, fmt.Errorf("arrays not supported")
				case "boolean":
					valSchema = boolSchema
				case "string":
					format, ok := v["format"]
					if ok {
						switch format {
						case "date-time":
							valSchema = dateTimeSchema
						default:
							return nil, fmt.Errorf("formaat %s not suported", format)
						}
					} else {
						valSchema = stringSchema
					}
				case "integer":
					err = lazyerrors.Errorf("integer not supported")
				case "number":
					err = lazyerrors.Errorf("float64 not supported")
				default:
					continue // descriptions etc
				}
			default:
				// ok
			}
			if err != nil {
				return nil, err
			}
			schema[k] = valSchema
		}
		return schema, nil
	}
	panic("unreachable code")
}

// DocumentSchema creates description of json schema doc in Tigris data format and adds $k to keep order.
func DocumentSchema(doc *types.Document) (map[string]any, error) {
	schema := make(map[string]any, doc.Len()+2)
	for _, key := range doc.Keys() {
		v := must.NotFail(doc.Get(key))
		valueSchema, err := valueSchema(v)
		if err != nil {
			return nil, err
		}
		schema[key] = valueSchema
	}
	schema["$k"] = doc.Keys()

	externalSchema := make(map[string]any, 3)
	externalSchema["$k"] = []string{"type", "properties"}
	externalSchema["type"] = "object"
	externalSchema["properties"] = schema
	return externalSchema, nil
}

// valueSchema returns schema for value.
func valueSchema(v any) (map[string]any, error) {
	switch v := v.(type) {
	case *types.Document:
		return DocumentSchema(v)

	case *types.Array:
		return nil, lazyerrors.Errorf("arrays not supported yet")

	case float64:
		return map[string]any{"type": "number"}, nil

	case string:
		return map[string]any{"type": "string"}, nil

	case types.Binary:
		schema := map[string]any{
			"$k":   []string{"type", "properties"},
			"type": "object",
			"properties": map[string]any{
				"$b": map[string]any{"type": "string", "format": "byte"},   // binary data
				"s":  map[string]any{"type": "integer", "format": "int32"}, // binary subtype
			},
		}
		return schema, nil

	case types.ObjectID:
		return map[string]any{
			"$k":   []string{"type", "properties"},
			"type": "object",
			"properties": map[string]any{
				"$o": map[string]any{"type": "string"},
			},
		}, nil

	case bool:
		return map[string]any{"type": "boolean"}, nil

	case time.Time:
		return map[string]any{
			"type":   "string",
			"format": "date-time",
		}, nil

	case types.NullType:
		return nil, lazyerrors.Errorf("cannot determine type")

	case types.Regex:
		return map[string]any{
			"$k":   []string{"type", "properties"},
			"type": "object",
			"properties": map[string]any{
				"$r": map[string]any{"type": "string"},
				"o":  map[string]any{"type": "string"},
			},
		}, nil

	case int32:
		return nil, lazyerrors.Errorf("int32 not supported yet")

	case types.Timestamp:
		return map[string]any{
			"$k":   []string{"type", "properties"},
			"type": "object",
			"properties": map[string]any{
				"$t": map[string]any{"type": "string"},
			},
		}, nil

	case int64:
		return nil, lazyerrors.Errorf("int64 not supported yet")

	default:
		return nil, lazyerrors.Errorf("%v not supported yet", v)
	}
}
