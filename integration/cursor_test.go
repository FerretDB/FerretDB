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
	"context"
	"sync"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestCursorStress(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), nil)
	if err != nil {
		t.Error(err)
	}

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

		_, err = client.Database("test").Collection("foo").InsertMany(context.TODO(), insertDocs)
		require.NoError(t, err)
	}

	t.Cleanup(func() { require.NoError(t, client.Database("test").Drop(context.TODO())) })

	N := 10

	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			// create N clients
			client, err := mongo.Connect(context.TODO(), nil)
			if err != nil {
				t.Error(err)
			}

			coll := client.Database("test").Collection("foo")

			opts := &options.FindOptions{
				// set a small batch size to increase the frequency of getMores
				BatchSize: pointer.ToInt32(20),
				Sort:      bson.D{{"_id", 1}},
			}

			cur, err := coll.Find(context.TODO(), bson.D{}, opts)
			require.NoError(t, err)

			// iterate the cursor until it is exhausted or there is an error getting the next document
			for {
				if cur.TryNext(context.TODO()) {
					assert.True(t, cur.Next(context.TODO()))
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
