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
)

// updateCompatTestCase describes update compatibility test case.
type updateCompatTestCase struct {
	update        bson.D                   // required if replace is nil
	replace       bson.D                   // required if update is nil
	resultType    compatTestCaseResultType // defaults to nonEmptyResult
	skip          string                   // skips test if non-empty
	skipForTigris string                   // skips test for Tigris if non-empty
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

			// Use per-test setup because updates modify data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

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

							filter := bson.D{{"_id", id}}
							var targetUpdateRes, compatUpdateRes *mongo.UpdateResult
							var targetErr, compatErr error

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
								if errors.As(targetErr, &e) && e.Name == "DocumentValidationFailure" {
									if e.HasErrorCodeWithMessage(121, "json schema validation failed for field") {
										setup.SkipForTigrisWithReason(t, targetErr.Error())
									}
								}

								assert.Equal(t, compatErr, targetErr)
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
			skipForTigris: "TODO",
		},
		"ReplaceNull": {
			replace:       bson.D{{"v", nil}},
			skipForTigris: "TODO",
		},
		"ReplaceEmptyDocument": {
			replace: bson.D{},
		},
	}

	testUpdateCompat(t, testCases)
}
