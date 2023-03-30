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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestIndexesList(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                shareddata.AllProviders(),
		AddNonExistentCollection: true,
	})
	ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

	for i := range targetCollections {
		targetCollection := targetCollections[i]
		compatCollection := compatCollections[i]

		t.Run(targetCollection.Name(), func(t *testing.T) {
			t.Helper()
			t.Parallel()

			targetCur, targetErr := targetCollection.Indexes().List(ctx)
			compatCur, compatErr := compatCollection.Indexes().List(ctx)

			require.NoError(t, compatErr)
			assert.Equal(t, compatErr, targetErr)

			targetRes := FetchAll(t, ctx, targetCur)
			compatRes := FetchAll(t, ctx, compatCur)

			assert.Equal(t, compatRes, targetRes)
		})
	}
}

func TestIndexesCreate(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes creation is not supported for Tigris")

	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		models      []mongo.IndexModel
		altErrorMsg string                   // optional, alternative error message in case of error
		resultType  compatTestCaseResultType // defaults to nonEmptyResult
		skip        string                   // optional, skip test with a specified reason
	}{
		"empty": {
			models:     []mongo.IndexModel{},
			resultType: emptyResult,
		},
		"single-index": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
			},
		},
		"duplicate_id": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"_id", 1}}, // this index is already created by default
				},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2311",
		},
		"non-existent-field": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"field-does-not-exist", 1}}},
			},
		},
		"dot-notation": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v.foo", 1}}},
			},
		},
		"dangerous-key": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{
						{"v", 1},
						{"foo'))); DROP TABlE test._ferretdb_database_metadata; CREATE INDEX IF NOT EXISTS test ON test.test (((_jsonb->'foo", 1},
					},
				},
			},
		},
		"same-key": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}, {"v", 1}}},
			},
			resultType:  emptyResult,
			altErrorMsg: `Error in specification { v: -1, v: 1 }, the field "v" appears multiple times`,
		},
		"custom-name": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"foo", 1}, {"bar", -1}},
					Options: new(options.IndexOptions).SetName("custom-name"),
				},
			},
		},

		"multi-direction-different-indexes": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"v", 1}}},
			},
		},
		"multi-order": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"foo", -1}}},
				{Keys: bson.D{{"v", 1}}},
				{Keys: bson.D{{"bar", 1}}},
			},
		},
		"build-same-index": {
			models: []mongo.IndexModel{
				{Keys: bson.D{{"v", 1}}},
				{Keys: bson.D{{"v", 1}}},
			},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/2311",
		},
		"multi-with-invalid": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"foo", 1}, {"bar", 1}, {"v", -1}},
				},
				{
					Keys: bson.D{{"v", -1}, {"v", 1}},
				},
			},
			resultType:  emptyResult,
			altErrorMsg: `Error in specification { v: -1, v: 1 }, the field "v" appears multiple times`,
		},
		"same-key-different-names": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"v", -1}},
					Options: new(options.IndexOptions).SetName("foo"),
				},
				{
					Keys:    bson.D{{"v", -1}},
					Options: new(options.IndexOptions).SetName("bar"),
				},
			},
			resultType:  emptyResult,
			altErrorMsg: "One of the specified indexes already exists with a different name",
		},
		"same-name-different-keys": {
			models: []mongo.IndexModel{
				{
					Keys:    bson.D{{"foo", -1}},
					Options: new(options.IndexOptions).SetName("index-name"),
				},
				{
					Keys:    bson.D{{"bar", -1}},
					Options: new(options.IndexOptions).SetName("index-name"),
				},
			},
			resultType:  emptyResult,
			altErrorMsg: "One of the specified indexes already exists with a different key",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Helper()
			t.Parallel()

			// Use per-test setup because createIndexes modifies collection state,
			// however, we don't need to run index creation test for all the possible collections.
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

					targetRes, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.models)
					compatRes, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.models)

					if tc.altErrorMsg != "" {
						AssertMatchesCommandError(t, compatErr, targetErr)

						var expectedErr mongo.CommandError
						require.True(t, errors.As(compatErr, &expectedErr))
						expectedErr.Raw = nil
						AssertEqualAltError(t, expectedErr, tc.altErrorMsg, targetErr)
					} else {
						require.Equal(t, compatErr, targetErr)
					}

					assert.Equal(t, compatRes, targetRes)

					if compatErr == nil {
						nonEmptyResults = true
					}

					// List indexes to see they are identical after creation.
					targetCur, targetErr := targetCollection.Indexes().List(ctx)
					compatCur, compatErr := compatCollection.Indexes().List(ctx)

					require.NoError(t, compatErr)
					assert.Equal(t, compatErr, targetErr)

					targetIndexes := FetchAll(t, ctx, targetCur)
					compatIndexes := FetchAll(t, ctx, compatCur)

					assert.Equal(t, compatIndexes, targetIndexes)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case emptyResult:
				assert.False(t, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

// TestIndexesCreateRunCommand tests specific behavior for index creation that can be only provided through RunCommand.
func TestIndexesCreateRunCommand(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes creation is not supported for Tigris")

	t.Parallel()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)
	targetCollection := targetCollections[0]
	compatCollection := compatCollections[0]

	for name, tc := range map[string]struct { //nolint:vet // for readability
		collectionName any
		indexName      any
		key            any
		resultType     compatTestCaseResultType // defaults to nonEmptyResult
		skip           string                   // optional, skip test with a specified reason
	}{
		"invalid-collection-name": {
			collectionName: 42,
			key:            bson.D{{"v", -1}},
			indexName:      "custom-name",
			resultType:     emptyResult,
		},
		"nil-collection-name": {
			collectionName: nil,
			key:            bson.D{{"v", -1}},
			indexName:      "custom-name",
			resultType:     emptyResult,
		},
		"index-name-not-set": {
			collectionName: "test",
			key:            bson.D{{"v", -1}},
			indexName:      nil,
			resultType:     emptyResult,
			skip:           "https://github.com/FerretDB/FerretDB/issues/2311",
		},
		"empty-index-name": {
			collectionName: "test",
			key:            bson.D{{"v", -1}},
			indexName:      "",
			resultType:     emptyResult,
			skip:           "https://github.com/FerretDB/FerretDB/issues/2311",
		},
		"non-string-index-name": {
			collectionName: "test",
			key:            bson.D{{"v", -1}},
			indexName:      42,
			resultType:     emptyResult,
		},
		"existing-name-different-key-length": {
			collectionName: "test",
			key:            bson.D{{"_id", 1}, {"v", 1}},
			indexName:      "_id_", // the same name as the default index
			skip:           "https://github.com/FerretDB/FerretDB/issues/2311",
		},
		"invalid-key": {
			collectionName: "test",
			key:            42,
			resultType:     emptyResult,
		},
		"empty-key": {
			collectionName: "test",
			key:            bson.D{},
			resultType:     emptyResult,
		},
		"key-not-set": {
			collectionName: "test",
			resultType:     emptyResult,
			skip:           "https://github.com/FerretDB/FerretDB/issues/2311",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
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

			var targetRes bson.D
			targetErr := targetCollection.Database().RunCommand(
				ctx, bson.D{
					{"createIndexes", tc.collectionName},
					{"indexes", bson.A{indexesDoc}},
				},
			).Decode(&targetRes)

			var compatRes bson.D
			compatErr := compatCollection.Database().RunCommand(
				ctx, bson.D{
					{"createIndexes", tc.collectionName},
					{"indexes", bson.A{indexesDoc}},
				},
			).Decode(&compatRes)

			if tc.resultType == emptyResult {
				require.Nil(t, targetRes)
				require.Nil(t, compatRes)
			}

			AssertMatchesCommandError(t, compatErr, targetErr)
			assert.Equal(t, compatRes, targetRes)

			targetErr = targetCollection.Database().RunCommand(
				ctx, bson.D{{"listIndexes", tc.collectionName}},
			).Decode(&targetRes)

			compatErr = compatCollection.Database().RunCommand(
				ctx, bson.D{{"listIndexes", tc.collectionName}},
			).Decode(&targetRes)

			require.Nil(t, targetRes)
			require.Nil(t, compatRes)

			AssertMatchesCommandError(t, compatErr, targetErr)
		})
	}
}
