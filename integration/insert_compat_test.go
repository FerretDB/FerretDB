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

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/setup"
)

type insertCompatTestCase struct {
	insert     bson.D                   // required
	resultType compatTestCaseResultType // defaults to nonEmptyResult
}

// testInsertCompat tests insert compatibility test cases.
func testInsertCompat(t *testing.T, testCases map[string]insertCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			insert := tc.insert
			require.NotNil(t, insert, "insert should be set")

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					targetInsertRes, targetErr := targetCollection.InsertOne(ctx, insert)
					compatInsertRes, compatErr := compatCollection.InsertOne(ctx, insert)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						targetErr = UnsetRaw(t, targetErr)
						compatErr = UnsetRaw(t, compatErr)

						// Skip inserts that could not be performed due to Tigris schema validation.
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

					targetID, _ := pointer.Get(targetInsertRes).InsertedID.(primitive.ObjectID)
					compatID, _ := pointer.Get(compatInsertRes).InsertedID.(primitive.ObjectID)

					if !(targetID.IsZero() && compatID.IsZero()) {
						nonEmptyResults = true
					}

					var targetFindRes, compatFindRes bson.D
					targetCursor, err := targetCollection.Find(ctx, bson.D{{}})
					require.NoError(t, err)
					defer targetCursor.Close(ctx)
					targetCursor.Decode(&targetFindRes)

					compatCursor, err := compatCollection.Find(ctx, bson.D{{}})
					require.NoError(t, err)
					defer compatCursor.Close(ctx)
					compatCursor.Decode(&targetFindRes)

					AssertEqualDocuments(t, compatFindRes, targetFindRes)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results")
			case emptyResult:
				assert.False(t, nonEmptyResults, "expected empty results")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

func TestInsertCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]insertCompatTestCase{
		"InsertEmptyDocument": {
			insert: bson.D{},
		},
		"InsertIDArray": {
			insert:     bson.D{{"_id", bson.A{"foo", "bar"}}},
			resultType: emptyResult,
		},
		"InsertIDRegex": {
			insert:     bson.D{{"_id", "^regex$"}},
			resultType: emptyResult,
		},
	}

	testInsertCompat(t, testCases)
}
