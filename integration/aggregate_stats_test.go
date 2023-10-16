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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestAggregateCollStats(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	pipeline := bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{}}}}}}

	cursor, err := collection.Aggregate(ctx, pipeline)
	require.NoError(t, err)

	var res []bson.D
	err = cursor.All(ctx, &res)
	require.NoError(t, err)

	require.Len(t, res, 1)
	doc := ConvertDocument(t, res[0])

	assert.Equal(t, collection.Database().Name()+"."+collection.Name(), must.NotFail(doc.Get("ns")))

	v, _ := doc.Get("storageStats")
	require.NotNil(t, v)

	storageStats, ok := v.(*types.Document)
	require.True(t, ok)

	assert.NotZero(t, must.NotFail(storageStats.Get("size")))
	assert.NotZero(t, must.NotFail(storageStats.Get("count")))
	assert.NotZero(t, must.NotFail(storageStats.Get("avgObjSize")))
	assert.NotZero(t, must.NotFail(storageStats.Get("storageSize")))
	// TODO https://github.com/FerretDB/FerretDB/issues/2447
	// assert.NotZero(t, must.NotFail(storageStats.Get("freeStorageSize")))
	assert.Equal(t, false, must.NotFail(storageStats.Get("capped")))
	assert.NotZero(t, must.NotFail(storageStats.Get("nindexes")))
	assert.NotZero(t, must.NotFail(storageStats.Get("totalIndexSize")))
	assert.NotZero(t, must.NotFail(storageStats.Get("totalSize")))
	assert.NotZero(t, must.NotFail(storageStats.Get("indexSizes")))
	// TODO https://github.com/FerretDB/FerretDB/issues/2447
	// assert.Equal(t, int32(1), must.NotFail(storageStats.Get("scaleFactor")))
}

func TestAggregateCollStatsCommandErrors(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)

	for name, tc := range map[string]struct { //nolint:vet // used for test only
		command  bson.D          // required, command to run
		database *mongo.Database // defaults to collection.Database()

		err        *mongo.CommandError // required
		altMessage string              // optional, alternative error message
		skip       string              // optional, skip test with a specified reason
	}{
		"NonExistentDatabase": {
			database: collection.Database().Client().Database("non-existent"),
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{}}}}}}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code: 26,
				Name: "NamespaceNotFound",
				Message: `PlanExecutor error during aggregation :: caused by :: ` +
					`Unable to retrieve storageStats in $collStats stage :: caused by :: ` +
					`Collection [non-existent.TestAggregateCollStatsCommandErrors] not found.`,
			},
			altMessage: "ns not found: non-existent.TestAggregateCollStatsCommandErrors",
		},
		"NonExistentCollection": {
			command: bson.D{
				{"aggregate", "non-existent"},
				{"pipeline", bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{}}}}}}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code: 26,
				Name: "NamespaceNotFound",
				Message: `PlanExecutor error during aggregation :: caused by :: ` +
					`Unable to retrieve storageStats in $collStats stage :: caused by :: ` +
					`Collection [TestAggregateCollStatsCommandErrors.non-existent] not found.`,
			},
			altMessage: "ns not found: TestAggregateCollStatsCommandErrors.non-existent",
		},
		"NilCollStats": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{bson.D{{"$collStats", nil}}}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code:    5447000,
				Name:    "Location5447000",
				Message: `$collStats must take a nested object but found: $collStats: null`,
			},
			altMessage: `$collStats must take a nested object but found: { $collStats: null }`,
		},
		"StorageStatsNegativeScale": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{{"scale", -1000}}}}}}}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code:    51024,
				Name:    "Location51024",
				Message: `BSON field 'scale' value must be >= 1, actual value '-1000'`,
			},
		},
		"StorageStatsInvalidScale": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{{"scale", "invalid"}}}}}}}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: `BSON field '$collStats.storageStats.scale' is the wrong type 'string', expected types '[long, int, decimal, double']`,
			},
			altMessage: `BSON field '$collStats.storageStats.scale' is the wrong type 'string', expected types '[long, int, decimal, double]'`,
		},
		"CountCollStatsCount": {
			command: bson.D{
				{"aggregate", collection.Name()},
				{"pipeline", bson.A{
					bson.D{{"$count", "before"}},
					bson.D{{"$collStats", bson.D{{"count", bson.D{}}, {"storageStats", bson.D{}}}}},
					bson.D{{"$count", "after"}},
				}},
				{"cursor", bson.D{}},
			},
			err: &mongo.CommandError{
				Code:    40602,
				Name:    "Location40602",
				Message: `$collStats is only valid as the first stage in a pipeline`,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.err, "err must not be nil")

			db := tc.database
			if db == nil {
				db = collection.Database()
			}

			var res bson.D
			err := db.RunCommand(ctx, tc.command).Decode(&res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
			require.Nil(t, res)
		})
	}
}
