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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

// deleteCommandCompatTestCase describes delete compatibility test case.
type deleteCommandCompatTestCase struct { //nolint:vet // for readability
	deletes    bson.A                   // required
	resultType compatTestCaseResultType // defaults to nonEmptyResult

	skip string // optional, skip test with a specified reason
}

func testDeleteCommandCompat(t *testing.T, testCases map[string]deleteCommandCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			// Use per-test setup because deletes modify data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					var targetRes, compatRes bson.D
					targetErr := targetCollection.Database().RunCommand(ctx,
						bson.D{
							{"delete", targetCollection.Name()},
							{"deletes", tc.deletes},
						}).Decode(&targetRes)
					compatErr := compatCollection.Database().RunCommand(ctx,
						bson.D{
							{"delete", compatCollection.Name()},
							{"deletes", tc.deletes},
						}).Decode(&compatRes)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						targetErr = UnsetRaw(t, targetErr)
						compatErr = UnsetRaw(t, compatErr)
						assert.Equal(t, compatErr, targetErr)
					} else {
						require.NoError(t, compatErr, "compat error; target returned no error")
					}

					assert.Equal(t, compatRes, targetRes)

					targetDocs := FindAll(t, ctx, targetCollection)
					compatDocs := FindAll(t, ctx, compatCollection)

					t.Logf("Compat (expected) IDs: %v", CollectIDs(t, compatDocs))
					t.Logf("Target (actual)   IDs: %v", CollectIDs(t, targetDocs))
					AssertEqualDocumentsSlice(t, compatDocs, targetDocs)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be deleted)")
			case emptyResult:
				assert.False(t, nonEmptyResults, "expected empty results (no documents should be deleted)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

func TestDeleteCompatCommand(t *testing.T) {
	t.Parallel()

	testCases := map[string]deleteCommandCompatTestCase{
		"DuplicateFilter": {
			deletes: bson.A{
				bson.D{{"q", bson.D{{"v", "foo"}}}, {"limit", 1}},
				bson.D{{"q", bson.D{{"v", "foo"}}}, {"limit", 1}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2935",
		},
	}

	testDeleteCommandCompat(t, testCases)
}
