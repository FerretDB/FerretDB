// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
)

func TestIndexesList(t *testing.T) {
	t.Parallel()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)

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

// TestIndexesListRunCommand tests the behavior when listIndexes is called through RunCommand.
// It's handy to use it to test the correctness of errors.
func TestIndexesListRunCommand(t *testing.T) {
	t.Parallel()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)
	targetCollection := targetCollections[0]
	compatCollection := compatCollections[0]

	for name, tc := range map[string]struct {
		collectionName any
		expectedError  *mongo.CommandError
	}{
		"non-existent-collection": {
			collectionName: "non-existent-collection",
		},
		"invalid-collection-name": {
			collectionName: 42,
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var targetRes bson.D
			targetErr := targetCollection.Database().RunCommand(
				ctx, bson.D{{"listIndexes", tc.collectionName}},
			).Decode(&targetRes)

			var compatRes bson.D
			compatErr := compatCollection.Database().RunCommand(
				ctx, bson.D{{"listIndexes", tc.collectionName}},
			).Decode(&targetRes)

			require.Nil(t, targetRes)
			require.Nil(t, compatRes)

			AssertMatchesCommandError(t, compatErr, targetErr)
		})
	}
}

type createIndexTestCase struct {
	models []mongo.IndexModel
}

// testIndexesCreateMany tests the behavior when collection.Indexes().CreateMany() is called.
func testIndexesCreateMany(t *testing.T, testCases map[string]createIndexTestCase) {
	t.Helper()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)
	// TODO add non-existent collection test case

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()
			t.Parallel()

			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					targetRes, targetErr := targetCollection.Indexes().CreateMany(ctx, tc.models)
					compatRes, compatErr := compatCollection.Indexes().CreateMany(ctx, tc.models)

					require.Equal(t, compatErr, targetErr)
					assert.Equal(t, compatRes, targetRes)

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
		})
	}
}

func TestIndexesCreate(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes creation is not supported for Tigris")

	testCases := map[string]createIndexTestCase{
		"single-index": {
			models: []mongo.IndexModel{
				{
					Keys: bson.D{{"v", 1}},
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
	}

	testIndexesCreateMany(t, testCases)
}
