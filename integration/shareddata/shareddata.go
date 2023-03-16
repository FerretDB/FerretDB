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
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/maps"
)

// unset represents a field that should not be set.
var unset = struct{}{}

// Provider is implemented by shared data sets that provide documents.
type Provider interface {
	// Name returns provider name.
	Name() string

	// Handlers returns handlers compatible with this provider.
	Handlers() []string

	// Validators returns validators for the given handler and collection.
	// For example, for Tigris it should return a map with they key $tigrisSchemaString
	// and the value containing Tigris' JSON schema string.
	Validators(handler, collection string) map[string]any

	// Docs returns shared data documents.
	// All calls should return the same set of documents, but may do so in different order.
	Docs() []bson.D

	// IsCompatible returns true if the given backend is compatible with this provider.
	IsCompatible(backend string) bool
}

// AllProviders returns all providers in random order.
func AllProviders() Providers {
	providers := []Provider{
		Scalars,

		Doubles,
		BigDoubles,
		Strings,
		Binaries,
		ObjectIDs,
		Bools,
		DateTimes,
		Nulls,
		Regexes,
		Int32s,
		Timestamps,
		Int64s,
		Unsets,
		ObjectIDKeys,

		Composites,
		PostgresEdgeCases,

		DocumentsDoubles,
		DocumentsStrings,
		DocumentsDocuments,

		ArrayStrings,
		ArrayDoubles,
		ArrayInt32s,
		ArrayRegexes,
		ArrayDocuments,
	}

	// check that names are unique and randomize order
	res := make(map[string]Provider, len(providers))
	for _, p := range providers {
		n := p.Name()
		if _, ok := res[n]; ok {
			panic("duplicate provider name: " + n)
		}

		res[n] = p
	}

	return maps.Values(res)
}

// Providers are array of providers.
type Providers []Provider

// Remove specified providers and return remaining providers.
func (ps Providers) Remove(removeProviderNames ...string) Providers {
	res := make([]Provider, 0, len(ps))

	for _, p := range ps {
		keep := true

		for _, name := range removeProviderNames {
			if p.Name() == name {
				keep = false
				break
			}
		}

		if keep {
			res = append(res, p)
		}
	}

	return res
}

// Docs returns all documents from given providers.
func Docs(providers ...Provider) []any {
	var res []any
	for _, p := range providers {
		for _, doc := range p.Docs() {
			res = append(res, doc)
		}
	}
	return res
}

// IDs returns all document's _id values (that must be present in each document) from given providers.
func IDs(providers ...Provider) []any {
	var res []any
	for _, p := range providers {
		for _, doc := range p.Docs() {
			id, ok := doc.Map()["_id"]
			if !ok {
				panic(fmt.Sprintf("no _id in %+v", doc))
			}
			res = append(res, id)
		}
	}
	return res
}

// Values stores shared data documents as {"_id": key, "v": value} documents.
type Values[idType comparable] struct {
	name       string
	handlers   []string
	validators map[string]map[string]any // handler -> validator name -> validator
	data       map[idType]any
}

// Name implement Provider interface.
func (values *Values[idType]) Name() string {
	return values.name
}

// Handlers implement Provider interface.
func (values *Values[idType]) Handlers() []string {
	return values.handlers
}

// Validators implement Provider interface.
func (values *Values[idType]) Validators(handler, collection string) map[string]any {
	switch handler {
	case "tigris":
		validators := make(map[string]any, len(values.validators[handler]))
		for key, value := range values.validators[handler] {
			validators[key] = strings.ReplaceAll(value.(string), "%%collection%%", collection)
		}
		return validators

	default:
		return values.validators[handler]
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
	if backend == "mongodb" {
		return true
	}

	var handler string

	switch backend {
	case "ferretdb-pg":
		handler = "pg"
	case "ferretdb-tigris":
		handler = "tigris"
	default:
		panic(fmt.Sprintf("unknown backend: %s", backend))
	}

	return slices.Contains(values.handlers, handler)
}

// check interfaces
var (
	_ Provider = (*Values[string])(nil)
)
