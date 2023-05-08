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

func BenchmarkReplaceLargeDocument(b *testing.B) {
	provider := shareddata.BenchmarkLargeDocuments

	s := setup.SetupWithOpts(b, &setup.SetupOpts{
		BenchmarkProvider: provider,
	})
	ctx, coll := s.Ctx, s.Collection

	runsCount := 1

	b.Run(provider.Name(), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var doc bson.D
			err := coll.FindOne(ctx, bson.D{}).Decode(&doc)
			require.NoError(b, err)

			doc[runsCount].Value = i * 11111

			updateRes, err := coll.ReplaceOne(ctx, bson.D{}, doc)
			require.NoError(b, err)

			require.Equal(b, int64(1), updateRes.ModifiedCount)
		}
		runsCount++
	})
}

func BenchmarkInsertMany(b *testing.B) {
	ctx, coll := setup.Setup(b)
	db := coll.Database()

	provider := shareddata.BenchmarkLessSmallDocuments

	b.Run(provider.Name(), func(b *testing.B) {
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			iter := provider.NewIterator()

			for {
				docs, err := iterator.ConsumeValuesN(iter, 30)
				require.NoError(b, err)

				if docs == nil {
					break
				}

				insertDocs := make([]any, len(docs))
				for i := range insertDocs {
					insertDocs[i] = docs[i]
				}

				b.StartTimer()

				_, err = coll.InsertMany(ctx, insertDocs)
				require.NoError(b, err)

				b.StopTimer()

				require.NoError(b, coll.Drop(ctx))
				coll = db.Collection(coll.Name())
			}
		}
	})
}
