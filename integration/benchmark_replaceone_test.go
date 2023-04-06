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

package integration

import (
	"strconv"
	"testing"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func BenchmarkReplaceOne(b *testing.B) {
	ctx, coll := setup.Setup(b, shareddata.Composites)
	defer setup.Shutdown()

	b.Run("ReplaceWithSelf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			objectID := primitive.NewObjectID()
			filter := bson.D{{"_id", objectID}}

			doc, err := largeDocument(b, objectID)
			require.NoError(b, err)

			ior, err := coll.InsertOne(ctx, doc, nil)
			require.NoError(b, err)
			require.Equal(b, ior.InsertedID, objectID)

			replacement, err := largeDocument(b, objectID)
			require.NoError(b, err)

			res, err := coll.ReplaceOne(ctx, filter, replacement)
			require.NoError(b, err)
			require.Equal(b, int64(1), res.ModifiedCount)
		}
	})

}

// returns a 43474B BSON document.
func largeDocument(b *testing.B, objectID primitive.ObjectID) ([]byte, error) {
	ld := bson.M{}
	ld["_id"] = objectID

	// for now just concatenate all providers to create a large document.
	// XXX: use external data for benchmarking in the future -
	// https://www.percona.com/blog/sample-datasets-for-benchmarking-and-testing/
	docs := shareddata.Docs(shareddata.AllProviders()...)

	i := 0
	for _, doc := range docs {
		doc := doc.(primitive.D).Map()
		for _, v := range doc {
			ld[strconv.Itoa(i)] = v
			i++
		}
	}

	return bson.Marshal(ld)
}
