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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Inserter interface {
	Insert(context.Context, testing.TB, *mongo.Collection)
}

type Provider interface {
	Docs() []bson.D
}

func Map[idType comparable](t testing.TB, docs []bson.D) map[idType]bson.D {
	t.Helper()

	res := make(map[idType]bson.D, len(docs))
	for _, doc := range docs {
		id, ok := doc.Map()["_id"].(idType)
		require.True(t, ok)
		res[id] = doc
	}

	return res
}

type Values[idType constraints.Ordered] struct {
	data map[idType]any
}

func (values *Values[idType]) Insert(ctx context.Context, t testing.TB, collection *mongo.Collection) {
	for _, doc := range values.Docs() {
		_, err := collection.InsertOne(ctx, doc)
		require.NoError(t, err)
	}
}

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
	_ Inserter = (*Values[string])(nil)
	_ Provider = (*Values[string])(nil)
)
