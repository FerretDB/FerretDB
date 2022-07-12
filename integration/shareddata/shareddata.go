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
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// Provider is implemented by shared data sets that provide documents.
type Provider interface {
	// Docs returns shared data documents. All calls should return the same data.
	Docs() []bson.D
}

// Docs stores shared data documents as maps.
type Docs[idType constraints.Ordered] struct {
	data map[idType]map[string]any
}

// Docs implement Provider interface.
func (docs *Docs[idType]) Docs() []bson.D {
	ids := maps.Keys(docs.data)
	slices.Sort(ids)

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

// Values stores shared data documents as {"_id": key, "value": value} documents.
type Values[idType constraints.Ordered] struct {
	data map[idType]any
}

// Docs implement Provider interface.
func (values *Values[idType]) Docs() []bson.D {
	ids := maps.Keys(values.data)
	slices.Sort(ids)

	res := make([]bson.D, 0, len(values.data))
	for _, id := range ids {
		res = append(res, bson.D{{"_id", id}, {"value", values.data[id]}})
	}

	return res
}

// check interfaces
var (
	_ Provider = (*Docs[string])(nil)
	_ Provider = (*Values[string])(nil)
)
