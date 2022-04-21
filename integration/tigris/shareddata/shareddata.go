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

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// TypedValues stores shared data documents as {"_id": id, type: value} documents, where type is TigrisDB type name.
type TypedValues[idType constraints.Ordered] struct {
	data map[idType]any
}

// Docs implement Provider interface.
func (values *TypedValues[idType]) Docs() []bson.D {
	ids := maps.Keys(values.data)
	slices.Sort(ids)

	res := make([]bson.D, 0, len(values.data))
	for _, id := range ids {
		switch value := values.data[id].(type) {
		case string:
			res = append(res, bson.D{{"_id", id}, {"string", value}})
		case float64:
			res = append(res, bson.D{{"_id", id}, {"number", value}})
		default:
			panic("not implemented")
		}
	}

	return res
}

// check interfaces
var (
	_ shareddata.Provider = (*TypedValues[string])(nil)
)
