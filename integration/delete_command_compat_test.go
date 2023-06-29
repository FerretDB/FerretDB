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
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// deleteCommandCompatTestCase describes delete compatibility test case.
type deleteCommandCompatTestCase struct {
	skip    string // optional, skip test with a specified reason
	deletes bson.A // required
}

func testDeleteCommandCompat(t *testing.T, testCases map[string]deleteCommandCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			// Use per-test setup because deletes modify data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

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
						t.Logf("Compat error: %v", compatErr)

						// error messages are intentionally not compared
						AssertMatchesWriteError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					t.Logf("Compat (expected) result: %v", compatRes)
					t.Logf("Target (actual)   result: %v", targetRes)
					assert.Equal(t, compatRes, targetRes)

					targetDocs := FindAll(t, ctx, targetCollection)
					compatDocs := FindAll(t, ctx, compatCollection)

					t.Logf("Compat (expected) IDs: %v", CollectIDs(t, compatDocs))
					t.Logf("Target (actual)   IDs: %v", CollectIDs(t, targetDocs))
					AssertEqualDocumentsSlice(t, compatDocs, targetDocs)
				})
			}
		})
	}
}

func TestDeleteCommandCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]deleteCommandCompatTestCase{
		"OneLimited": {
			deletes: bson.A{
				bson.D{{"q", bson.D{{"v", int32(0)}}}, {"limit", 1}},
			},
		},
		"TwoLimited": {
			deletes: bson.A{
				bson.D{{"q", bson.D{{"v", int32(42)}}}, {"limit", 1}},
				bson.D{{"q", bson.D{{"v", "foo"}}}, {"limit", 1}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2935",
		},
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

func TestDeleteCommandCompatNotExistingDatabase(t *testing.T) {
	t.Parallel()

	res := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers: []shareddata.Provider{shareddata.Int32s},
	})
	ctx, targetColl, compatColl := res.Ctx, res.TargetCollections[0], res.CompatCollections[0]

	targetDB := targetColl.Database().Client().Database("doesnotexist")
	compatDB := compatColl.Database().Client().Database("doesnotexist")

	var targetRes, compatRes bson.D

	targetErr := targetDB.RunCommand(ctx, bson.D{
		{"delete", "test"},
		{"deletes", bson.A{
			bson.D{{"q", bson.D{{"v", "foo"}}}, {"limit", 1}},
		}},
	}).Decode(&targetRes)
	compatErr := compatDB.RunCommand(ctx, bson.D{
		{"delete", "test"},
		{"deletes", bson.A{
			bson.D{{"q", bson.D{{"v", "foo"}}}, {"limit", 1}},
		}},
	}).Decode(&compatRes)

	assert.NoError(t, targetErr)
	assert.NoError(t, compatErr)

	assert.Equal(t, compatRes, targetRes)
}
