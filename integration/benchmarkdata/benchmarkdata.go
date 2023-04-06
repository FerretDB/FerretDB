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

package benchmarkdata

import (
	"context"
	"sort"
	"strconv"

	"github.com/FerretDB/FerretDB/integration/shareddata"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Data represents the benchmark dataset that is inserted into
// the benchmark collection.
//
// It takes collection as an argument and inserts the benchmark data.
// The function should produce determinisitic sets of documents, as
// they are inserted to each benchmark target.
type Data func(context.Context, *mongo.Collection) error

// SimpleData is a set of 400 simple documents.
var SimpleData Data = func(ctx context.Context, coll *mongo.Collection) error {
	values := []any{
		"foo", 42, "42", bson.D{{"42", "hello"}},
	}
	var docs []any

	for i := 0; i < len(values)*100; {
		for _, doc := range values {
			docs = append(docs, bson.D{{"_id", i}, {"v", doc}})
			i++
		}
	}

	_, err := coll.InsertMany(ctx, docs)
	return err
}

// LargeDocument returns a document of size 43474B by concatenating
// all providers. The order of fields is non-deterministic because AllProviders
// returns all providers in a random order.
var LargeDocument Data = func(ctx context.Context, coll *mongo.Collection) error {
	var docs = shareddata.Docs()

	ap := shareddata.AllProviders()

	names := []string{}
	for _, p := range ap {
		names = append(names, p.Name())
	}
	sort.Strings(names)

	// XXX confirm that this gives a deterministic order.
	for _, name := range names {
		if ap.In(name) {
			docs = append(docs, shareddata.Docs(ap...)...)
		}
	}

	ld := bson.M{}
	ld["_id"] = primitive.NewObjectID()

	i := 0
	for _, doc := range docs {
		doc := doc.(primitive.D).Map()
		for _, v := range doc {
			ld[strconv.Itoa(i)] = v
			i++
		}
	}

	md, _ := bson.Marshal(ld)

	_, err := coll.InsertOne(ctx, md)
	return err
}
