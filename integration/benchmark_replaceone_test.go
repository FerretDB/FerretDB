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
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func BenchmarkReplaceOne(b *testing.B) {
	ctx, coll := setup.Setup(b, shareddata.Composites)
	defer setup.Shutdown()

	b.Run("ReplaceWithFilter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			objectID := primitive.NewObjectID()
			doc := largeDocument(b, objectID)
			b.Log(doc.Values()...)

			_, err := coll.InsertOne(ctx, doc, nil)
			require.NoError(b, err)

			ii, err := coll.InsertOne(
				ctx,
				largeDocument(b, objectID),
			)
			b.Log(err, ii.InsertedID, objectID)

			replacement := doc
			replacement.Set("_id", primitive.NewObjectID())

			filter := bson.D{{"_id", objectID}}

			res, err := coll.ReplaceOne(ctx, filter, replacement)
			require.Equal(b, 1, res.ModifiedCount)
		}
	})

}

func largeDocument(b *testing.B, objectID primitive.ObjectID) *types.Document {
	ld := types.Document{}
	ld.Set("_id", objectID)

	docs := shareddata.Int64s.Docs()

	i := 0
	for _, doc := range docs {
		m := doc.Map()
		for _, v := range m {
			ld.Set(strconv.Itoa(i), v)
			i++
		}
	}

	if !objectID.IsZero() {
		b.Log(true)
		b.Log(ld.Keys()[0] == objectID.String())
	}

	return &ld
}
