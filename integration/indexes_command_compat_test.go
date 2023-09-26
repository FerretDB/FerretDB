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
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// TestCreateIndexesCommandCompat tests specific behavior for index creation that can be only provided through RunCommand.
func TestCreateIndexesCommandCompat(t *testing.T) {
	t.Parallel()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)
	targetCollection := targetCollections[0]
	compatCollection := compatCollections[0]

	for name, tc := range map[string]struct { //nolint:vet // for readability
		collectionName any
		indexName      any
		key            any
		unique         any
		resultType     compatTestCaseResultType // defaults to nonEmptyResult

		skip         string // optional, skip test with a specified reason
		skipForOldPg string // optional, skip test for the old PG backend with a specified reason
	}{
		"InvalidCollectionName": {
			collectionName: 42,
			key:            bson.D{{"v", -1}},
			indexName:      "custom-name",
			resultType:     emptyResult,
		},
		"NilCollectionName": {
			collectionName: nil,
			key:            bson.D{{"v", -1}},
			indexName:      "custom-name",
			resultType:     emptyResult,
		},
		"EmptyCollectionName": {
			collectionName: "",
			key:            bson.D{{"v", -1}},
			indexName:      "custom-name",
			resultType:     emptyResult,
		},
		"IndexNameNotSet": {
			collectionName: "test",
			key:            bson.D{{"v", -1}},
			indexName:      nil,
			resultType:     emptyResult,
		},
		"EmptyIndexName": {
			collectionName: "test",
			key:            bson.D{{"v", -1}},
			indexName:      "",
			resultType:     emptyResult,
		},
		"NonStringIndexName": {
			collectionName: "test",
			key:            bson.D{{"v", -1}},
			indexName:      42,
			resultType:     emptyResult,
		},
		"ExistingNameDifferentKeyLength": {
			collectionName: "test",
			key:            bson.D{{"_id", 1}, {"v", 1}},
			indexName:      "_id_",                                             // the same name as the default index
			skipForOldPg:   "https://github.com/FerretDB/FerretDB/issues/3435", // old PG handler doesn't validate this case correctly
		},
		"InvalidKey": {
			collectionName: "test",
			key:            42,
			resultType:     emptyResult,
		},
		"EmptyKey": {
			collectionName: "test",
			key:            bson.D{},
			resultType:     emptyResult,
		},
		"KeyNotSet": {
			collectionName: "test",
			resultType:     emptyResult,
		},
		"UniqueFalse": {
			collectionName: "unique_false",
			key:            bson.D{{"v", 1}},
			indexName:      "unique_false",
			unique:         false,
		},
		"UniqueTypeDocument": {
			collectionName: "test",
			key:            bson.D{{"v", 1}},
			indexName:      "test",
			unique:         bson.D{},
			resultType:     emptyResult,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			if tc.skipForOldPg != "" {
				t.Skipf("skipping for the old Postgres backend: %s", tc.skipForOldPg)
			}

			t.Helper()
			t.Parallel()

			indexesDoc := bson.D{}

			if tc.key != nil {
				indexesDoc = append(indexesDoc, bson.E{Key: "key", Value: tc.key})
			}

			if tc.indexName != nil {
				indexesDoc = append(indexesDoc, bson.E{"name", tc.indexName})
			}

			if tc.unique != nil {
				indexesDoc = append(indexesDoc, bson.E{Key: "unique", Value: tc.unique})
			}

			var targetRes bson.D
			targetErr := targetCollection.Database().RunCommand(ctx, bson.D{
				{"createIndexes", tc.collectionName},
				{"indexes", bson.A{indexesDoc}},
			}).Decode(&targetRes)

			var compatRes bson.D
			compatErr := compatCollection.Database().RunCommand(ctx, bson.D{
				{"createIndexes", tc.collectionName},
				{"indexes", bson.A{indexesDoc}},
			}).Decode(&compatRes)

			if targetErr != nil {
				t.Logf("Target error: %v", targetErr)
				t.Logf("Compat error: %v", compatErr)

				// error messages are intentionally not compared
				AssertMatchesCommandError(t, compatErr, targetErr)

				return
			}
			require.NoError(t, compatErr, "compat error; target returned no error")

			if tc.resultType == emptyResult {
				require.Nil(t, targetRes)
				require.Nil(t, compatRes)
			}

			assert.Equal(t, compatRes, targetRes)

			targetCursor, targetErr := targetCollection.Indexes().List(ctx)
			compatCursor, compatErr := compatCollection.Indexes().List(ctx)

			if targetCursor != nil {
				defer targetCursor.Close(ctx)
			}
			if compatCursor != nil {
				defer compatCursor.Close(ctx)
			}

			require.NoError(t, targetErr)
			require.NoError(t, compatErr)

			targetListRes := FetchAll(t, ctx, targetCursor)
			compatListRes := FetchAll(t, ctx, compatCursor)

			assert.Equal(t, compatListRes, targetListRes)

			targetSpec, targetErr := targetCollection.Indexes().ListSpecifications(ctx)
			compatSpec, compatErr := compatCollection.Indexes().ListSpecifications(ctx)

			require.NoError(t, compatErr)
			require.NoError(t, targetErr)

			assert.Equal(t, compatSpec, targetSpec)
		})
	}
}

