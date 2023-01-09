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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

type getMoreCompatTestCase struct {
	skipForTigris string
	filter        bson.D
}

// testQueryCompat tests query compatibility test cases.
func testGetMoreCompat(t *testing.T, testCases map[string]getMoreCompatTestCase) {
	t.Helper()

	// Use shared setup because find queries can't modify data.
	// TODO Use read-only user. https://github.com/FerretDB/FerretDB/issues/1025
	ctx, targetCollections, compatCollections := setup.SetupCompat(t)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skipForTigris != "" {
				setup.SkipForTigrisWithReason(t, tc.skipForTigris)
			}

			t.Parallel()

			filter := tc.filter
			require.NotNil(t, filter, "filter should be set")

			opts := options.Find().SetSort(bson.D{{"_id", 1}})

			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					targetCursor, targetErr := targetCollection.Find(ctx, filter, opts)
					compatCursor, compatErr := compatCollection.Find(ctx, filter, opts)

					if targetCursor != nil {
						defer targetCursor.Close(ctx)
					}
					if compatCursor != nil {
						defer compatCursor.Close(ctx)
					}

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					targetRes := targetCollection.Database().RunCommand(ctx, bson.D{
						{"getMore", int64(1)},
						{"collection", targetCollection.Name()},
					})
					compatRes := compatCollection.Database().RunCommand(ctx, bson.D{
						{"getMore", compatCursor.ID()},
						{"collection", compatCollection.Name()},
					})

					if targetRes.Err() != nil {
						t.Logf("Target error: %v", targetRes.Err())
						AssertMatchesCommandError(t, compatRes.Err(), targetRes.Err())

						return
					}
				})
			}
		})
	}
}

func TestGetMoreCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]getMoreCompatTestCase{
		"getMore": {
			filter: bson.D{{"_id", bson.D{{"$gt", 0}}}},
		},
	}

	testGetMoreCompat(t, testCases)
}
