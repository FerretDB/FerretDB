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

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

// insertCompatTestCase describes insert compatibility test case.
type insertCompatTestCase struct {
	insert bson.D // required
	skip   string // skips test if non-empty
}

// TestInsertCompat checks that insert works for various cases
func TestInsertCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]insertCompatTestCase{
		"InsertNull": {
			insert: bson.D{{"_id", "insert_new_null"}, {"v", nil}},
		},
	}

	testInsertCompat(t, testCases)
}

// testInsertCompat tests insert compatibility test cases.
func testInsertCompat(t *testing.T, testCases map[string]insertCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			// Use per-test setup because inserts modify data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			insert := tc.insert
			require.NotNil(t, insert)
			id := insert.Map()["_id"]

			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					t.Run(fmt.Sprint(id), func(t *testing.T) {
						t.Helper()

						targetinsertRes, targetErr := targetCollection.InsertOne(ctx, insert)
						compatinsertRes, compatErr := compatCollection.InsertOne(ctx, insert)

						if targetErr != nil {
							targetErr = UnsetRaw(t, targetErr)
							compatErr = UnsetRaw(t, compatErr)
							assert.Equal(t, compatErr, targetErr)
						} else {
							require.NoError(t, compatErr)
						}

						assert.Equal(t, compatinsertRes, targetinsertRes)

						var targetFindRes, compatFindRes bson.D
						require.NoError(t, targetCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&targetFindRes))
						require.NoError(t, compatCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&compatFindRes))
						AssertEqualDocuments(t, compatFindRes, targetFindRes)
					})
				})
			}
		})
	}
}
