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

package tjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/lazyerrors"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// SchemaType represents JSON value type in JSON Schema.
type SchemaType string

// JSON value types defined by the JSON Schema.
const (
	Integer SchemaType = "integer"
	Number  SchemaType = "number"
	String  SchemaType = "string"
	Boolean SchemaType = "boolean"
	Array   SchemaType = "array"
	Object  SchemaType = "object"
)

// SchemaFormat represents additional information about JSON value type in JSON Schema.
type SchemaFormat string

// JSON value formats.
const (
	EmptyFormat SchemaFormat = ""

	// For Number.
	Double SchemaFormat = "double"
	Float  SchemaFormat = "float"

	// For Integer.
	Int64 SchemaFormat = "int64"
	Int32 SchemaFormat = "int32"

	// For String.
	Byte     SchemaFormat = "byte"
	UUID     SchemaFormat = "uuid"
	DateTime SchemaFormat = "date-time"
)

// Schema represents a supported subset of JSON Schema.
type Schema struct {
	Title       string             `json:"title,omitempty"`
	Description string             `json:"description,omitempty"`
	Type        SchemaType         `json:"type,omitempty"`
	Format      SchemaFormat       `json:"format,omitempty"`
	Properties  map[string]*Schema `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	PrimaryKey  []string           `json:"primary_key,omitempty"`
}

// Schemas for scalar types.
//
//nolint:unused // remove when they are used
var (
	doubleSchema = &Schema{
		Type: Number,
	}
	stringSchema = &Schema{
		Type: String,
	}
	binarySchema = &Schema{
		Type: Object,
		Properties: map[string]*Schema{
			"$b": {Type: String, Format: Byte},
			"s":  int32Schema,
		},
	}
	objectIDSchema = &Schema{
		Type:   String,
		Format: Byte,
	}
	boolSchema = &Schema{
		Type: Boolean,
	}
	dateTimeSchema = &Schema{
		Type:   String,
		Format: DateTime,
	}
	// No schema for null, it is a special case.
	regexSchema = &Schema{
		Type: Object,
		Properties: map[string]*Schema{
			"$r": {Type: String},
			"o":  {Type: String},
		},
	}
	int32Schema = &Schema{
		Type:   Integer,
		Format: Int32,
	}
	timestampSchema = &Schema{
		Type: Object,
		Properties: map[string]*Schema{
			"$t": {Type: String},
		},
	}
	int64Schema = &Schema{
		Type: Integer,
	}
)

// Equal returns true if the schemas are equal.
func (s *Schema) Equal(other *Schema) bool {
	if s == other {
		return true
	}

	// TODO compare significant fields only (ignore title, description, etc.)
	// TODO compare format according to type (for example, for Number, EmptyFormat == Double)
	// https://github.com/FerretDB/FerretDB/issues/683
	return reflect.DeepEqual(s, other)
}

// Marshal returns the JSON encoding of the schema.
func (s *Schema) Marshal() ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, lazyerrors.Error(err)
	}

	return b, nil
}

// Unmarshal parses the JSON-encoded schema.
func (s *Schema) Unmarshal(b []byte) error {
	r := bytes.NewReader(b)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	if err := dec.Decode(s); err != nil {
		return lazyerrors.Error(err)
	}
	if err := checkConsumed(dec, r); err != nil {
		return lazyerrors.Error(err)
	}

	return nil
}

// DocumentSchema returns a JSON Schema for the given document.
func DocumentSchema(doc *types.Document) (*Schema, error) {
	if !doc.Has("_id") {
		return nil, lazyerrors.New("document must have an _id")
	}

	schema := Schema{
		Type:       Object,
		Properties: make(map[string]*Schema, doc.Len()+1),
		PrimaryKey: []string{"_id"},
	}

	schema.Properties["$k"] = &Schema{Type: Array, Items: stringSchema}

	for _, k := range doc.Keys() {
		v := must.NotFail(doc.Get(k))
		s, err := valueSchema(v)
		if err != nil {
			return nil, lazyerrors.Error(err)
		}
		schema.Properties[k] = s
	}

	return &schema, nil
}

// valueSchema returns a schema for the given value.
func valueSchema(v any) (*Schema, error) {
	switch v := v.(type) {
	case *types.Document:
		return DocumentSchema(v)
	case *types.Array:
		return nil, lazyerrors.Errorf("%T is not supported yet", v)
	case float64:
		return doubleSchema, nil
	case string:
		return stringSchema, nil
	case types.Binary:
		return binarySchema, nil
	case types.ObjectID:
		return objectIDSchema, nil
	case bool:
		return boolSchema, nil
	case time.Time:
		// return dateTimeSchema, nil
		return nil, lazyerrors.Errorf("%T is not supported yet", v)
	case types.NullType:
		return nil, lazyerrors.Errorf("%T is not supported yet", v)
	case types.Regex:
		// return regexSchema, nil
		return nil, lazyerrors.Errorf("%T is not supported yet", v)
	case int32:
		return int32Schema, nil
	case types.Timestamp:
		// return timestampSchema, nil
		return nil, lazyerrors.Errorf("%T is not supported yet", v)
	case int64:
		return int64Schema, nil
	default:
		panic(fmt.Sprintf("not reached: %T", v))
	}
}
