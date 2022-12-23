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
)

func BenchmarkPushdowns(b *testing.B) {
	ctx, coll := setup.Setup(b)

	res, err := coll.InsertOne(ctx, bson.D{{}})
	require.NoError(b, err)

	id := res.InsertedID

	// run benchmark with query pushdown
	b.Run("ObjectID", func(b *testing.B) {
		cur, err := coll.Find(ctx, bson.D{{"_id", id}})
		require.NoError(b, err)

		var res []bson.D
		err = cur.All(ctx, &res)
		require.NoError(b, err)

		b.Log(res)
	})

	// run benchmark without query pushdown
	b.Run("NoPushdown", func(b *testing.B) {
		cur, err := coll.Find(ctx, bson.D{{"_id", "test"}})
		require.NoError(b, err)

		var res []bson.D
		err = cur.All(ctx, &res)
		require.NoError(b, err)

		b.Log(res)
	})
}
