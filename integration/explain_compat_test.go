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

// explainCompatTestCase describes explain compatibility test case.
type explainCompatTestCase struct {
	command    string                   // required
	filter     bson.D                   // ignored if nil
	pipeline   bson.A                   // ignored if nil
	resultType compatTestCaseResultType // defaults to nonEmptyResult

	skip string // skip test for all handlers, must have issue number mentioned
}

// testExplainCompatError tests explain compatibility test cases.
// This test does not work for successful aggregate pipeline tests,
// due to compat requiring cursor option.
// If you see following error, use `testAggregateStagesCompat` test instead.
//
//	`(FailedToParse) The 'cursor' option is required, except for aggregate with the explain argument`
func testExplainCompatError(t *testing.T, testCases map[string]explainCompatTestCase) {
	t.Helper()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		// Use a provider that works for all handlers.
		Providers: []shareddata.Provider{shareddata.Int32s},
	})

	ctx := s.Ctx
	targetCollection := s.TargetCollections[0]
	compatCollection := s.CompatCollections[0]

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			t.Run(targetCollection.Name(), func(t *testing.T) {
				t.Helper()

				explainTarget := bson.D{{tc.command, targetCollection.Name()}}
				explainCompat := bson.D{{tc.command, compatCollection.Name()}}

				if tc.filter != nil {
					explainTarget = append(explainTarget, bson.E{Key: "filter", Value: tc.filter})
					explainCompat = append(explainCompat, bson.E{Key: "filter", Value: tc.filter})
				}

				if tc.pipeline != nil {
					explainTarget = append(explainTarget, bson.E{Key: "pipeline", Value: tc.pipeline})
					explainCompat = append(explainCompat, bson.E{Key: "pipeline", Value: tc.pipeline})
				}

				var targetRes, compatRes bson.D
				targetErr := targetCollection.Database().RunCommand(
					ctx,
					bson.D{{"explain", explainTarget}},
				).Decode(&targetRes)
				compatErr := compatCollection.Database().RunCommand(
					ctx,
					bson.D{{"explain", explainCompat}},
				).Decode(&compatRes)

				if targetErr != nil {
					t.Logf("Target error: %v", targetErr)
					t.Logf("Compat error: %v", compatErr)
					AssertMatchesCommandError(t, compatErr, targetErr)

					return
				}
				require.NoError(t, compatErr, "compat error; target returned no error")

				targetMap := targetRes.Map()
				compatMap := compatRes.Map()

				// check that the response of ok and command are the same
				// only check these field because other field such as version
				// different in compat and target
				assert.Equal(t, compatMap["ok"], targetMap["ok"])
				assert.Equal(t, compatMap["command"], targetMap["command"])

				// check queryPlanner is set
				assert.NotEmpty(t, targetMap["queryPlanner"])

				var nonEmptyResults bool
				if len(targetRes) > 0 || len(compatRes) > 0 {
					nonEmptyResults = true
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
		})
	}
}

func TestExplainCompatError(t *testing.T) {
	t.Parallel()

	testCases := map[string]explainCompatTestCase{
		"AggregateMissingPipeline": {
			command: "aggregate",
		},
		"AggregateInvalidPipeline": {
			command:  "aggregate",
			pipeline: bson.A{1},
		},
		"Count": {
			command: "count",
		},
		"Find": {
			command: "find",
			filter:  bson.D{{"v", int32(42)}},
		},
	}

	testExplainCompatError(t, testCases)
}
