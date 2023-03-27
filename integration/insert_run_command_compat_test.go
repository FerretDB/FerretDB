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
)

type insertRunCommandCompatTestCase struct {
	altErrorMsg string // optional, alternative error message in case of error
	ordered     any    // required, ordered parameter
	documents   []any  // required, slice of bson.D to be insert

	skip string // optional, reason to skip the test
}

// testInsertRunCommandCompat tests insert compatibility test cases with invalid parameters.
// It uses runCommand instead of insertOne or insertMany to let more invalid parameters being used.
func testInsertRunCommandCompat(t *testing.T, testCases map[string]insertRunCommandCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					// RunCommand must be used to give ability set various invalid parameters
					// and prevent the driver from doing any validation.
					var targetRes, compatRes bson.D
					targetErr := targetCollection.Database().RunCommand(ctx, bson.D{
						{"insert", targetCollection.Name()},
						{"documents", tc.documents},
						{"ordered", tc.ordered},
					}).Decode(&targetRes)
					compatErr := compatCollection.Database().RunCommand(ctx, bson.D{
						{"insert", compatCollection.Name()},
						{"documents", tc.documents},
						{"ordered", tc.ordered},
					}).Decode(&compatRes)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						targetErr = UnsetRaw(t, targetErr)
						compatErr = UnsetRaw(t, compatErr)

						if tc.altErrorMsg != "" {
							AssertMatchesCommandError(t, compatErr, targetErr)

							var expectedErr mongo.CommandError
							require.True(t, errors.As(compatErr, &expectedErr))
							AssertEqualAltError(t, expectedErr, tc.altErrorMsg, targetErr)
						} else {
							require.Equal(t, compatErr, targetErr)
						}

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					t.Logf("Compat (expected) result: %v", compatRes)
					t.Logf("Target (actual)   result: %v", targetRes)
					assert.Equal(t, compatRes, targetRes)
				})
			}
		})
	}
}

func TestInsertRunCommandCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]insertRunCommandCompatTestCase{
		"InsertOrderedInvalid": {
			documents: []any{
				bson.D{{"_id", "foo"}},
			},
			ordered:     "foo",
			altErrorMsg: "BSON field 'ordered' is the wrong type 'string', expected type 'bool'",
		},

		"InsertEmpty": {
			documents: []any{
				bson.D{{}},
			},
			ordered: true,
			skip:    "https://github.com/FerretDB/FerretDB/issues/1714",
		},
	}

	testInsertRunCommandCompat(t, testCases)
}
