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
)

type QueryBenchmarkCase struct {
	filter bson.D
}

func BenchmarkFoo(b *testing.B) {
	ctx, coll, collNoPushdown, compatColl := setup.SetupBenchmark(b, setup.SimpleData)

	for name, bm := range map[string]QueryBenchmarkCase{
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
					cur, err := compatColl.Find(ctx, bm.filter)
					require.NoError(b, err)

					var res []bson.D
					require.NoError(b, cur.All(ctx, &res))
				}
			})
		})
	}
}

func BenchmarkLargeReplace(b *testing.B) {
	for {
		b.Run("insert", func(b *testing.B) {
			// insert the same data all the time
			// TODO consider running benchmark only once
		})
		b.Run("Pushdown", func(b *testing.B) {
			// find and replace
		})
		b.Run("NoPushdown", func(b *testing.B) {
			// find and replace
		})
	}
}

func BenchmarkPushdowns(b *testing.B) {
	s := setup.SetupWithOpts(b, &setup.SetupOpts{
		DatabaseName:   b.Name(),
		CollectionName: b.Name(),
		Providers:      []shareddata.Provider{shareddata.Scalars},
	})

	ctx, coll := s.Ctx, s.Collection

	res, err := coll.InsertOne(ctx, bson.D{{}})
	require.NoError(b, err)

	id := res.InsertedID

	b.Run("ObjectID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cur, err := coll.Find(ctx, bson.D{{"_id", id}})
			require.NoError(b, err)

			var res []bson.D
			err = cur.All(ctx, &res)
			require.NoError(b, err)

			require.NotEmpty(b, res)
		}
	})

	b.Run("StringID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cur, err := coll.Find(ctx, bson.D{{"_id", "string"}})
			require.NoError(b, err)

			var res []bson.D
			err = cur.All(ctx, &res)
			require.NoError(b, err)

			require.NotEmpty(b, res)
		}
	})

	b.Run("NoPushdown", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cur, err := coll.Find(ctx, bson.D{{"v", bson.D{{"$eq", 42.0}}}})
			require.NoError(b, err)

			var res []bson.D
			err = cur.All(ctx, &res)
			require.NoError(b, err)

			require.NotEmpty(b, res)
		}
	})
}