// TestCreateIndexesCommandCompatCheckFields check that the response contains response's fields
// such as numIndexBefore, numIndexAfter, createdCollectionAutomatically
// contain the correct values.
func TestCreateIndexesCommandCompatCheckFields(tt *testing.T) {
	tt.Parallel()

	t := setup.FailsForSQLite(tt, "https://github.com/FerretDB/FerretDB/issues/3331")

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)
	targetCollection := targetCollections[0]
	compatCollection := compatCollections[0]

	// Create an index for a non-existent collection, expect createdCollectionAutomatically to be true.
	collectionName := "newCollection"
	indexesDoc := bson.D{{"key", bson.D{{"v", 1}}}, {"name", "v_1"}}

	var targetRes bson.D
	targetErr := targetCollection.Database().RunCommand(ctx, bson.D{
		{"createIndexes", collectionName},
		{"indexes", bson.A{indexesDoc}},
	}).Decode(&targetRes)

	var compatRes bson.D
	compatErr := compatCollection.Database().RunCommand(ctx, bson.D{
		{"createIndexes", collectionName},
		{"indexes", bson.A{indexesDoc}},
	}).Decode(&compatRes)

	require.NoError(t, compatErr)
	require.NoError(t, targetErr, "target error; compat returned no error")

	compatDoc := ConvertDocument(t, compatRes)
	createdCollectionAutomatically, err := compatDoc.Get("createdCollectionAutomatically")
	require.NoError(t, err)
	require.True(t, createdCollectionAutomatically.(bool)) // must be true because a new collection was created

	AssertEqualDocuments(t, compatRes, targetRes)

	// Now this collection exists, so we create another index and expect createdCollectionAutomatically to be false.
	indexesDoc = bson.D{{"key", bson.D{{"foo", 1}}}, {"name", "foo_1"}}

	targetErr = targetCollection.Database().RunCommand(ctx, bson.D{
		{"createIndexes", collectionName},
		{"indexes", bson.A{indexesDoc}},
	}).Decode(&targetRes)

	compatErr = compatCollection.Database().RunCommand(ctx, bson.D{
		{"createIndexes", collectionName},
		{"indexes", bson.A{indexesDoc}},
	}).Decode(&compatRes)

	require.NoError(t, compatErr)
	require.NoError(t, targetErr, "target error; compat returned no error")

	compatDoc = ConvertDocument(t, compatRes)
	createdCollectionAutomatically, err = compatDoc.Get("createdCollectionAutomatically")
	require.NoError(t, err)
	require.False(t, createdCollectionAutomatically.(bool)) // must be false because the collection already exists

	AssertEqualDocuments(t, compatRes, targetRes)

	// Call index creation for the index that already exists, expect note to be set.
	indexesDoc = bson.D{{"key", bson.D{{"foo", 1}}}, {"name", "foo_1"}}

	targetErr = targetCollection.Database().RunCommand(ctx, bson.D{
		{"createIndexes", collectionName},
		{"indexes", bson.A{indexesDoc}},
	}).Decode(&targetRes)

	compatErr = compatCollection.Database().RunCommand(ctx, bson.D{
		{"createIndexes", collectionName},
		{"indexes", bson.A{indexesDoc}},
	}).Decode(&compatRes)

	require.NoError(t, compatErr)
	require.NoError(t, targetErr, "target error; compat returned no error")

	compatDoc = ConvertDocument(t, compatRes)
	createdCollectionAutomatically, err = compatDoc.Get("note")
	require.NoError(t, err)

	// note must be set because no new indexes were created:
	require.Equal(t, "all indexes already exist", createdCollectionAutomatically.(string))

	AssertEqualDocuments(t, compatRes, targetRes)
}

