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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration/setup"
)

// queryCompatTestCase describes query compatibility test case.
//
//nolint:vet // for readability
type queryCompatCommandTestCase struct {
	filter     bson.D // required
	sort       bson.D // defaults to `bson.D{{"_id", 1}}`
	projection bson.D // nil for leaving projection unset

	limit          *int64                   // defaults to nil to leave unset
	optSkip        any                      // defaults to nil to leave unset
	resultType     compatTestCaseResultType // defaults to nonEmptyResult
	resultPushdown bool                     // defaults to false

	skip          string // skip test for all handlers, must have issue number mentioned
	skipForTigris string // skip test for Tigris
}

// testQueryCompat tests query compatibility test cases.
func testQueryCompatCommand(t *testing.T, testCases map[string]queryCompatCommandTestCase) {
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

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			var rest bson.D
			if tc.sort != nil {
				rest = append(rest, bson.E{Key: "sort", Value: tc.sort})
			} else {
				rest = append(rest, bson.E{Key: "sort", Value: bson.D{{"_id", 1}}})
			}

			if tc.limit != nil {
				rest = append(rest, bson.E{Key: "limit", Value: *tc.limit})
			}

			if tc.optSkip != nil {
				rest = append(rest, bson.E{Key: "skip", Value: tc.optSkip})
			}

			if tc.projection != nil {
				rest = append(rest, bson.E{Key: "projection", Value: tc.projection})
			}

			t.Parallel()

			filter := tc.filter
			require.NotNil(t, filter, "filter should be set")

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					// don't add sort, limit, skip, and projection because we don't pushdown them yet
					explainQuery := bson.D{{"explain", bson.D{
						{"find", targetCollection.Name()},
						{"filter", filter},
					}}}

					var explainRes bson.D
					require.NoError(t, targetCollection.Database().RunCommand(ctx, explainQuery).Decode(&explainRes))

					var msg string
					if setup.IsPushdownDisabled() {
						tc.resultPushdown = false
						msg = "Query pushdown is disabled, but target resulted with pushdown"
					}

					assert.Equal(t, tc.resultPushdown, explainRes.Map()["pushdown"], msg)

					targetCommand := append(
						bson.D{
							{"find", targetCollection.Name()},
							{"filter", filter},
						},
						rest...,
					)
					compatCommand := append(
						bson.D{
							{"find", compatCollection.Name()},
							{"filter", filter},
						},
						rest...,
					)

					targetResult := targetCollection.Database().RunCommand(ctx, targetCommand)
					compatResult := compatCollection.Database().RunCommand(ctx, compatCommand)

					targetErr := targetResult.Err()
					compatErr := compatResult.Err()

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					var targetRes, compatRes bson.D
					require.NoError(t, targetResult.Decode(&targetRes))
					require.NoError(t, compatResult.Decode(&compatRes))

					AssertEqualDocuments(t, compatRes, targetRes)

					targetDocs := targetRes.Map()["cursor"].(bson.D).Map()["firstBatch"].(primitive.A)
					compatDocs := compatRes.Map()["cursor"].(bson.D).Map()["firstBatch"].(primitive.A)

					if len(targetDocs) > 0 || len(compatDocs) > 0 {
						nonEmptyResults = true
					}
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results")
			case emptyResult:
				assert.False(t, nonEmptyResults, "expected empty results")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

func TestQueryCompatCommandSkip(t *testing.T) {
	testCases := map[string]queryCompatCommandTestCase{
		"MaxInt64": {
			filter:     bson.D{},
			optSkip:    math.MaxInt64,
			resultType: emptyResult,
		},
		"Int64Overflow": {
			filter:     bson.D{},
			optSkip:    float64(1 << 86),
			resultType: emptyResult,
		},
		"NegativeInt64": {
			filter:     bson.D{},
			optSkip:    int64(-2),
			resultType: emptyResult,
		},
		"NegativeFloat64": {
			filter:     bson.D{},
			optSkip:    -2.8,
			resultType: emptyResult,
		},
		"Float64": {
			filter:  bson.D{},
			optSkip: 2.8,
		},
		"Float64Ceil": {
			filter:  bson.D{},
			optSkip: 2.1,
		},
	}

	testQueryCompatCommand(t, testCases)
}
