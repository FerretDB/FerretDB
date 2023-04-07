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

	"github.com/FerretDB/FerretDB/integration/benchmarkdata"
	"github.com/FerretDB/FerretDB/integration/setup"
)

func BenchmarkQuery(b *testing.B) {
	s := setup.SetupBenchmark(b, benchmarkdata.SimpleData)
	ctx := s.Ctx

	coll := s.TargetCollection
	collNoPushdown := s.TargetNoPushdownCollection
	collCompat := s.CompatCollection

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
		b.Run(name, func(b *testing.B) {
			b.Run("Pushdown", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					cur, err := coll.Find(ctx, bm.filter)
					require.NoError(b, err)

					var res []bson.D
					require.NoError(b, cur.All(ctx, &res))
				}
			})
			b.Run("NoPushdown", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					cur, err := collNoPushdown.Find(ctx, bm.filter)
					require.NoError(b, err)

					var res []bson.D
					require.NoError(b, cur.All(ctx, &res))
				}
			})
			b.Run("Compat", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					cur, err := collCompat.Find(ctx, bm.filter)
					require.NoError(b, err)

					var res []bson.D
					require.NoError(b, cur.All(ctx, &res))
				}
			})
		})
	}
}

func BenchmarkReplaceOne(b *testing.B) {
	s := setup.SetupBenchmark(b, benchmarkdata.LargeDocument)
	ctx := s.Ctx

	coll := s.TargetCollection
	collCompat := s.CompatCollection

	for name, bm := range map[string]struct {
		filter bson.D
	}{
		"NoFilter": { // there's only ever one document to replace.
			filter: bson.D{},
		},
	} {
		b.Run(name, func(b *testing.B) {
			b.Run("AlterFourElements", func(b *testing.B) {
				// TODO create issue, we alter _id which is immutable
				for i := 0; i < b.N; i++ {
					res := bson.D{}
					err := coll.FindOne(ctx, bm.filter).Decode(&res)
					require.NoError(b, err)

					m := res.Map()
					m["1"] = "foo"
					m["2"] = "bar"
					m["3"] = "buz"
					m["4"] = "baz"
					replacement, err := bson.Marshal(m)
					require.NoError(b, err)

					ures, err := coll.ReplaceOne(ctx, bm.filter, replacement)
					require.NoError(b, err)
					require.Equal(b, int64(1), ures.ModifiedCount)
				}
			})
			b.Run("CompatAlterFourElements", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					res := bson.D{}
					err := collCompat.FindOne(ctx, bm.filter).Decode(&res)
					require.NoError(b, err)

					m := res.Map()
					m["1"] = "foo"
					m["2"] = "bar"
					m["3"] = "buz"
					m["4"] = "baz"
					replacement, err := bson.Marshal(m)
					require.NoError(b, err)

					ures, err := collCompat.ReplaceOne(ctx, bm.filter, replacement)
					require.NoError(b, err)
					require.Equal(b, int64(1), ures.ModifiedCount)
				}
			})
		})
	}
}
