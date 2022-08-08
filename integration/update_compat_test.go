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

// updateCompatTestCase describes update compatibility test case.
type updateCompatTestCase struct {
	update bson.D // required
	skip   string // skips test if non-empty
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

			// Use per-test setup because update queries modify data.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			update := tc.update
			require.NotNil(t, update)

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

							targetUpdateRes, targetErr := targetCollection.UpdateByID(ctx, id, update)
							compatUpdateRes, compatErr := compatCollection.UpdateByID(ctx, id, update)

							if targetErr != nil {
								targetErr = UnsetRaw(t, targetErr)
								compatErr = UnsetRaw(t, compatErr)
								assert.Equal(t, compatErr, targetErr)
							} else {
								require.NoError(t, compatErr)
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
		})
	}
}
