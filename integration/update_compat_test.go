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
	update      bson.D                   // required if replace is nil
	replace     bson.D                   // required if update is nil
	filter      bson.D                   // defaults to bson.D{{"_id", id}}
	updateOpts  *options.UpdateOptions   // defaults to nil
	replaceOpts *options.ReplaceOptions  // defaults to nil
	resultType  compatTestCaseResultType // defaults to nonEmptyResult
	providers   []shareddata.Provider    // defaults to shareddata.AllProviders()

	skip           string // skips test if non-empty
	failsForSQLite string // optional, if set, the case is expected to fail for SQLite due to given issue
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
								targetUpdateRes, targetErr = targetCollection.UpdateOne(ctx, filter, update, tc.updateOpts)
								compatUpdateRes, compatErr = compatCollection.UpdateOne(ctx, filter, update, tc.updateOpts)
							} else {
								targetUpdateRes, targetErr = targetCollection.ReplaceOne(ctx, filter, replace, tc.replaceOpts)
								compatUpdateRes, compatErr = compatCollection.ReplaceOne(ctx, filter, replace, tc.replaceOpts)
							}

							if targetErr != nil {
								t.Logf("Target error: %v", targetErr)

								if targetErr.Error() == "update document must have at least one element" {
									// mongo go driver sent error that the update document is empty.
									require.Equal(t, compatErr, targetErr)

									return
								}

								// AssertMatchesWriteError compares error types and codes, it does not compare messages.
								AssertMatchesWriteError(t, compatErr, targetErr)
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

// testUpdateManyCompatTestCase describes update compatibility test case.
type testUpdateManyCompatTestCase struct { //nolint:vet // used for testing only
	update     bson.D                   // required if replace is nil
	filter     bson.D                   // defaults to bson.D{{"_id", id}}
	updateOpts *options.UpdateOptions   // defaults to nil
	resultType compatTestCaseResultType // defaults to nonEmptyResult
	providers  []shareddata.Provider    // defaults to shareddata.AllProviders()

	skip string // skips test if non-empty
}

// testUpdateManyCompat tests update compatibility test cases.
func testUpdateManyCompat(t *testing.T, testCases map[string]testUpdateManyCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
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

			update := tc.update
			require.NotNil(t, update, "`update` must be set")

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					allDocs := FindAll(t, ctx, targetCollection)

					for _, doc := range allDocs {
						id, _ := ConvertDocument(t, doc).Get("_id")
						require.NotNil(t, id)

						t.Run(fmt.Sprint(id), func(t *testing.T) {
							t.Helper()

							filter := tc.filter
							if tc.filter == nil {
								filter = bson.D{{"_id", id}}
							}

							var targetUpdateRes, compatUpdateRes *mongo.UpdateResult
							var targetErr, compatErr error

							targetUpdateRes, targetErr = targetCollection.UpdateMany(ctx, filter, update, tc.updateOpts)
							compatUpdateRes, compatErr = compatCollection.UpdateMany(ctx, filter, update, tc.updateOpts)

							if targetErr != nil {
								t.Logf("Target error: %v", targetErr)
								t.Logf("Compat error: %v", compatErr)

								// error messages are intentionally not compared
								AssertMatchesWriteError(t, compatErr, targetErr)

								return
							}
							require.NoError(t, compatErr, "compat error; target returned no error")

							if pointer.Get(targetUpdateRes).ModifiedCount > 0 || pointer.Get(compatUpdateRes).ModifiedCount > 0 {
								nonEmptyResults = true
							}

							assert.Equal(t, compatUpdateRes, targetUpdateRes)

							opts := options.Find().SetSort(bson.D{{"_id", 1}})
							targetCursor, targetErr := targetCollection.Find(ctx, bson.D{}, opts)
							compatCursor, compatErr := compatCollection.Find(ctx, bson.D{}, opts)

							if targetCursor != nil {
								defer targetCursor.Close(ctx)
							}
							if compatCursor != nil {
								defer compatCursor.Close(ctx)
							}

							require.NoError(t, targetErr)
							require.NoError(t, compatErr)

							targetRes := FetchAll(t, ctx, targetCursor)
							compatRes := FetchAll(t, ctx, compatCursor)

							AssertEqualDocumentsSlice(t, compatRes, targetRes)

							if len(targetRes) > 0 || len(compatRes) > 0 {
								nonEmptyResults = true
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

// updateCommandCompatTestCase describes update command compatibility test case.
type updateCommandCompatTestCase struct {
	multi      any                      // defaults to false, if true updates multiple documents
	update     bson.D                   // required
	filter     bson.D                   // defaults to bson.D{{"_id", id}}
	resultType compatTestCaseResultType // defaults to nonEmptyResult

	skip string // skips test if non-empty
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
								t.Logf("Compat error: %v", compatErr)

								// error messages are intentionally not compared
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
									t.Logf("Compat error: %v", compatErr)

									// error messages are intentionally not compared
									AssertMatchesCommandError(t, compatErr, targetErr)

									return
								}
								require.NoError(t, compatErr, "compat error; target returned no error")

								targetRes := FetchAll(t, ctx, targetCursor)
								compatRes := FetchAll(t, ctx, compatCursor)

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
	paths      []types.Path             // paths to check after update
	update     bson.D                   // required
	filter     bson.D                   // defaults to bson.D{{"_id", id}}
	resultType compatTestCaseResultType // defaults to nonEmptyResult

	skip string // skips test if non-empty
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
								// AssertMatchesWriteError compares error types and codes, it does not compare messages.
								AssertMatchesWriteError(t, compatErr, targetErr)
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
			replace: bson.D{{"v", ""}},
		},
		"ReplaceNull": {
			replace: bson.D{{"v", nil}},
		},
		"ReplaceEmptyDocument": {
			replace: bson.D{},
		},
		"ReplaceNonExistentUpsert": {
			filter:         bson.D{{"non-existent", "no-match"}},
			replace:        bson.D{{"_id", "new"}},
			replaceOpts:    options.Replace().SetUpsert(true),
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3183",
		},
		"UpdateNonExistentUpsert": {
			filter:         bson.D{{"_id", "non-existent"}},
			update:         bson.D{{"$set", bson.D{{"v", int32(42)}}}},
			updateOpts:     options.Update().SetUpsert(true),
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3183",
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
			filter:  bson.D{{"v.array.0", bson.D{{"$eq", int32(42)}}}, {"_id", "document-composite"}},
			replace: bson.D{{"replacement-value", int32(1)}},
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
		"TrueEmptyDocument": {
			update: bson.D{},
			multi:  true,
			skip:   "https://github.com/FerretDB/FerretDB/issues/2630",
		},
		"FalseEmptyDocument": {
			update: bson.D{},
			multi:  false,
		},
	}

	testUpdateCommandCompat(t, testCases)
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
