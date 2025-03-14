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

	"github.com/FerretDB/FerretDB/v2/internal/util/xiter"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

func BenchmarkFind(b *testing.B) {
	for _, provider := range shareddata.AllBenchmarkProviders() {
		b.Run(provider.Name(), func(b *testing.B) {
			s := setup.SetupWithOpts(b, &setup.SetupOpts{
				BenchmarkProvider: provider,
			})

			for name, filter := range map[string]bson.D{
				"Int32IDIndex":         {{"_id", int32(42)}},
				"Int32One":             {{"id", int32(42)}},
				"Int32Many":            {{"v", int32(42)}},
				"Int32ManyDotNotation": {{"v.foo", int32(42)}},
			} {
				if provider == shareddata.BenchSettings && name != "Int32IDIndex" {
					continue
				}

				b.Run(name, func(b *testing.B) {
					var firstDocs, docs int

					for b.Loop() {
						cursor, err := s.Collection.Find(s.Ctx, filter)
						if err != nil {
							b.Fatal(err)
						}

						docs = 0
						for cursor.Next(s.Ctx) {
							docs++
						}

						if err = cursor.Err(); err != nil {
							b.Fatal(err)
						}

						if err = cursor.Close(s.Ctx); err != nil {
							b.Fatal(err)
						}

						if firstDocs == 0 {
							firstDocs = docs
						}
					}

					require.Positive(b, firstDocs)
					require.Equal(b, firstDocs, docs)

					b.ReportMetric(float64(docs), "docs-returned")
				})
			}
		})
	}
}

func BenchmarkInsert(b *testing.B) {
	for _, provider := range shareddata.AllBenchmarkProviders() {
		var total int
		for range provider.Docs() {
			total++
		}

		var batchSizes []int
		for _, batchSize := range []int{1, 10, 100, 1000} {
			if batchSize <= total {
				batchSizes = append(batchSizes, batchSize)
			}
		}

		for _, batchSize := range batchSizes {
			b.Run(fmt.Sprintf("%s/Batch%d", provider.Name(), batchSize), func(b *testing.B) {
				ctx, collection := setup.Setup(b)

				for b.Loop() {
					if err := collection.Drop(ctx); err != nil {
						b.Fatal(err)
					}

					for docs := range xiter.Chunk(provider.Docs(), batchSize) {
						if _, err := collection.InsertMany(ctx, docs); err != nil {
							b.Fatal(err)
						}
					}
				}
			})
		}
	}
}
