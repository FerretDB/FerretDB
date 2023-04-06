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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Data func(context.Context, *mongo.Collection) error

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
