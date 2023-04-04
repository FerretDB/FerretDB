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

func TestAggregateCompatCollStats(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		skip       string                   // skip test for all handlers, must have issue number mentioned
		collStats  bson.D                   // required
		resultType compatTestCaseResultType // defaults to nonEmptyResult
	}{
		"NilCollStats": {
			collStats:  nil,
			resultType: emptyResult,
		},
		"EmptyCollStats": {
			collStats: bson.D{},
		},
		"Count": {
			collStats: bson.D{{"count", bson.D{}}},
		},
		"StorageStats": {
			collStats: bson.D{{"storageStats", bson.D{}}},
		},
		"StorageStatsWithScale": {
			collStats: bson.D{{"storageStats", bson.D{{"scale", 1000}}}},
		},
		"CountAndStorageStats": {
			collStats: bson.D{{"count", bson.D{}}, {"storageStats", bson.D{}}},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Helper()
			t.Parallel()

			// It's enough to use a couple of providers: one for some collection and one for a non-existent collection.
			s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
				Providers:                []shareddata.Provider{shareddata.ArrayDocuments},
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					command := bson.A{bson.D{{"$collStats", tc.collStats}}}

					targetCursor, targetErr := targetCollection.Aggregate(ctx, command)
					compatCursor, compatErr := compatCollection.Aggregate(ctx, command)

					if targetCursor != nil {
						defer targetCursor.Close(ctx)
					}
					if compatCursor != nil {
						defer compatCursor.Close(ctx)
					}

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					var targetRes, compatRes []bson.D
					require.NoError(t, targetCursor.All(ctx, &targetRes))
					require.NoError(t, compatCursor.All(ctx, &compatRes))

					// $collStats returns one document per shard.
					require.Equal(t, 1, len(compatRes))
					require.Equal(t, 1, len(targetRes))

					// Check the keys are the same
					targetKeys := CollectKeys(t, targetRes[0])
					compatKeys := CollectKeys(t, compatRes[0])

					require.Equal(t, compatKeys, targetKeys)

					if len(targetRes) > 0 || len(compatRes) > 0 {
						nonEmptyResults = true
					}

					// TODO Check the returned values when possible: https://github.com/FerretDB/FerretDB/issues/2349
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case emptyResult:
				assert.False(t, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}
