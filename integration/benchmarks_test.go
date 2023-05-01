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
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/iterator"
)

func BenchmarkQuery(b *testing.B) {
	provider := shareddata.BenchmarkSmallDocuments

	b.Run(provider.Name(), func(b *testing.B) {
		s := setup.SetupWithOpts(b, &setup.SetupOpts{
			BenchmarkProvider: provider,
		})

		total, err := iterator.ConsumeCount(provider.NewIterator())
		require.NoError(b, err)

		for name, bc := range map[string]struct {
			filter bson.D
		}{
			"ID": {
				filter: bson.D{{"_id", int32(42)}},
			},
			"String": {
				filter: bson.D{{"v", "foo"}},
			},
			"DotNotation": {
				filter: bson.D{{"v.foo", int32(42)}},
			},
		} {
			b.Run(name, func(b *testing.B) {
				var firstDocs, docs []bson.D

				for i := 0; i < b.N; i++ {
					cursor, err := s.Collection.Find(s.Ctx, bc.filter)
					require.NoError(b, err)

					docs = FetchAll(b, s.Ctx, cursor)
					require.NotEmpty(b, docs)

					if firstDocs == nil {
						firstDocs = docs
					}
				}

				b.StopTimer()

				require.Len(b, docs, len(firstDocs))

				b.ReportMetric(float64(len(docs)), "docs-returned")
				b.ReportMetric(float64(total), "docs-total")
			})
		}
	})
}

// func BenchmarkReplaceLargeDocument(b *testing.B) {
// 	provider := shareddata.BenchmarkLargeDocuments

// 	s := setup.SetupWithOpts(b, &setup.SetupOpts{
// 		BenchmarkProvider: provider,
// 	})
// 	ctx, coll := s.Ctx, s.Collection

// 	filter := bson.D{{"_id", 0}}
// 	runsCount := 1

// 	b.Run(provider.Hash(), func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			var doc bson.D
// 			err := coll.FindOne(ctx, filter).Decode(&doc)
// 			require.NoError(b, err)

// 			doc[runsCount].Value = i * 11111

// 			updateRes, err := coll.ReplaceOne(ctx, filter, doc)
// 			require.NoError(b, err)

// 			require.Equal(b, int64(1), updateRes.ModifiedCount)
// 		}
// 		runsCount++
// 	})
// }

// func BenchmarkInsertMany(b *testing.B) {
// 	id := 0

// 	ctx, coll := setup.Setup(b)
// 	db := coll.Database()

// 	b.Run("InsertMany-D10", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			b.StartTimer()
// 			for j := 0; j < 40; j++ {
// 				_, err := coll.InsertMany(ctx, []any{
// 					bson.D{{"_id", id}, {"test", "test1"}},
// 					bson.D{{"_id", (id + 1)}, {"test", "test2"}},
// 					bson.D{{"_id", (id + 2)}, {"test", "test3"}},
// 					bson.D{{"_id", (id + 3)}, {"test", "test4"}},
// 					bson.D{{"_id", (id + 4)}, {"test", "test5"}},
// 					bson.D{{"_id", (id + 5)}, {"test", "test6"}},
// 					bson.D{{"_id", (id + 6)}, {"test", "test7"}},
// 					bson.D{{"_id", (id + 7)}, {"test", "test8"}},
// 					bson.D{{"_id", (id + 8)}, {"test", "test9"}},
// 					bson.D{{"_id", (id + 9)}, {"test", "test10"}},
// 					bson.D{{"_id", (id + 10)}, {"test", "test11"}},
// 					bson.D{{"_id", (id + 11)}, {"test", "test12"}},
// 					bson.D{{"_id", (id + 12)}, {"test", "test13"}},
// 					bson.D{{"_id", (id + 13)}, {"test", "test14"}},
// 					bson.D{{"_id", (id + 14)}, {"test", "test15"}},
// 				})
// 				require.NoError(b, err)
// 				id = id + 15
// 			}
// 			b.StopTimer()
// 			require.NoError(b, coll.Drop(ctx))
// 			coll = db.Collection(coll.Name())
// 		}
// 	})
// }
