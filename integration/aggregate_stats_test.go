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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestAggregateCollStats(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	// call validate to updated statistics
	err := collection.Database().RunCommand(ctx, bson.D{{"validate", collection.Name()}}).Err()
	require.NoError(t, err)

	pipeline := bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{}}}}}}

	cursor, err := collection.Aggregate(ctx, pipeline)
	require.NoError(t, err)

	res := FetchAll(t, ctx, cursor)
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
	assert.Zero(t, must.NotFail(storageStats.Get("freeStorageSize")))
	assert.Equal(t, false, must.NotFail(storageStats.Get("capped")))
	assert.NotZero(t, must.NotFail(storageStats.Get("nindexes")))
	assert.NotZero(t, must.NotFail(storageStats.Get("totalIndexSize")))
	assert.NotZero(t, must.NotFail(storageStats.Get("totalSize")))
	assert.NotZero(t, must.NotFail(storageStats.Get("indexSizes")))
	assert.Equal(t, int32(1), must.NotFail(storageStats.Get("scaleFactor")))

	cappedCollectionName := testutil.CollectionName(t) + "capped"
	opts := options.CreateCollection().SetCapped(true).SetSizeInBytes(1000).SetMaxDocuments(10)
	err = collection.Database().CreateCollection(ctx, cappedCollectionName, opts)
	require.NoError(t, err)

	cappedCollection := collection.Database().Collection(cappedCollectionName)
	cursor, err = cappedCollection.Aggregate(ctx, pipeline)
	require.NoError(t, err)

	res = FetchAll(t, ctx, cursor)
	require.Len(t, res, 1)
	v, _ = ConvertDocument(t, res[0]).Get("storageStats")
	require.NotNil(t, v)

	storageStats, ok = v.(*types.Document)
	require.True(t, ok)
	assert.Equal(t, true, must.NotFail(storageStats.Get("capped")))
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
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

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

func TestAggregateCollStatsCommandIndexSizes(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t, shareddata.DocumentsStrings)

	cursorNoScale, err := collection.Aggregate(ctx, bson.A{
		bson.D{{"$collStats", bson.D{{"storageStats", bson.D{}}}}},
	})
	require.NoError(t, err)

	defer cursorNoScale.Close(ctx)

	scale := int32(1000)
	cursor, err := collection.Aggregate(ctx, bson.A{
		bson.D{{"$collStats", bson.D{{"storageStats", bson.D{{"scale", scale}}}}}},
	})
	require.NoError(t, err)

	defer cursor.Close(ctx)

	resNoScale := FetchAll(t, ctx, cursorNoScale)
	require.Equal(t, 1, len(resNoScale))

	res := FetchAll(t, ctx, cursor)
	require.Equal(t, 1, len(res))

	docNoScale := ConvertDocument(t, resNoScale[0])
	doc := ConvertDocument(t, res[0])

	storageStatsNoScale := must.NotFail(docNoScale.Get("storageStats")).(*types.Document)
	storageStats := must.NotFail(doc.Get("storageStats")).(*types.Document)

	size := must.NotFail(storageStats.Get("size"))
	switch sizeNoScale := must.NotFail(storageStatsNoScale.Get("size")).(type) {
	case int32:
		require.EqualValues(t, sizeNoScale/scale, size)
	case int64:
		require.EqualValues(t, sizeNoScale/int64(scale), size)
	default:
		t.Fatalf("unknown type %v", sizeNoScale)
	}

	avgObjSizeNoScale := must.NotFail(storageStatsNoScale.Get("avgObjSize"))
	avgObjSize := must.NotFail(storageStats.Get("avgObjSize"))
	require.EqualValues(t, avgObjSizeNoScale, avgObjSize)

	storageSize := must.NotFail(storageStats.Get("storageSize"))
	switch sizeNoScale := must.NotFail(storageStatsNoScale.Get("storageSize")).(type) {
	case int32:
		require.EqualValues(t, sizeNoScale/scale, storageSize)
	case int64:
		require.EqualValues(t, sizeNoScale/int64(scale), storageSize)
	default:
		t.Fatalf("unknown type %v", sizeNoScale)
	}

	freeStorageSize := must.NotFail(storageStats.Get("freeStorageSize"))
	switch sizeNoScale := must.NotFail(storageStatsNoScale.Get("freeStorageSize")).(type) {
	case int32:
		require.EqualValues(t, sizeNoScale/scale, freeStorageSize)
	case int64:
		require.EqualValues(t, sizeNoScale/int64(scale), freeStorageSize)
	default:
		t.Fatalf("unknown type %v", sizeNoScale)
	}

	totalIndexSize := must.NotFail(storageStats.Get("totalIndexSize"))
	switch sizeNoScale := must.NotFail(storageStatsNoScale.Get("totalIndexSize")).(type) {
	case int32:
		require.EqualValues(t, sizeNoScale/scale, totalIndexSize)
	case int64:
		require.EqualValues(t, sizeNoScale/int64(scale), totalIndexSize)
	default:
		t.Fatalf("unknown type %v", sizeNoScale)
	}

	totalSize := must.NotFail(storageStats.Get("totalSize"))
	switch sizeNoScale := must.NotFail(storageStatsNoScale.Get("totalSize")).(type) {
	case int32:
		require.EqualValues(t, sizeNoScale/scale, totalSize)
	case int64:
		require.EqualValues(t, sizeNoScale/int64(scale), totalSize)
	default:
		t.Fatalf("unknown type %v", sizeNoScale)
	}

	indexSizesNoScale := must.NotFail(storageStatsNoScale.Get("indexSizes")).(*types.Document)
	indexSizes := must.NotFail(storageStats.Get("indexSizes")).(*types.Document)

	require.Equal(t, []string{"_id_"}, indexSizesNoScale.Keys())
	require.Equal(t, []string{"_id_"}, indexSizes.Keys())

	for _, index := range indexSizesNoScale.Keys() {
		size := must.NotFail(indexSizes.Get(index))
		switch sizeNoScale := must.NotFail(indexSizesNoScale.Get(index)).(type) {
		case int32:
			require.EqualValues(t, sizeNoScale/scale, size)
		case int64:
			require.EqualValues(t, sizeNoScale/int64(scale), size)
		default:
			t.Fatalf("unknown type %v", sizeNoScale)
		}
	}
}
