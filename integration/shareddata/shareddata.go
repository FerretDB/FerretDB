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

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// unset represents a field that should not be set.
var unset = struct{}{}

// Provider is implemented by shared data sets that provide documents.
type Provider interface {
	// Docs returns shared data documents.
	// All calls should return the same set of documents, but may do so in different order.
	Docs() []bson.D
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

// Maps stores shared data documents as maps.
//
// TODO replace constraints.Ordered with comparable: https://github.com/FerretDB/FerretDB/issues/914
type Maps[idType constraints.Ordered] struct {
	data map[idType]map[string]any
}

// Docs implement Provider interface.
func (docs *Maps[idType]) Docs() []bson.D {
	ids := maps.Keys(docs.data)
	slices.Sort(ids) // TODO remove: https://github.com/FerretDB/FerretDB/issues/914

	res := make([]bson.D, 0, len(docs.data))
	for _, id := range ids {
		doc := docs.data[id]

		d := make(bson.D, 0, len(doc)+1)
		d = append(d, bson.E{"_id", id})
		for k, v := range doc {
			d = append(d, bson.E{k, v})
		}

		res = append(res, d)
	}

	return res
}

// Values stores shared data documents as {"_id": key, "v": value} documents.
//
// TODO replace constraints.Ordered with comparable: https://github.com/FerretDB/FerretDB/issues/914
type Values[idType constraints.Ordered] struct {
	data map[idType]any
}

// Docs implement Provider interface.
func (values *Values[idType]) Docs() []bson.D {
	ids := maps.Keys(values.data)
	slices.Sort(ids) // TODO remove: https://github.com/FerretDB/FerretDB/issues/914

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

// check interfaces
var (
	_ Provider = (*Maps[string])(nil)
	_ Provider = (*Values[string])(nil)
)
