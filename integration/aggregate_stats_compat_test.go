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

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                []shareddata.Provider{shareddata.ArrayDocuments},
		AddNonExistentCollection: true,
	})
	ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

	for name, tc := range map[string]struct {
		pipeline bson.A // required

		resultType compatTestCaseResultType // defaults to nonEmptyResult
		skip       string                   // skip test for all handlers, must have issue number mentioned
	}{
		"EmptyCollStats": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{}}}},
		},
		"Count": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{{"count", bson.D{}}}}}},
		},
		"StorageStats": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{}}}}}},
		},
		"StorageStatsWithScale": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{{"scale", 1000}}}}}}},
		},
		"StorageStatsFloatScale": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{{"storageStats", bson.D{{"scale", 42.42}}}}}}},
		},
		"CountAndStorageStats": {
			pipeline: bson.A{bson.D{{"$collStats", bson.D{{"count", bson.D{}}, {"storageStats", bson.D{}}}}}},
		},
		"CollStatsCount": {
			pipeline: bson.A{
				bson.D{{"$collStats", bson.D{{"count", bson.D{}}, {"storageStats", bson.D{}}}}},
				bson.D{{"$count", "after"}},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.pipeline, "pipeline must be set")

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					targetCursor, targetErr := targetCollection.Aggregate(ctx, tc.pipeline)
					compatCursor, compatErr := compatCollection.Aggregate(ctx, tc.pipeline)

					if targetCursor != nil {
						defer targetCursor.Close(ctx)
					}
					if compatCursor != nil {
						defer compatCursor.Close(ctx)
					}

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)

						// error messages are intentionally not compared
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					targetRes := FetchAll(t, ctx, targetCursor)
					compatRes := FetchAll(t, ctx, compatCursor)

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

					targetDoc := ConvertDocument(t, targetRes[0])
					compatDoc := ConvertDocument(t, compatRes[0])

					targetNs, _ := targetDoc.Get("ns")
					compatNs, _ := compatDoc.Get("ns")
					require.Equal(t, compatNs, targetNs)

					targetCount, _ := targetDoc.Get("count")
					compatCount, _ := compatDoc.Get("count")
					require.EqualValues(t, compatCount, targetCount)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be returned)")
			case emptyResult:
				assert.False(t, nonEmptyResults, "expected empty results (no documents should be returned)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}
