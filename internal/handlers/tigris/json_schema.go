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

package tigris

import (
	"errors"
	"fmt"

	"github.com/FerretDB/FerretDB/internal/handlers/common"
	"github.com/FerretDB/FerretDB/internal/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

// getJSONSchema returns a marshaled JSON schema received from validator -> $jsonSchema.
func getJSONSchema(doc *types.Document) (*tjson.Schema, error) {
	v, err := common.GetRequiredParam[*types.Document](doc, "validator")
	if err != nil {
		return nil, err
	}

	schema, err := common.GetRequiredParam[*types.Document](v, "$jsonSchema")
	if err != nil {
		return nil, err
	}

	return schemaFromDocument(schema)
}

// schema creates a new TJSON Schema from types.Document format.
// The given doc should contain the keys typical for schema (e.g. title, type etc).
// In fact, this function coverts a document to tjson.JSONSchema, so the given doc should represent a valid JSON schema.
// If you need a function that returns a possible schema for the given document, see tjson.DocumentSchema
func schemaFromDocument(doc *types.Document) (*tjson.Schema, error) {
	schema := tjson.Schema{}

	if v := doc.Remove("title"); v != nil {
		title, ok := v.(string)
		if !ok {
			return nil, errors.New("invalid schema, the following key should be a string: title")
		}

		schema.Title = title
	}

	if v := doc.Remove("description"); v != nil {
		description, ok := v.(string)
		if !ok {
			return nil, errors.New("invalid schema, the following key should be a string: description")
		}

		schema.Description = description
	}

	if v := doc.Remove("type"); v != nil {
		schemaType, ok := v.(string)
		if !ok {
			return nil, errors.New("invalid schema, the following key should be a string: type")
		}

		schema.Type = tjson.SchemaType(schemaType)
	}

	if v := doc.Remove("format"); v != nil {
		format, ok := v.(string)
		if !ok {
			return nil, errors.New("invalid schema, the following key should be a string: format")
		}

		schema.Format = tjson.SchemaFormat(format)
	}

	if v := doc.Remove("primary_key"); v != nil {
		arr, ok := v.(*types.Array)
		if !ok {
			return nil, errors.New("invalid schema, the following key should be an array: primary_key")
		}

		schema.PrimaryKey = make([]string, arr.Len())

		for i := 0; i < arr.Len(); i++ {
			str, ok := must.NotFail(arr.Get(i)).(string)
			if !ok {
				return nil, errors.New("invalid schema, primary_key values should be strings")
			}
			schema.PrimaryKey[i] = str
		}
	}

	if v := doc.Remove("properties"); v != nil {
		schema.Properties = map[string]*tjson.Schema{}

		props, ok := v.(*types.Document)
		if !ok {
			return nil, errors.New("invalid schema, the following key should be a document: properties")
		}

		for _, key := range v.(*types.Document).Keys() {
			prop, err := common.GetRequiredParam[*types.Document](props, key)
			if err != nil {
				return nil, err
			}

			propSchema, err := schemaFromDocument(prop)
			if err != nil {
				return nil, err
			}

			schema.Properties[key] = propSchema
		}
	}

	if v := doc.Remove("items"); v != nil {
		items, ok := v.(*types.Document)
		if !ok {
			return nil, errors.New("invalid schema, the following key should be a document: items")
		}

		sch, err := schemaFromDocument(items)
		if err != nil {
			return nil, err
		}

		schema.Items = sch
	}

	// If any other fields are left, the doc doesn't represent a valid schema.
	if len(doc.Keys()) > 0 {
		msg := fmt.Sprintf("invalid schema, the following keys are not supported: %s", doc.Keys())
		return nil, errors.New(msg)
	}

	return &schema, nil
}
