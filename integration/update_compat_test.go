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
	"fmt"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// updateCompatTestCase describes update compatibility test case.
type updateCompatTestCase struct {
	update        bson.D                   // required if replace is nil
	replace       bson.D                   // required if update is nil
	resultType    compatTestCaseResultType // defaults to nonEmptyResult
	skip          string                   // skips test if non-empty
	skipForTigris string                   // skips test for Tigris if non-empty

	// TODO remove is possible: https://github.com/FerretDB/FerretDB/issues/1668
	providers []shareddata.Provider // defaults to shareddata.AllProviders()
}

// testUpdateCompat tests update compatibility test cases.
func testUpdateCompat(t *testing.T, testCases map[string]updateCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}
			if tc.skipForTigris != "" {
				setup.SkipForTigrisWithReason(t, tc.skipForTigris)
			}

			update, replace := tc.update, tc.replace
			if update != nil {
				require.Nil(t, replace, "`replace` must be nil if `update` is set")
			} else {
				require.NotNil(t, replace, "`replace` must be set if `update` is nil")
			}

			t.Parallel()

			providers := shareddata.AllProviders()
			if tc.providers != nil {
				providers = tc.providers
			}

			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                providers,
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					allDocs := FindAll(t, ctx, targetCollection)

					// try to update each document one by one
					for _, doc := range allDocs {
						id, ok := doc.Map()["_id"]
						require.True(t, ok)

						t.Run(fmt.Sprint(id), func(t *testing.T) {
							t.Helper()

							var targetUpdateRes, compatUpdateRes *mongo.UpdateResult
							var targetErr, compatErr error

							// TODO replace with UpdateMany? What about Replace?
							// https://github.com/FerretDB/FerretDB/issues/1507
							if update != nil {
								targetUpdateRes, targetErr = targetCollection.UpdateOne(ctx, bson.D{{"_id", id}}, update)
								compatUpdateRes, compatErr = compatCollection.UpdateOne(ctx, bson.D{{"_id", id}}, update)
							} else {
								targetUpdateRes, targetErr = targetCollection.ReplaceOne(ctx, bson.D{{"_id", id}}, replace)
								compatUpdateRes, compatErr = compatCollection.ReplaceOne(ctx, bson.D{{"_id", id}}, replace)
							}

							if targetErr != nil {
								t.Logf("Target error: %v", targetErr)
								targetErr = UnsetRaw(t, targetErr)
								compatErr = UnsetRaw(t, compatErr)

								// Skip updates that could not be performed due to Tigris schema validation.
								var e mongo.CommandError
								if errors.As(targetErr, &e) && e.HasErrorCode(documentValidationFailureCode) {
									setup.SkipForTigrisWithReason(t, targetErr.Error())
								}

								AssertMatchesWriteErrorCode(t, compatErr, targetErr)
							} else {
								require.NoError(t, compatErr, "compat error; target returned no error")
							}

							if pointer.Get(targetUpdateRes).ModifiedCount > 0 || pointer.Get(compatUpdateRes).ModifiedCount > 0 {
								nonEmptyResults = true
							}

							assert.Equal(t, compatUpdateRes, targetUpdateRes)

							var targetFindRes, compatFindRes bson.D
							require.NoError(t, targetCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&targetFindRes))
							require.NoError(t, compatCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&compatFindRes))
							AssertEqualDocuments(t, compatFindRes, targetFindRes)
						})
					}
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

func TestUpdateCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"UpdateEmptyDocument": {
			update:     bson.D{},
			resultType: emptyResult,
		},
		"ReplaceSimple": {
			replace: bson.D{{"v", "foo"}},
		},
		"ReplaceEmpty": {
			replace:       bson.D{{"v", ""}},
			skipForTigris: "https://github.com/FerretDB/FerretDB/issues/1061",
		},
		"ReplaceNull": {
			replace: bson.D{{"v", nil}},
		},
		"ReplaceEmptyDocument": {
			replace: bson.D{},
		},
	}

	testUpdateCompat(t, testCases)
}

func TestReplaceKeepOrderCompat(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers: shareddata.Providers{shareddata.Int32s},
	})

	ctx := s.Ctx
	targetCollection := s.TargetCollections[0]
	compatCollection := s.CompatCollections[0]

	replace := bson.D{{"_id", "arr"}, {"c", int32(1)}, {"b", int32(2)}, {"a", int32(3)}}
	filter := bson.D{{"_id", "arr"}}

	_, err := targetCollection.InsertOne(ctx, filter)
	require.NoError(t, err)
	_, err = compatCollection.InsertOne(ctx, filter)
	require.NoError(t, err)

	_, err = targetCollection.ReplaceOne(ctx, filter, replace)
	require.NoError(t, err)
	_, err = compatCollection.ReplaceOne(ctx, filter, replace)
	require.NoError(t, err)

	targetResult := targetCollection.FindOne(ctx, filter)
	require.NoError(t, targetResult.Err())
	compatResult := compatCollection.FindOne(ctx, filter)
	require.NoError(t, compatResult.Err())

	var targetDoc, compatDoc bson.D
	require.NoError(t, targetResult.Decode(&targetDoc))
	require.NoError(t, compatResult.Decode(&compatDoc))

	assert.Equal(t, compatDoc, targetDoc)
}
