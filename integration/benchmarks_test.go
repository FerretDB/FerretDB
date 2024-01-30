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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

func BenchmarkFind(b *testing.B) {
	provider := shareddata.BenchmarkSmallDocuments

	b.Run(provider.Name(), func(b *testing.B) {
		s := setup.SetupWithOpts(b, &setup.SetupOpts{
			BenchmarkProvider: provider,
		})

		for name, bc := range map[string]struct {
			filter bson.D
		}{
			"Int32ID": {
				filter: bson.D{{"_id", int32(42)}},
			},
			"Int32One": {
				filter: bson.D{{"id", int32(42)}},
			},
			"Int32Many": {
				filter: bson.D{{"v", int32(42)}},
			},
			"Int32ManyDotNotation": {
				filter: bson.D{{"v.foo", int32(42)}},
			},
		} {
			b.Run(name, func(b *testing.B) {
				var firstDocs, docs int

				for i := 0; i < b.N; i++ {
					cursor, err := s.Collection.Find(s.Ctx, bc.filter)
					require.NoError(b, err)

					docs = 0
					for cursor.Next(s.Ctx) {
						docs++
					}

					require.NoError(b, cursor.Close(s.Ctx))
					require.NoError(b, cursor.Err())
					require.Positive(b, docs)

					if firstDocs == 0 {
						firstDocs = docs
					}
				}

				b.StopTimer()

				require.Equal(b, firstDocs, docs)

				b.ReportMetric(float64(docs), "docs-returned")
			})
		}
	})
}

func BenchmarkReplaceOne(b *testing.B) {
	provider := shareddata.BenchmarkSettingsDocuments

	s := setup.SetupWithOpts(b, &setup.SetupOpts{
		BenchmarkProvider: provider,
	})
	ctx, collection := s.Ctx, s.Collection

	// use the last document by the natural order to make non-pushdown path slower

	cursor, err := collection.Find(ctx, bson.D{})
	require.NoError(b, err)

	var lastRaw bson.Raw
	for cursor.Next(ctx) {
		lastRaw = cursor.Current
	}
	require.NoError(b, cursor.Err())
	require.NoError(b, cursor.Close(ctx))

	var doc bson.D
	require.NoError(b, bson.Unmarshal(lastRaw, &doc))
	require.Equal(b, "_id", doc[0].Key)
	require.NotEmpty(b, doc[0].Value)
	require.NotZero(b, doc[1].Value)

	b.Run(provider.Name(), func(b *testing.B) {
		filter := bson.D{{"_id", doc[0].Value}}
		var res *mongo.UpdateResult

		for i := 0; i < b.N; i++ {
			doc[1].Value = int64(i + 1)

			res, err = collection.ReplaceOne(ctx, filter, doc)
			require.NoError(b, err)
			require.Equal(b, int64(1), res.MatchedCount)
			require.Equal(b, int64(1), res.ModifiedCount)
		}

		b.StopTimer()

		var actual bson.D
		err = collection.FindOne(ctx, filter).Decode(&actual)
		require.NoError(b, err)
		AssertEqualDocuments(b, doc, actual)
	})
}

func BenchmarkInsertMany(b *testing.B) {
	ctx, collection := setup.Setup(b)

	for _, provider := range shareddata.AllBenchmarkProviders() {
		total, err := iterator.ConsumeCount(provider.NewIterator())
		require.NoError(b, err)

		var batchSizes []int
		for _, batchSize := range []int{1, 10, 100, 1000} {
			if batchSize <= total {
				batchSizes = append(batchSizes, batchSize)
			}
		}

		for _, batchSize := range batchSizes {
			b.Run(fmt.Sprintf("%s/Batch%d", provider.Name(), batchSize), func(b *testing.B) {
				b.StopTimer()

				for i := 0; i < b.N; i++ {
					require.NoError(b, collection.Drop(ctx))

					iter := provider.NewIterator()

					for {
						docs, err := iterator.ConsumeValuesN(iter, batchSize)
						require.NoError(b, err)

						if docs == nil {
							break
						}

						insertDocs := make([]any, len(docs))
						for i := range insertDocs {
							insertDocs[i] = docs[i]
						}

						b.StartTimer()

						_, err = collection.InsertMany(ctx, insertDocs)
						require.NoError(b, err)

						b.StopTimer()
					}
				}
			})
		}
	}
}

func BenchmarkInsertManyIntoDifferentCollections(b *testing.B) {
	const numCollections = 25

	ctx, collection := setup.Setup(b)

	b.Log("got here 1")

	provider := shareddata.BenchmarkSettingsDocuments
	iter := provider.NewIterator()

	insertDocs := []any{}

	for {
		docs, err := iterator.ConsumeValues(iter)
		if errors.Is(err, iterator.ErrIteratorDone) || docs == nil { // why nil?
			break
		}

		require.NoError(b, err)

		for _, doc := range docs {
			insertDocs = append(insertDocs, doc)
		}
	}

	b.Log("got here 2")

	// TODO insert concurrently into each collection when it's working
	// var wg sync.WaitGroup
	// wg.Add(numCollections)
	for i := 0; i < numCollections; i++ {
		r := rune('a' + i)
		name := string(r)
		err := collection.Database().CreateCollection(ctx, name)
		require.NoError(b, err)

		b.StartTimer()
		collection.Database().Collection(name).InsertMany(ctx, insertDocs)
		b.StopTimer()
	}
}
