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

	"github.com/FerretDB/FerretDB/internal/util/must"

	"github.com/FerretDB/FerretDB/internal/tjson"
	"github.com/FerretDB/FerretDB/internal/types"
)

// getJSONSchema returns a masrshaled JSON schema received from validator -> $jsonSchema.
func getJSONSchema(doc *types.Document) (*tjson.Schema, error) {
	v, err := doc.Get("validator")
	if err != nil {
		return nil, errors.New("required parameter `validator` is missing")
	}

	s, err := v.(*types.Document).Get("$jsonSchema")
	if err != nil {
		return nil, errors.New("required parameter `$jsonSchema` is missing")
	}

	return schemaFromDocument(s.(*types.Document))
}

// schema creates a new TJSON Schema based on the types.Document format.
// The given doc should contain the keys typical for schema (e.g. title, type etc).
func schemaFromDocument(doc *types.Document) (*tjson.Schema, error) {
	schema := tjson.Schema{}

	if v := doc.Remove("title"); v != nil {
		schema.Title = v.(string)
	}

	if v := doc.Remove("description"); v != nil {
		schema.Description = v.(string)
	}

	if v := doc.Remove("type"); v != nil {
		schema.Type = tjson.SchemaType(v.(string))
	}

	if v := doc.Remove("format"); v != nil {
		schema.Format = tjson.SchemaFormat(v.(string))
	}

	if v := doc.Remove("primary_key"); v != nil {
		arr := v.(*types.Array)
		schema.PrimaryKey = make([]string, arr.Len())
		for i := 0; i < arr.Len(); i++ {
			schema.PrimaryKey = append(schema.PrimaryKey, must.NotFail(arr.Get(i)).(string))
		}
	}

	if v := doc.Remove("properties"); v != nil {
		schema.Properties = map[string]*tjson.Schema{}

		for _, key := range v.(*types.Document).Keys() {
			prop, err := v.(*types.Document).Get(key)
			if err != nil {
				panic(err)
			}

			propSchema, err := schemaFromDocument(prop.(*types.Document))
			if err != nil {
				return nil, err
			}
			schema.Properties[key] = propSchema
		}
	}

	if v := doc.Remove("items"); v != nil {
		sch, err := schemaFromDocument(v.(*types.Document))
		if err != nil {
			return nil, err
		}
		schema.Items = sch
	}

	// If any other fields are left, the doc doesn't represent a valid schema.
	if len(doc.Keys()) > 0 {
		msg := fmt.Sprintf("invalid schema, the follwing keys are not supported: %s", doc.Keys())
		return nil, errors.New(msg)
	}

	return &schema, nil
}
