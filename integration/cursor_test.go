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
	"sync"
	"testing"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCursorStress(t *testing.T) {
	ctx, collection := setup.Setup(t)

	defaultBatchSize := 101

	provider := shareddata.BenchmarkSmallDocuments

	iter := provider.NewIterator()

	for {
		docs, err := iterator.ConsumeValuesN(iter, defaultBatchSize)
		require.NoError(t, err)

		if docs == nil {
			break
		}

		insertDocs := make([]any, len(docs))
		for i := range insertDocs {
			insertDocs[i] = docs[i]
		}

		_, err = collection.InsertMany(ctx, insertDocs)
		require.NoError(t, err)
	}

	N := 10

	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// create N clients
			client, err := mongo.Connect(ctx, nil)
			if err != nil {
				t.Error(err)
			}

			coll := client.Database("TestCursorStress").Collection(collection.Name())

			cur, err := coll.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"_id", 1}}))
			require.NoError(t, err)

			// Iterate the cursor until the cursor is exhausted or there is an error getting the next document
			for {
				if cur.TryNext(ctx) {
					assert.True(t, cur.Next(ctx))
				}

				if err := cur.Err(); err != nil {
					t.Error(err)
				}
				if cur.ID() == 0 {
					break
				}
			}
		}()
	}

	wg.Wait()
}
