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
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

func TestAggregateCompatCollStats(tt *testing.T) {
	tt.Parallel()

	for name, tc := range map[string]struct {
		skip           string                   // skip test for all handlers, must have issue number mentioned
		collStats      bson.D                   // required
		resultType     compatTestCaseResultType // defaults to nonEmptyResult
		failsForSQLite string                   // non-empty value expects test to fail for SQLite backend
	}{
		"NilCollStats": {
			collStats:  nil,
			resultType: emptyResult,
		},
		"EmptyCollStats": {
			collStats:      bson.D{},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3259",
		},
		"Count": {
			collStats:      bson.D{{"count", bson.D{}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3259",
		},
		"StorageStats": {
			collStats:      bson.D{{"storageStats", bson.D{}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3259",
		},
		"StorageStatsWithScale": {
			collStats:      bson.D{{"storageStats", bson.D{{"scale", 1000}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3259",
		},
		"StorageStatsNegativeScale": {
			collStats:  bson.D{{"storageStats", bson.D{{"scale", -1000}}}},
			resultType: emptyResult,
		},
		"StorageStatsFloatScale": {
			collStats:      bson.D{{"storageStats", bson.D{{"scale", 42.42}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3259",
		},
		"StorageStatsInvalidScale": {
			collStats:  bson.D{{"storageStats", bson.D{{"scale", "invalid"}}}},
			resultType: emptyResult,
		},
		"CountAndStorageStats": {
			collStats:      bson.D{{"count", bson.D{}}, {"storageStats", bson.D{}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3259",
		},
	} {
		name, tc := name, tc
		tt.Run(name, func(tt *testing.T) {
			if tc.skip != "" {
				tt.Skip(tc.skip)
			}

			tt.Helper()
			tt.Parallel()

			// It's enough to use a couple of providers: one for some collection and one for a non-existent collection.
			s := setup.SetupCompatWithOpts(tt, &setup.SetupCompatOpts{
				Providers:                []shareddata.Provider{shareddata.ArrayDocuments},
				AddNonExistentCollection: true,
			})
			ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				tt.Run(targetCollection.Name(), func(tt *testing.T) {
					tt.Helper()

					var t testtb.TB = tt
					if tc.failsForSQLite != "" {
						t = setup.FailsForSQLite(tt, tc.failsForSQLite)
					}

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

					// Check the returned values when possible
					// TODO https://github.com/FerretDB/FerretDB/issues/2349
				})
			}

			// TODO https://github.com/FerretDB/FerretDB/issues/3259
			if setup.IsSQLite(tt) {
				return
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(tt, nonEmptyResults, "expected non-empty results (some documents should be modified)")
			case emptyResult:
				assert.False(tt, nonEmptyResults, "expected empty results (no documents should be modified)")
			default:
				tt.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}