func TestDropIndexesCommandCompat(tt *testing.T) {
	tt.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		toCreate []mongo.IndexModel // optional, if set, create the given indexes before drop is called
		toDrop   any                // required, index to drop

		resultType compatTestCaseResultType // optional, defaults to nonEmptyResult

		skip string // optional, skip test with a specified reason
	}{
		"MultipleIndexesByName": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"v", 1}, {"foo", 1}}},
				{Keys: bson.D{{"v.foo", -1}}},
			},
			toDrop: bson.A{"v_-1", "v_1_foo_1"},
		},
		"MultipleIndexesByKey": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"v.foo", -1}}},
			},
			toDrop:     bson.A{bson.D{{"v", -1}}, bson.D{{"v.foo", -1}}},
			resultType: emptyResult,
		},
		"NonExistentMultipleIndexes": {
			toDrop:     bson.A{"non-existent", "invalid"},
			resultType: emptyResult,
		},
		"MultipleIndexesWithDefault": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", 1}}},
			},
			toDrop:     bson.A{"v_1", "_id_"},
			resultType: emptyResult,
		},
		"InvalidMultipleIndexType": {
			toDrop:     bson.A{1},
			resultType: emptyResult,
		},
		"DocumentIndex": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
			},
			toDrop: bson.D{{"v", -1}},
		},
		"SimilarIndexes": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", 1}, {"foo", 1}}},
				{Keys: bson.D{{"v", 1}, {"bar", 1}}},
			},
			toDrop: bson.D{{"v", 1}, {"bar", 1}},
		},
		"DropAllExpression": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"foo.bar", 1}}},
				{Keys: bson.D{{"foo", 1}, {"bar", 1}}},
			},
			toDrop: "*",
		},
		"WrongExpression": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"foo.bar", 1}}},
				{Keys: bson.D{{"foo", 1}, {"bar", 1}}},
			},
			toDrop:     "***",
			resultType: emptyResult,
		},
		"NonExistentDescendingID": {
			toDrop:     bson.D{{"_id", -1}},
			resultType: emptyResult,
		},
		"MultipleKeyIndex": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"_id", -1}, {"v", 1}}},
			},
			toDrop: bson.D{
				{"_id", -1},
				{"v", 1},
			},
		},
	} {
		name, tc := name, tc
		tt.Run(name, func(tt *testing.T) {
			if tc.skip != "" {
				tt.Skip(tc.skip)
			}

			tt.Helper()
			tt.Parallel()

			require.NotNil(tt, tc.toDrop, "toDrop must not be nil")

			// It's enough to use a single provider for drop indexes test as indexes work the same for different collections.
			s := setup.SetupCompatWithOpts(tt, &setup.SetupCompatOpts{
				Providers:                []shareddata.Provider{shareddata.Composites},
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				tt.Run(targetCollection.Name(), func(tt *testing.T) {
					tt.Helper()

					var t testtb.TB = tt

					if tc.toCreate != nil {
						_, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.toCreate)
						_, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.toCreate)
						require.NoError(t, compatErr)
						require.NoError(t, targetErr)

						// List indexes to see they are identical after creation.
						targetCursor, targetListErr := targetCollection.Indexes().List(ctx)
						compatCursor, compatListErr := compatCollection.Indexes().List(ctx)

						if targetCursor != nil {
							defer targetCursor.Close(ctx)
						}
						if compatCursor != nil {
							defer compatCursor.Close(ctx)
						}

						require.NoError(t, targetListErr)
						require.NoError(t, compatListErr)

						targetList := FetchAll(t, ctx, targetCursor)
						compatList := FetchAll(t, ctx, compatCursor)

						require.Equal(t, compatList, targetList)
					}

					targetCommand := bson.D{
						{"dropIndexes", targetCollection.Name()},
						{"index", tc.toDrop},
					}

					compatCommand := bson.D{
						{"dropIndexes", compatCollection.Name()},
						{"index", tc.toDrop},
					}

					var targetRes bson.D
					targetErr := targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetRes)

					var compatRes bson.D
					compatErr := compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatRes)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)

						// error messages are intentionally not compared
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					if tc.resultType == emptyResult {
						require.Nil(t, targetRes)
						require.Nil(t, compatRes)
					}

					require.Equal(t, compatRes, targetRes)

					if compatErr == nil {
						nonEmptyResults = true
					}

					// List indexes to see they are identical after deletion.
					targetCursor, targetListErr := targetCollection.Indexes().List(ctx)
					compatCursor, compatListErr := compatCollection.Indexes().List(ctx)

					if targetCursor != nil {
						defer targetCursor.Close(ctx)
					}
					if compatCursor != nil {
						defer compatCursor.Close(ctx)
					}

					require.NoError(t, targetListErr)
					require.NoError(t, compatListErr)

					targetList := FetchAll(t, ctx, targetCursor)
					compatList := FetchAll(t, ctx, compatCursor)

					assert.Equal(t, compatList, targetList)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				require.True(tt, nonEmptyResults, "expected non-empty results (some indexes should be deleted)")
			case emptyResult:
				require.False(tt, nonEmptyResults, "expected empty results (no indexes should be deleted)")
			default:
				tt.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}
