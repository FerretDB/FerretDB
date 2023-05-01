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

package shareddata

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Provider is implemented by shared data sets that provide documents.
type Provider interface {
	// Name returns provider name.
	Name() string

	// Validators returns validators for the given backend and collection.
	// For example, for ferretdb-tigris it should return a map with the key $tigrisSchemaString
	// and the value containing Tigris' JSON schema string.
	Validators(backend, collection string) map[string]any

	// Docs returns shared data documents.
	// All calls should return the same set of documents, but may do so in different order.
	Docs() []bson.D

	// IsCompatible returns true if the given backend is compatible with this provider.
	IsCompatible(backend string) bool
}

// Values stores shared data documents as {"_id": key, "v": value} documents.
type Values[idType comparable] struct {
	name       string
	backends   []string
	validators map[string]map[string]any // backend -> validator name -> validator
	data       map[idType]any
}

// Name implement Provider interface.
func (values *Values[idType]) Name() string {
	return values.name
}

// Validators implement Provider interface.
func (values *Values[idType]) Validators(backend, collection string) map[string]any {
	switch backend {
	case "ferretdb-tigris":
		validators := make(map[string]any, len(values.validators[backend]))
		for key, value := range values.validators[backend] {
			validators[key] = strings.ReplaceAll(value.(string), "%%collection%%", collection)
		}
		return validators
	default:
		return values.validators[backend]
	}
}

// Docs implement Provider interface.
func (values *Values[idType]) Docs() []bson.D {
	ids := maps.Keys(values.data)

	res := make([]bson.D, 0, len(values.data))
	for _, id := range ids {
		doc := bson.D{{"_id", id}}
		v := values.data[id]
		if v != unset {
			doc = append(doc, bson.E{"v", v})
		}
		res = append(res, doc)
	}

	return res
}

// IsCompatible returns true if the given backend is compatible with this provider.
func (values *Values[idType]) IsCompatible(backend string) bool {
	return slices.Contains(values.backends, backend)
}

// field represents a field in a document.
type field struct {
	Value any
	Key   string
}

// Fields is a slice of ordered field name value pair.
// To avoid fields being inserted in different order between compat and target, use a slice instead of a map.
type Fields []field

// NewTopLevelFieldsProvider creates a new TopLevelValues provider.
func NewTopLevelFieldsProvider[id comparable](name string, backends []string, validators map[string]map[string]any, data map[id]Fields) Provider {
	return &topLevelValues[id]{
		name:       name,
		backends:   backends,
		validators: validators,
		data:       data,
	}
}

// topLevelValues stores shared data documents as {"_id": key, "field1": value1, "field2": value2, ...} documents.
//
//nolint:vet // for readability
type topLevelValues[id comparable] struct {
	name       string
	backends   []string
	validators map[string]map[string]any // backend -> validator name -> validator
	data       map[id]Fields
}

// Name implements Provider interface.
func (t *topLevelValues[id]) Name() string {
	return t.name
}

// Validators implements Provider interface.
func (t *topLevelValues[id]) Validators(backend, collection string) map[string]any {
	switch backend {
	case "ferretdb-tigris":
		validators := make(map[string]any, len(t.validators[backend]))

		for key, value := range t.validators[backend] {
			validators[key] = strings.ReplaceAll(value.(string), "%%collection%%", collection)
		}

		return validators
	default:
		return t.validators[backend]
	}
}

// Docs implements Provider interface.
func (t *topLevelValues[id]) Docs() []bson.D {
	ids := maps.Keys(t.data)

	res := make([]bson.D, 0, len(t.data))

	for _, id := range ids {
		doc := bson.D{{"_id", id}}

		fields := t.data[id]

		for _, field := range fields {
			doc = append(doc, bson.E{Key: field.Key, Value: field.Value})
		}

		res = append(res, doc)
	}

	return res
}

// IsCompatible implements Provider interface.
func (t *topLevelValues[id]) IsCompatible(backend string) bool {
	return slices.Contains(t.backends, backend)
}

// check interfaces
var (
	_ Provider = (*Values[string])(nil)
	_ Provider = (*topLevelValues[string])(nil)
)
