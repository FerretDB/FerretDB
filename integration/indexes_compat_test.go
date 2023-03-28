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
	t.Helper()

	setup.SkipForTigrisWithReason(t, "Indexes creation is not supported for Tigris")

	for name, tc := range map[string]struct { //nolint:vet // for readability
		models      []mongo.IndexModel
		altErrorMsg string                   // optional, alternative error message in case of error
		resultType  compatTestCaseResultType // defaults to nonEmptyResult
	}{
		"empty": {
			models:     []mongo.IndexModel{},
			resultType: emptyResult,
		},
		"single-index": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"v", -1}},
				},
			},
		},
		"single-duplicate": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"_id", 1}}, // this index is already created by default
				},
			},
		},
		"single-non-existent-field": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"field-does-not-exist", 1}},
				},
			},
		},
		"single-dot-notation": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"v.foo", 1}},
				},
			},
		},
		"single-dangerous-key": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{
						{"v", 1},
						{"foo'))); DROP TABlE test._ferretdb_database_metadata; CREATE INDEX IF NOT EXISTS test ON test.test (((_jsonb->'foo", 1},
					},
				},
			},
		},
		"single-same-key": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"v", -1}, {"v", 1}},
				},
			},
			resultType: emptyResult,
		},

		"multi-direction-different-indexes": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"v", -1}},
				},
				{
					Keys: bson.D{{"v", 1}},
				},
			},
		},
		"multi-order": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"foo", -1}},
				},
				{
					Keys: bson.D{{"v", 1}},
				},
				{
					Keys: bson.D{{"bar", 1}},
				},
			},
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
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()
			t.Parallel()

			// Use per-test setup because createIndexes modifies collection state.
			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                shareddata.AllProviders(),
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

// TestIndexesInvalidCollectionName tests the behavior when we try to create an index with invalid parameters.
func TestIndexesInvalidCollectionName(t *testing.T) {
	t.Parallel()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)
	targetCollection := targetCollections[0]
	compatCollection := compatCollections[0]

	for name, tc := range map[string]struct {
		collectionName any
		name           string
		key            bson.D
	}{
		"invalid-collection-name": {
			collectionName: 42,
			key:            bson.D{{"v", -1}},
			name:           "custom-index",
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var targetRes bson.D
			targetErr := targetCollection.Database().RunCommand(
				ctx, bson.D{
					{"createIndexes", tc.collectionName},
					{"indexes", bson.A{bson.D{{"key", tc.key}, {"name", tc.name}}}},
				},
			).Decode(&targetRes)

			var compatRes bson.D
			compatErr := compatCollection.Database().RunCommand(
				ctx, bson.D{
					{"createIndexes", tc.collectionName},
					{"indexes", bson.A{bson.D{{"key", tc.key}}}},
				},
			).Decode(&compatRes)

			require.Nil(t, targetRes)
			require.Nil(t, compatRes)

			AssertMatchesCommandError(t, compatErr, targetErr)

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
