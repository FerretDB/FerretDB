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
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// updateCompatTestCase describes update compatibility test case.
type updateCompatTestCase struct {
	update bson.D
	skip   string
}

// testUpdateCompat tests update compatibility test cases.
func testUpdateCompat(t *testing.T, testCases map[string]updateCompatTestCase) {
	t.Helper()

	providers := []shareddata.Provider{
		shareddata.FixedScalars,
		shareddata.Scalars,
		shareddata.Composites,
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			// Use per-test setup because ypdate queries modify data.
			ctx, targetCollection, compatCollection := setup.SetupCompat(t, providers...)

			update := tc.update
			require.NotNil(t, update)

			ids := shareddata.IDs(providers...)
			for _, id := range ids {
				id := id
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
}
