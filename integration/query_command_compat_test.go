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

// queryCountCompatTestCase describes command query compatibility test case.
type queryCountCompatTestCase struct {
	query   bson.D // optional
	command bson.D // defaults to count collection
}

// testQueryCountCompat tests command query compatibility test cases.
func testQueryCountCompat(t *testing.T, testCases map[string]queryCountCompatTestCase) {
	t.Helper()

	// Use shared setup because find queries can't modify data.
	// TODO Use read-only user. https://github.com/FerretDB/FerretDB/issues/1025
	ctx, targetCollections, compatCollections := setup.SetupCompat(t)

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

					var targetCommand, compatCommand bson.D
					switch {
					case tc.command != nil:
						targetCommand = tc.command
						compatCommand = tc.command
					case tc.query != nil:
						targetCommand = bson.D{{"count", targetCollection.Name()}, {"query", tc.query}}
						compatCommand = bson.D{{"count", compatCollection.Name()}, {"query", tc.query}}
					default:
						targetCommand = bson.D{{"count", targetCollection.Name()}}
						compatCommand = bson.D{{"count", compatCollection.Name()}}
					}

					var targetRes, compatRes bson.D
					targetErr := targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetRes)
					compatErr := compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatRes)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}

					require.NoError(t, compatErr, "compat error; target returned no error")

					AssertEqualDocuments(t, compatRes, targetRes)
				})
			}
		})
	}
}

func TestQueryCommandCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCountCompatTestCase{
		"CountAllDocuments": {},
		"CountExactlyOneDocument": {
			query: bson.D{{"v", true}},
		},
		"CountExactlyOneDocumentWithIdFilter": {
			query: bson.D{{"_id", "bool-true"}},
		},
		"CountArrays": {
			query: bson.D{{"v", bson.D{{"$type", "array"}}}},
		},
		"CountNonExistingCollection": {
			command: bson.D{
				{"count", "doesnotexist"},
				{"query", bson.D{{"v", true}}},
			},
		},
	}

	testQueryCountCompat(t, testCases)
}
