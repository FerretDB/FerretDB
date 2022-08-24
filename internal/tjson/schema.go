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
	"errors"
	"fmt"
	"time"

	"github.com/FerretDB/FerretDB/internal/handlers/common"

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
// For composite types schemas are equal if their types and subschemas are equal.
// For scalar types schemas are equal if their types and formats are equal.
func (s *Schema) Equal(other *Schema) bool {
	if s == other {
		return true
	}

	if s.Type != other.Type {
		return false
	}

	switch s.Type {
	case Object:
		// If `s` and `other` are objects, compare their properties.
		if len(s.Properties) != len(other.Properties) {
			return false
		}
		for k, v := range s.Properties {
			vOther, ok := other.Properties[k]
			if !ok {
				return false
			}
			if eq := v.Equal(vOther); !eq {
				return false
			}
		}
		return true
	case Array:
		// If `s` and `other` are arrays, compare their items.
		if s.Items == nil || other.Items == nil {
			panic("schema.Equal: array with nil items")
		}
		return s.Items.Equal(other.Items)
	case String, Integer, Number, Boolean:
		// For scalar types, it's enough to compare their formats.
		if s.Format == other.Format {
			return true
		}
	default:
		panic(fmt.Sprintf("schema.Equal: unknown type `%s`", s.Type))
	}

	// If formats don't match, normalize schemas: empty format is equal to double for numbers and int64 for integers,
	// see https://docs.tigrisdata.com/overview/schema#data-types.
	formatS, formatOther := s.Format, other.Format
	switch s.Type {
	case Number:
		if s.Format == EmptyFormat {
			formatS = Double
		}
		if other.Format == EmptyFormat {
			formatOther = Double
		}
	case Integer:
		if s.Format == EmptyFormat {
			formatS = Int64
		}
		if other.Format == EmptyFormat {
			formatOther = Int64
		}
	case Array, Boolean, Object, String:
		// do nothing: these types don't have "default" format
	}
	return formatS == formatOther
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

// UnmarshalFromDocument creates a new TJSON Schema from types.Document format.
// TODO The given doc should contain the keys typical for schema (e.g. title, type etc).
// TODO In fact, this function coverts a document to tjson.JSONSchema, so the given doc should represent a valid JSON schema.
// TODO If you need a function that returns a possible schema for the given document, see tjson.DocumentSchema.
func (s *Schema) UnmarshalFromDocument(doc *types.Document) error {
	if v := doc.Remove("title"); v != nil {
		title, ok := v.(string)
		if !ok {
			return errors.New("invalid schema, the following key should be a string: title")
		}

		s.Title = title
	}

	if v := doc.Remove("description"); v != nil {
		description, ok := v.(string)
		if !ok {
			return errors.New("invalid schema, the following key should be a string: description")
		}

		s.Description = description
	}

	if v := doc.Remove("type"); v != nil {
		schemaType, ok := v.(string)
		if !ok {
			return errors.New("invalid schema, the following key should be a string: type")
		}

		s.Type = SchemaType(schemaType)
	}

	if v := doc.Remove("format"); v != nil {
		format, ok := v.(string)
		if !ok {
			return errors.New("invalid schema, the following key should be a string: format")
		}

		s.Format = SchemaFormat(format)
	}

	if v := doc.Remove("primary_key"); v != nil {
		arr, ok := v.(*types.Array)
		if !ok {
			return errors.New("invalid schema, the following key should be an array: primary_key")
		}

		s.PrimaryKey = make([]string, arr.Len())

		for i := 0; i < arr.Len(); i++ {
			str, ok := must.NotFail(arr.Get(i)).(string)
			if !ok {
				return errors.New("invalid schema, primary_key values should be strings")
			}
			s.PrimaryKey[i] = str
		}
	}

	if v := doc.Remove("properties"); v != nil {
		s.Properties = map[string]*Schema{}

		props, ok := v.(*types.Document)
		if !ok {
			return errors.New("invalid schema, the following key should be a document: properties")
		}

		for _, key := range v.(*types.Document).Keys() {
			prop, err := common.GetRequiredParam[*types.Document](props, key)
			if err != nil {
				return err
			}

			subsschema := new(Schema)
			err = subsschema.UnmarshalFromDocument(prop)
			if err != nil {
				return err
			}

			s.Properties[key] = subsschema
		}
	}

	if v := doc.Remove("items"); v != nil {
		items, ok := v.(*types.Document)
		if !ok {
			return errors.New("invalid schema, the following key should be a document: items")
		}

		subschema := new(Schema)
		err := subschema.UnmarshalFromDocument(items)
		if err != nil {
			return err
		}

		s.Items = subschema
	}

	// If any other fields are left, the doc doesn't represent a valid schema.
	if len(doc.Keys()) > 0 {
		msg := fmt.Sprintf("invalid schema, the following keys are not supported: %s", doc.Keys())
		return errors.New(msg)
	}

	return nil
}

// DocumentSchema returns a JSON Schema for the given top-level document.
// Top-level documents are documents that must have _id which will be used as primary key.
func DocumentSchema(doc *types.Document) (*Schema, error) {
	if !doc.Has("_id") {
		return nil, lazyerrors.New("document must have an _id")
	}

	return subdocumentSchema(doc, "_id")
}

// subdocumentSchema returns a JSON Schema for the given subdocument.
// Subdocument is a "nested" document that can be used as a property of another document or subdocument.
// The difference between subdocument and document is that subdocument doesn't have to contain the _id key.
func subdocumentSchema(doc *types.Document, pkey ...string) (*Schema, error) {
	schema := Schema{
		Type:       Object,
		Properties: make(map[string]*Schema, doc.Len()+1),
		PrimaryKey: pkey,
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
		return subdocumentSchema(v)
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
		return dateTimeSchema, nil
	case types.NullType:
		return nil, lazyerrors.Errorf("%T is not supported yet", v)
	case types.Regex:
		return regexSchema, nil
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
