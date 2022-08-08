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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

// deleteCompatTestCase describes delete compatibility test case.
type deleteCompatTestCase struct {
	filter  bson.D // required
	ordered bool   // defaults to false
	skip    string // skips test if non-empty
}

func TestDeleteCompat(t *testing.T) {
	testCases := map[string]deleteCompatTestCase{
		"Empty": {
			filter: bson.D{},
		},
	}

	testDeleteCompat(t, testCases)
}

// testDeleteCompat tests delete compatibility test cases.
func testDeleteCompat(t *testing.T, testCases map[string]deleteCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			// Use per-test setup because delete queries modify data.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			filter := tc.filter
			require.NotNil(t, filter)

			opts := options.BulkWrite().SetOrdered(tc.ordered)

			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					dmm := mongo.NewDeleteManyModel().SetFilter(filter)

					targetRes, targetErr := targetCollection.BulkWrite(ctx, []mongo.WriteModel{dmm}, opts)
					compatRes, compatErr := compatCollection.BulkWrite(ctx, []mongo.WriteModel{dmm}, opts)

					if targetErr != nil {
						targetErr = UnsetRaw(t, targetErr)
						compatErr = UnsetRaw(t, compatErr)
						assert.Equal(t, compatErr, targetErr)
					} else {
						require.NoError(t, compatErr)
					}

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
