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

package oplog

import (
	"testing"

	"github.com/FerretDB/FerretDB/internal/util/must"

	"github.com/stretchr/testify/assert"

	"github.com/FerretDB/FerretDB/internal/types"

	"github.com/FerretDB/FerretDB/integration"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestOplogInsert(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t)

	_, err := coll.InsertOne(ctx, bson.D{{"foo", "bar"}})
	require.NoError(t, err)

	local := coll.Database().Client().Database("local")

	var lastOplogEntry bson.D

	opts := options.FindOne().SetSort(bson.D{{"$natural", -1}})
	err = local.Collection("oplog.rs").FindOne(ctx, bson.D{}, opts).Decode(&lastOplogEntry)
	require.NoError(t, err)

	expectedKeys := []string{"lsid", "txnNumber", "op", "ns", "ui", "o", "o2", "stmId", "ts", "t", "v", "wall", "prevOpTime"}

	actual := integration.ConvertDocument(t, lastOplogEntry)
	actualKeys := actual.Keys()

	require.ElementsMatch(t, expectedKeys, actualKeys)

	// Exact values might vary, so we just check types.
	assert.IsType(t, types.Document{}, must.NotFail(actual.Get("lsid")))
	assert.IsType(t, int64(0), must.NotFail(actual.Get("txnNumber")))
	assert.IsType(t, int32(0), must.NotFail(actual.Get("stmId")))
	assert.IsType(t, types.Timestamp(0), must.NotFail(actual.Get("ts")))
	assert.IsType(t, int64(0), must.NotFail(actual.Get("t")))
	assert.IsType(t, types.Timestamp(0), must.NotFail(actual.Get("wall")))
	assert.IsType(t, types.Timestamp(0), must.NotFail(actual.Get("prevOpTime")))

	actual.Remove("lsId")
	actual.Remove("txnNumber")  // transaction number
	actual.Remove("stmId")      // statement ID within transaction
	actual.Remove("ts")         // timestamp
	actual.Remove("t")          // term
	actual.Remove("wall")       // wall clock time
	actual.Remove("prevOpTime") // previous operation time

	// Exact values are known, so we check them.
	/*
			expected := types.NewDocument({
				{"op", "i"}, // operation - i, u, d, n, c
				{"ns", "test.coll"}, // namespace
		{"o", types.NewDocument({}),
		{"o2", types.NewDocument({}),
		{"v", 2}, // protocol version
			})
	*/
}
