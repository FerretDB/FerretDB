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

func BenchmarkPushdowns(b *testing.B) {
	ctx, coll := setup.Setup(b, shareddata.AllProviders()...)

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
