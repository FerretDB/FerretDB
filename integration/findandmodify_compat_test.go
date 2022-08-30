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

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestFindAndModifyCompatSimple(t *testing.T) {
	testCases := map[string]findAndModifyCompatTestCase{
		"EmptyQueryRemove": {
			skipForTigris: "Arrays are not supported yet - https://github.com/FerretDB/FerretDB/issues/908",
			command: bson.D{
				{"query", bson.D{}},
				{"remove", true},
			},
		},
	}

	testFindAndModifyCompat(t, testCases)
}

// findAndModifyCompatTestCase describes findAndModify compatibility test case.
type findAndModifyCompatTestCase struct {
	command       bson.D
	skip          string // skips test if non-empty
	skipForTigris string // skips test for Tigris if non-empty
}

// testUpdateCompat tests update compatibility test cases.
func testFindAndModifyCompat(t *testing.T, testCases map[string]findAndModifyCompatTestCase) {
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

			// Use per-test setup because findAndModify modifies data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			//	var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					targetCommand := bson.D{{"findAndModify", targetCollection.Name()}}
					targetCommand = append(targetCommand, tc.command...)
					if targetCommand.Map()["sort"] == nil {
						targetCommand = append(targetCommand, bson.D{{"sort", bson.D{{"_id", 1}}}}...)
					}

					compatCommand := bson.D{{"findAndModify", compatCollection.Name()}}
					compatCommand = append(compatCommand, tc.command...)
					if compatCommand.Map()["sort"] == nil {
						compatCommand = append(compatCommand, bson.D{{"sort", bson.D{{"_id", 1}}}}...)
					}

					var targetRes, compatRes bson.D
					var targetErr, compatErr error
					targetErr = targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetRes)
					compatErr = compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatRes)
					require.Equal(t, targetErr, compatErr)
					AssertEqualDocuments(t, targetRes, compatRes)

					/*var targetFindRes, compatFindRes bson.D
					require.NoError(t, targetCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&targetFindRes))
					require.NoError(t, compatCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&compatFindRes))
					AssertEqualDocuments(t, compatFindRes, targetFindRes)*/
				})
			}
		})
	}
	/*
		switch tc.resultType {
		case nonEmptyResult:
			assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be modified)")
		case emptyResult:
			assert.False(t, nonEmptyResults, "expected empty results (no documents should be modified)")
		default:
			t.Fatalf("unknown result type %v", tc.resultType)
		}*/
}
