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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestIndexesDrop(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes are not supported for Tigris")

	t.Parallel()

	for name, tc := range map[string]struct {
		models        []mongo.IndexModel       // optional, if not nil create indexes before dropping
		dropIndexName string                   // name of a single index to drop
		dropAll       bool                     // set true for drop all indexes, if true dropIndexName must be empty.
		altErrorMsg   string                   // optional, alternative error message in case of error
		resultType    compatTestCaseResultType // defaults to nonEmptyResult
		skip          string                   // optional, skip test with a specified reason
	}{
		"DropAllCommand": {
			dropAll: true,
		},
		"ID": {
			dropIndexName: "_id_",
			resultType:    emptyResult,
		},
		"DescendingID": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"_id", -1}}},
			},
			dropIndexName: "_id_",
			resultType:    emptyResult,
		},
		"AscendingValue": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", 1}}},
			},
			dropIndexName: "v_1",
		},
		"DescendingValue": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
			},
			dropIndexName: "v_-1",
		},
		"DropAllExpression": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
			},
			dropIndexName: "*",
		},
		"DropByKey": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
			},
			dropIndexName: "v",
		},
		"DropByIDKey": {
			dropIndexName: "_id",
		},
		"NonExistent": {
			dropIndexName: "nonexistent_1",
			resultType:    emptyResult,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Helper()
			t.Parallel()

			if tc.dropAll {
				require.Empty(t, tc.dropIndexName, "index name must be empty when dropping all indexes")
			}

			// Use single provider for drop indexes test.
			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                []shareddata.Provider{shareddata.Composites},
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					if tc.models != nil {
						_, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.models)
						_, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.models)
						require.NoError(t, compatErr)
						require.NoError(t, targetErr)
					}

					var targetRes, compatRes bson.Raw
					var targetErr, compatErr error

					if tc.dropAll {
						targetRes, targetErr = targetCollection.Indexes().DropAll(ctx)
						compatRes, compatErr = compatCollection.Indexes().DropAll(ctx)
					} else {
						targetRes, targetErr = targetCollection.Indexes().DropOne(ctx, tc.dropIndexName)
						compatRes, compatErr = compatCollection.Indexes().DropOne(ctx, tc.dropIndexName)
					}

					if tc.altErrorMsg != "" {
						AssertMatchesCommandError(t, compatErr, targetErr)

						var expectedErr mongo.CommandError
						require.True(t, errors.As(compatErr, &expectedErr))
						expectedErr.Raw = nil
						AssertEqualAltError(t, expectedErr, tc.altErrorMsg, targetErr)
					} else {
						require.Equal(t, compatErr, targetErr)
					}

					require.Equal(t, compatRes, targetRes)

					if targetErr == nil {
						nonEmptyResults = true
					}

					// List indexes to see they are identical after creation.
					targetCur, targetErr := targetCollection.Indexes().List(ctx)
					compatCur, compatErr := compatCollection.Indexes().List(ctx)

					require.NoError(t, compatErr)
					require.Equal(t, compatErr, targetErr)

					targetIndexes := FetchAll(t, ctx, targetCur)
					compatIndexes := FetchAll(t, ctx, compatCur)

					require.Equal(t, compatIndexes, targetIndexes)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				require.True(t, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case emptyResult:
				require.False(t, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

func TestIndexesDropRunCommand(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes creation is not supported for Tigris")

	t.Parallel()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)
	targetCollection := targetCollections[0]
	compatCollection := compatCollections[0]

	for name, tc := range map[string]struct { //nolint:vet // for readability
		models         []mongo.IndexModel       // optional, if not nil create indexes before dropping
		collectionName string                   // collection name to use
		index          any                      // index name
		resultType     compatTestCaseResultType // defaults to nonEmptyResult
		skip           string                   // optional, skip test with a specified reason
		altErrorMsg    string                   // optional, alternative error message in case of error
	}{
		"InvalidType": {
			collectionName: targetCollection.Name(),
			index:          true,
			resultType:     emptyResult,
		},
		"NonExistentField": {
			collectionName: targetCollection.Name(),
			index:          "non-existent",
			resultType:     emptyResult,
		},
		"InvalidCollection": {
			collectionName: "non-existent",
			resultType:     emptyResult,
		},
		"MultipleIndexes": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"v.foo", -1}}},
			},
			collectionName: targetCollection.Name(),
			index:          bson.A{"v_1", "v_1_foo_1"},
		},
		"NonExistentMultipleIndexes": {
			collectionName: targetCollection.Name(),
			index:          bson.A{"non-existent", "invalid"},
		},
		"InvalidMultipleIndexType": {
			collectionName: targetCollection.Name(),
			index:          bson.A{1},
			resultType:     emptyResult,
		},
		"InvalidDocumentIndex": {
			collectionName: targetCollection.Name(),
			index:          bson.D{{"invalid", "invalid"}},
			resultType:     emptyResult,
		},
		"DocumentIndexValue": {
			collectionName: targetCollection.Name(),
			index:          bson.D{{"v", 1}},
		},
		"DocumentIndexID": {
			collectionName: targetCollection.Name(),
			index:          bson.D{{"_id", 1}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Helper()
			t.Parallel()

			if tc.models != nil {
				_, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.models)
				_, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.models)
				require.NoError(t, compatErr)
				require.NoError(t, targetErr)
			}

			command := bson.D{
				{"dropIndexes", tc.collectionName},
				{"index", tc.index},
			}

			var targetRes bson.D
			targetErr := targetCollection.Database().RunCommand(ctx, command).Decode(&targetRes)

			var compatRes bson.D
			compatErr := compatCollection.Database().RunCommand(ctx, command).Decode(&compatRes)

			if tc.altErrorMsg != "" {
				AssertMatchesCommandError(t, compatErr, targetErr)

				var expectedErr mongo.CommandError
				require.True(t, errors.As(compatErr, &expectedErr))
				expectedErr.Raw = nil
				AssertEqualAltError(t, expectedErr, tc.altErrorMsg, targetErr)
			} else {
				require.Equal(t, compatErr, targetErr)
			}

			if tc.resultType == emptyResult {
				require.Nil(t, targetRes)
				require.Nil(t, compatRes)
			}

			require.Equal(t, compatRes, targetRes)

			targetErr = targetCollection.Database().RunCommand(
				ctx, bson.D{{"listIndexes", tc.collectionName}},
			).Decode(&targetRes)

			compatErr = compatCollection.Database().RunCommand(
				ctx, bson.D{{"listIndexes", tc.collectionName}},
			).Decode(&compatRes)

			require.Equal(t, compatRes, targetRes)

			AssertMatchesCommandError(t, compatErr, targetErr)
		})
	}
}
