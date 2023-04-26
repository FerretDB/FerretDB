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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func BenchmarkQuery(b *testing.B) {
	provider := shareddata.MixedBenchmarkValues

	s := setup.SetupWithOpts(b, &setup.SetupOpts{
		BenchmarkProvider: &provider,
	})

	ctx, coll := s.Ctx, s.Collection

	for name, bm := range map[string]struct {
		filter bson.D
	}{
		"String": {
			filter: bson.D{{"v", "foo"}},
		},
		"DotNotation": {
			filter: bson.D{{"v.42", "hello"}},
		},
	} {
		b.Run(fmt.Sprint(name, "_", provider.Hash()), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				cur, err := coll.Find(ctx, bm.filter)
				require.NoError(b, err)

				var res []bson.D
				require.NoError(b, cur.All(ctx, &res))
			}
		})
	}
}

func BenchmarkReplaceLargeDocument(b *testing.B) {
	provider := shareddata.LargeDocumentBenchmarkValues

	s := setup.SetupWithOpts(b, &setup.SetupOpts{
		BenchmarkProvider: &provider,
	})
	ctx, coll := s.Ctx, s.Collection

	filter := bson.D{{"_id", 0}}
	runsCount := 1

	b.Run(provider.Hash(), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var doc bson.D
			err := coll.FindOne(ctx, filter).Decode(&doc)
			require.NoError(b, err)

			doc[runsCount].Value = i * 11111

			updateRes, err := coll.ReplaceOne(ctx, filter, doc)
			require.NoError(b, err)

			require.Equal(b, int64(1), updateRes.ModifiedCount)
		}
		runsCount++
	})
}

func BenchmarkInsertMany(b *testing.B) {
	id := 0

	ctx, coll := setup.Setup(b)
	db := coll.Database()

	b.Run("InsertMany-D10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StartTimer()
			for j := 0; j < 40; j++ {
				_, err := coll.InsertMany(ctx, []any{
					bson.D{{"_id", id}, {"test", "test1"}},
					bson.D{{"_id", (id + 1)}, {"test", "test2"}},
					bson.D{{"_id", (id + 2)}, {"test", "test3"}},
					bson.D{{"_id", (id + 3)}, {"test", "test4"}},
					bson.D{{"_id", (id + 4)}, {"test", "test5"}},
					bson.D{{"_id", (id + 5)}, {"test", "test6"}},
					bson.D{{"_id", (id + 6)}, {"test", "test7"}},
					bson.D{{"_id", (id + 7)}, {"test", "test8"}},
					bson.D{{"_id", (id + 8)}, {"test", "test9"}},
					bson.D{{"_id", (id + 9)}, {"test", "test10"}},
					bson.D{{"_id", (id + 10)}, {"test", "test11"}},
					bson.D{{"_id", (id + 11)}, {"test", "test12"}},
					bson.D{{"_id", (id + 12)}, {"test", "test13"}},
					bson.D{{"_id", (id + 13)}, {"test", "test14"}},
					bson.D{{"_id", (id + 14)}, {"test", "test15"}},
				})
				require.NoError(b, err)
				id = id + 15
			}
			b.StopTimer()
			coll.Drop(ctx)
			coll = db.Collection("Benchmark/InsertMany-D10")
		}
	})
}
