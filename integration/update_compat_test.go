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
	"time"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

// updateCompatTestCase describes update compatibility test case.
type updateCompatTestCase struct {
	update        bson.D                   // required if replace is nil
	replace       bson.D                   // required if update is nil
	filter        bson.D                   // defaults to bson.D{{"_id", id}}
	resultType    compatTestCaseResultType // defaults to nonEmptyResult
	skip          string                   // skips test if non-empty
	skipForTigris string                   // skips test for Tigris if non-empty
	providers     []shareddata.Provider    // defaults to shareddata.AllProviders()
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

			update, replace := tc.update, tc.replace
			if update != nil {
				require.Nil(t, replace, "`replace` must be nil if `update` is set")
			} else {
				require.NotNil(t, replace, "`replace` must be set if `update` is nil")
			}

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					allDocs := FindAll(t, ctx, targetCollection)

					for _, doc := range allDocs {
						id, ok := doc.Map()["_id"]
						require.True(t, ok)

						t.Run(fmt.Sprint(id), func(t *testing.T) {
							t.Helper()

							filter := tc.filter
							if tc.filter == nil {
								filter = bson.D{{"_id", id}}
							}

							var targetUpdateRes, compatUpdateRes *mongo.UpdateResult
							var targetErr, compatErr error

							// TODO replace with UpdateMany/ReplaceMany
							// https://github.com/FerretDB/FerretDB/issues/1507
							if update != nil {
								targetUpdateRes, targetErr = targetCollection.UpdateOne(ctx, filter, update)
								compatUpdateRes, compatErr = compatCollection.UpdateOne(ctx, filter, update)
							} else {
								targetUpdateRes, targetErr = targetCollection.ReplaceOne(ctx, filter, replace)
								compatUpdateRes, compatErr = compatCollection.ReplaceOne(ctx, filter, replace)
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

// updateCommandCompatTestCase describes update command compatibility test case.
type updateCommandCompatTestCase struct {
	multi      any                      // defaults to false, if true updates multiple documents
	skip       string                   // skips test if non-empty
	update     bson.D                   // required
	filter     bson.D                   // defaults to bson.D{{"_id", id}}
	resultType compatTestCaseResultType // defaults to nonEmptyResult
}

// testUpdateCommandCompat tests command compatibility test cases used for update.
// This is used for updating multiple documents and testing multi flag values.
func testUpdateCommandCompat(t *testing.T, testCases map[string]updateCommandCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			// Use per-test setup because updates modify data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			update := tc.update
			require.NotNil(t, update, "`update` must be set")

			multi := tc.multi

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					allDocs := FindAll(t, ctx, targetCollection)

					for _, doc := range allDocs {
						id, ok := doc.Map()["_id"]
						require.True(t, ok)

						t.Run(fmt.Sprint(id), func(t *testing.T) {
							t.Helper()

							filter := tc.filter
							if tc.filter == nil {
								filter = bson.D{{"_id", id}}
							}

							targetCommand := bson.D{
								{"update", targetCollection.Name()},
								{"updates", bson.A{bson.D{
									{"q", filter},
									{"u", update},
									{"multi", multi},
								}}},
							}

							compatCommand := bson.D{
								{"update", compatCollection.Name()},
								{"updates", bson.A{bson.D{
									{"q", filter},
									{"u", update},
									{"multi", multi},
								}}},
							}

							var targetUpdateRes, compatUpdateRes bson.D
							var targetErr, compatErr error

							targetErr = targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetUpdateRes)
							compatErr = compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatUpdateRes)

							if targetErr != nil {
								t.Logf("Target error: %v", targetErr)
								targetErr = UnsetRaw(t, targetErr)
								compatErr = UnsetRaw(t, compatErr)

								// Skip updates that could not be performed due to Tigris schema validation.
								var e mongo.CommandError
								if errors.As(targetErr, &e) && e.HasErrorCode(documentValidationFailureCode) {
									setup.SkipForTigrisWithReason(t, targetErr.Error())
								}

								AssertMatchesCommandError(t, compatErr, targetErr)
							} else {
								require.NoError(t, compatErr, "compat error; target returned no error")
							}

							if targetUpdateRes != nil || compatUpdateRes != nil {
								nonEmptyResults = true
							}

							assert.Equal(t, compatUpdateRes, targetUpdateRes)

							if isMulti, ok := multi.(bool); ok && isMulti {
								// if multi == false, an item updated by compat and target are different.

								opts := options.Find().SetSort(bson.D{{"_id", 1}})
								targetCursor, targetErr := targetCollection.Find(ctx, filter, opts)
								compatCursor, compatErr := compatCollection.Find(ctx, filter, opts)

								if targetCursor != nil {
									defer targetCursor.Close(ctx)
								}
								if compatCursor != nil {
									defer compatCursor.Close(ctx)
								}

								if targetErr != nil {
									t.Logf("Target error: %v", targetErr)
									AssertMatchesCommandError(t, compatErr, targetErr)

									return
								}
								require.NoError(t, compatErr, "compat error; target returned no error")

								var targetRes, compatRes []bson.D
								require.NoError(t, targetCursor.All(ctx, &targetRes))
								require.NoError(t, compatCursor.All(ctx, &compatRes))

								t.Logf("Compat (expected) IDs: %v", CollectIDs(t, compatRes))
								t.Logf("Target (actual)   IDs: %v", CollectIDs(t, targetRes))
								AssertEqualDocumentsSlice(t, compatRes, targetRes)
							}
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

// updateCurrentDateCompatTestCase describes update current date compatibility test case.
type updateCurrentDateCompatTestCase struct {
	skip       string                   // skips test if non-empty
	paths      []types.Path             // paths to check after update
	update     bson.D                   // required
	filter     bson.D                   // defaults to bson.D{{"_id", id}}
	resultType compatTestCaseResultType // defaults to nonEmptyResult
}

// testUpdateCompat tests update compatibility test cases for current date.
// It checks current date in compat and target are within acceptable difference.
func testUpdateCurrentDateCompat(t *testing.T, testCases map[string]updateCurrentDateCompatTestCase) {
	t.Helper()

	maxDifference := 2 * time.Minute

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			// Use per-test setup because updates modify data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			update := tc.update
			require.NotNil(t, update, "`update` must be set")

			paths := tc.paths

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					allDocs := FindAll(t, ctx, targetCollection)

					for _, doc := range allDocs {
						id, ok := doc.Map()["_id"]
						require.True(t, ok)

						t.Run(fmt.Sprint(id), func(t *testing.T) {
							t.Helper()

							filter := tc.filter
							if tc.filter == nil {
								filter = bson.D{{"_id", id}}
							}

							var targetUpdateRes, compatUpdateRes *mongo.UpdateResult
							var targetErr, compatErr error

							targetUpdateRes, targetErr = targetCollection.UpdateOne(ctx, filter, update)
							compatUpdateRes, compatErr = compatCollection.UpdateOne(ctx, filter, update)

							if targetErr != nil {
								t.Logf("Target error: %v", targetErr)
								targetErr = UnsetRaw(t, targetErr)
								compatErr = UnsetRaw(t, compatErr)

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

							targetDoc := ConvertDocument(t, targetFindRes)
							compatDoc := ConvertDocument(t, compatFindRes)

							for _, path := range paths {
								testutil.CompareAndSetByPathTime(t, compatDoc, targetDoc, maxDifference, path)
							}

							assert.Equal(t, compatDoc, targetDoc)
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

func TestUpdateCompatArray(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCompatTestCase{
		"ReplaceDocumentFilter": {
			filter:  bson.D{{"v", bson.D{{"$eq", true}}}},
			replace: bson.D{{"foo", int32(1)}},
		},
		"ReplaceDotNotationFilter": {
			filter:        bson.D{{"v.array.0", bson.D{{"$eq", int32(42)}}}, {"_id", "document-composite"}},
			replace:       bson.D{{"replacement-value", int32(1)}},
			skipForTigris: "Tigris does not support language keyword 'array' as field name",
		},
	}

	testUpdateCompat(t, testCases)
}

func TestUpdateCompatMultiFlagCommand(t *testing.T) {
	t.Parallel()

	testCases := map[string]updateCommandCompatTestCase{
		"False": {
			filter: bson.D{{"v", int32(42)}},
			update: bson.D{{"$set", bson.D{{"v", int32(43)}}}},
			multi:  false,
		},
		"True": {
			filter: bson.D{{"v", int32(42)}},
			update: bson.D{{"$set", bson.D{{"v", int32(43)}}}},
			multi:  true,
		},
		"String": {
			filter:     bson.D{{"v", int32(42)}},
			update:     bson.D{{"$set", bson.D{{"v", int32(43)}}}},
			multi:      "false",
			resultType: emptyResult,
		},
		"Int": {
			filter:     bson.D{{"v", int32(42)}},
			update:     bson.D{{"$set", bson.D{{"v", int32(43)}}}},
			multi:      int32(0),
			resultType: emptyResult,
		},
	}

	testUpdateCommandCompat(t, testCases)
}

func TestReplaceKeepOrderCompat(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Schema violation")

	t.Parallel()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)

	targetCollection := targetCollections[0]
	compatCollection := compatCollections[0]

	replace := bson.D{{"_id", int32(1)}, {"c", int32(1)}, {"b", int32(2)}, {"a", int32(3)}}

	_, err := targetCollection.InsertOne(ctx, bson.D{{"_id", 1}})
	require.NoError(t, err)
	_, err = compatCollection.InsertOne(ctx, bson.D{{"_id", 1}})
	require.NoError(t, err)

	_, err = targetCollection.ReplaceOne(ctx, bson.D{{"_id", 1}}, replace)
	require.NoError(t, err)
	_, err = compatCollection.ReplaceOne(ctx, bson.D{{"_id", 1}}, replace)
	require.NoError(t, err)

	targetResult := targetCollection.FindOne(ctx, bson.D{{"_id", 1}})
	require.NoError(t, targetResult.Err())
	compatResult := compatCollection.FindOne(ctx, bson.D{{"_id", 1}})
	require.NoError(t, compatResult.Err())

	var targetDoc, compatDoc bson.D
	require.NoError(t, targetResult.Decode(&targetDoc))
	require.NoError(t, compatResult.Decode(&compatDoc))

	assert.Equal(t, compatDoc, targetDoc)
}
