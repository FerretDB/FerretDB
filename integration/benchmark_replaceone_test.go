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
	"fmt"
	"testing"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var objectID = primitive.ObjectID{100, 45, 201, 30, 9, 97, 166, 6, 75, 239, 151, 226}

func BenchmarkReplaceOne(b *testing.B) {
	ctx, coll := setup.Setup(b, shareddata.AllProviders()...)

	filter := bson.D{{"_id", objectID}}

	b.Run("ReplaceWithFilter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := coll.InsertOne(ctx, largeDocument(), nil)
			require.NoError(b, err)

			res, err := coll.ReplaceOne(ctx, filter, filter)
			require.Equal(b, 1, res.ModifiedCount)
		}
	})

}

func largeDocument() types.Document {
	ld := types.Document{}
	ld.Set("_id", objectID)

	docs := shareddata.Composites.Docs()

	i := 0
	for _, doc := range docs {
		m := doc.Map()
		delete(m, "_id")
		for _, v := range m {
			// keys are single letters and are in alphabetical order
			// so that we generate the same document each time.
			k := fmt.Sprint(i + 'a')
			ld.Set(k, v)
			i++
		}
	}

	return ld
}
