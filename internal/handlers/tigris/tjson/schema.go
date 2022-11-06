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
	Title      string             `json:"title,omitempty"`
	Type       SchemaType         `json:"type,omitempty"`
	Format     SchemaFormat       `json:"format,omitempty"`
	Properties map[string]*Schema `json:"properties,omitempty"`
	Items      *Schema            `json:"items,omitempty"`
	PrimaryKey []string           `json:"primary_key,omitempty"`

	// those fields are not used, but required to be there for DisallowUnknownFields
	Description    string `json:"description,omitempty"`
	CollectionType string `json:"collection_type,omitempty"`
}

// Schemas for scalar types.
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
		return err
	}
	if err := checkConsumed(dec, r); err != nil {
		return err
	}

	// Add $k properties that are necessary for documents.
	s.addDocumentProperties()

	// If Type is not set, it's a high-level schema, so we set Type as Object to make it explicit.
	if s.Type == "" {
		s.Type = Object
	}

	return nil
}

// addDocumentProperties adds missing $k properties to all the schema's documents (top-level and nested).
func (s *Schema) addDocumentProperties() {
	if s.Type == Array && s.Items.Type == Object {
		s.Items.addDocumentProperties()
		return
	}

	if s.Type != Object && s.Type != "" {
		return
	}

	for _, subschema := range s.Properties {
		if subschema.Type != Object {
			continue
		}

		subschema.addDocumentProperties()
	}

	// If the current object is not a special object (binary, regex, timestamp, decimal),
	// and $k property is not set, then $k property is missing and needs to be added.
	specials := []string{"$k", "$b", "$r", "$t", "$n"}
	for _, special := range specials {
		if _, ok := s.Properties[special]; ok {
			return
		}
	}

	if s.Properties == nil {
		s.Properties = map[string]*Schema{}
	}

	s.Properties["$k"] = &Schema{Type: Array, Items: stringSchema}
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

		// If s == nil it's a field with null, according to Tigris' logic, we don't set schema for this field.
		if s != nil {
			schema.Properties[k] = s
		}
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
		// According to the current Tigris' logic the field that is set as null is valid but not present in the schema
		return nil, nil
	case types.Regex:
		return regexSchema, nil
	case int32:
		return int32Schema, nil
	case types.Timestamp:
		return timestampSchema, nil
	case int64:
		return int64Schema, nil
	default:
		panic(fmt.Sprintf("not reached: %T", v))
	}
}
