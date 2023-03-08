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
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// aggregateStagesCompatTestCase describes aggregation stages compatibility test case.
type aggregateStagesCompatTestCase struct {
	skip       string                   // skip test for all handlers, must have issue number mentioned
	pipeline   bson.A                   // required, unspecified $sort appends bson.D{{"$sort", bson.D{{"_id", 1}}}}
	resultType compatTestCaseResultType // defaults to nonEmptyResult
}

// testAggregateStagesCompat tests aggregation stages compatibility test cases.
func testAggregateStagesCompat(t *testing.T, testCases map[string]aggregateStagesCompatTestCase) {
	t.Helper()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			pipeline := tc.pipeline
			require.NotNil(t, pipeline, "pipeline should be set")

			var hasSortStage bool
			for _, stage := range pipeline {
				stage, ok := stage.(bson.D)
				if !ok {
					continue
				}

				if _, hasSortStage = stage.Map()["$sort"]; hasSortStage {
					break
				}
			}

			if !hasSortStage {
				// add sort stage to sort by _id because compat does not have deterministic order.
				pipeline = append(pipeline, bson.D{{"$sort", bson.D{{"_id", 1}}}})
			}

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					targetCursor, targetErr := targetCollection.Aggregate(ctx, pipeline)
					compatCursor, compatErr := compatCollection.Aggregate(ctx, pipeline)

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

					AssertEqualDocumentsSlice(t, compatRes, targetRes)

					if len(targetRes) > 0 || len(compatRes) > 0 {
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

// aggregateCommandCompatTestCase describes aggregate compatibility test case.
type aggregateCommandCompatTestCase struct {
	skip       string                   // skip test for all handlers, must have issue number mentioned
	command    bson.D                   // required
	resultType compatTestCaseResultType // defaults to nonEmptyResult
}

// testAggregateCommandCompat tests aggregate pipeline compatibility test cases using one collection.
// Use testAggregateStagesCompat for testing stages of aggregation.
func testAggregateCommandCompat(t *testing.T, testCases map[string]aggregateCommandCompatTestCase) {
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

			command := tc.command
			require.NotNil(t, command, "command should be set")

			var nonEmptyResults bool

			t.Run(targetCollection.Name(), func(t *testing.T) {
				t.Helper()

				var targetRes, compatRes []bson.D
				targetErr := targetCollection.Database().RunCommand(ctx, command).Decode(&targetRes)
				compatErr := compatCollection.Database().RunCommand(ctx, command).Decode(&compatRes)

				if targetErr != nil {
					t.Logf("Target error: %v", targetErr)
					t.Logf("Compat error: %v", compatErr)
					AssertMatchesCommandError(t, compatErr, targetErr)

					return
				}
				require.NoError(t, compatErr, "compat error; target returned no error")

				AssertEqualDocumentsSlice(t, compatRes, targetRes)

				if len(targetRes) > 0 || len(compatRes) > 0 {
					nonEmptyResults = true
				}
			})

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

func TestAggregateCommandCompat(t *testing.T) {
	testCases := map[string]aggregateCommandCompatTestCase{
		"CollectionAgnostic": {
			command: bson.D{
				{"aggregate", 1},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/1890",
		},
		"FailedToParse": {
			command: bson.D{
				{"aggregate", 2},
			},
			resultType: emptyResult,
		},
		"PipelineTypeMismatch": {
			command: bson.D{
				{"aggregate", "collection-name"},
				{"pipeline", 1},
			},
			resultType: emptyResult,
		},
		"StageTypeMismatch": {
			command: bson.D{
				{"aggregate", "collection-name"},
				{"pipeline", bson.A{1}},
			},
			resultType: emptyResult,
		},
	}

	testAggregateCommandCompat(t, testCases)
}

func TestAggregateCompatStages(t *testing.T) {
	testCases := map[string]aggregateStagesCompatTestCase{
		"MatchAndCount": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", 42}}}},
				bson.D{{"$count", "v"}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatCount(t *testing.T) {
	testCases := map[string]aggregateStagesCompatTestCase{
		"Value": {
			pipeline: bson.A{bson.D{{"$count", "v"}}},
		},
		"NonExistent": {
			pipeline: bson.A{bson.D{{"$count", "nonexistent"}}},
		},
		"Location15948": {
			pipeline:   bson.A{bson.D{{"$count", "_id"}}},
			resultType: emptyResult,
		},
		"Location40156": {
			pipeline:   bson.A{bson.D{{"$count", 1}}},
			resultType: emptyResult,
		},
		"Location40157": {
			pipeline:   bson.A{bson.D{{"$count", ""}}},
			resultType: emptyResult,
		},
		"Location40160": {
			pipeline:   bson.A{bson.D{{"$count", "v.foo"}}},
			resultType: emptyResult,
		},
		"Location40158": {
			pipeline:   bson.A{bson.D{{"$count", "$foo"}}},
			resultType: emptyResult,
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatGroup(t *testing.T) {
	testCases := map[string]aggregateStagesCompatTestCase{
		"NullID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
			}}}},
		},
		"DistinctID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v"},
			}}}},
			skip: "values with equal sort order can return in not deterministic order.",
		},
		"NonExistentID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$invalid"},
			}}}},
		},
		"NonExpressionID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "v"},
			}}}},
		},
		"NonStringID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", bson.A{}},
			}}}},
		},
		"OperatorNameAsExpression": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$add"},
			}}}},
		},
		"DotNotationID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.foo"},
			}}}},
		},
		"Location16872": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$"},
			}}}},
			resultType: emptyResult,
		},
		"Location15947": {
			pipeline:   bson.A{bson.D{{"$group", 1}}},
			resultType: emptyResult,
		},
		"Location15955": {
			pipeline:   bson.A{bson.D{{"$group", bson.D{}}}},
			resultType: emptyResult,
		},
		"MissingID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"bla", 1},
			}}}},
			resultType: emptyResult,
		},
		"InvalidAccumulator": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"v", 1},
			}}}},
			resultType: emptyResult,
		},
		"EmptyAccumulator": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"v", bson.D{}},
			}}}},
			resultType: emptyResult,
		},
		"Location40238": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"v", bson.D{{"$count", "v"}, {"$count", "v"}}},
			}}}},
			resultType: emptyResult,
		},
		"Location40234": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"v", bson.D{{"invalid", "v"}}},
			}}}},
			resultType: emptyResult,
			skip:       "once all accumulator is implemented, it should return error",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatGroupCount(t *testing.T) {
	testCases := map[string]aggregateStagesCompatTestCase{
		"Count": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
		},
		"CountID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$_id"},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
		},
		"CountValue": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v"},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatMatch(t *testing.T) {
	testCases := map[string]aggregateStagesCompatTestCase{
		"ID": {
			pipeline: bson.A{bson.D{{"$match", bson.D{{"_id", "string"}}}}},
		},
		"Int": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", 42}}}},
			},
		},
		"String": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
			},
		},
		"Document": {
			pipeline: bson.A{bson.D{{"$match", bson.D{{"v", bson.D{{"foo", int32(42)}}}}}}},
		},
		"Array": {
			pipeline: bson.A{bson.D{{"$match", bson.D{{"v", bson.A{int32(42)}}}}}},
		},
		"Regex": {
			pipeline: bson.A{bson.D{{"$match", bson.D{{"v", bson.D{{"$eq", primitive.Regex{Pattern: "foo", Options: "i"}}}}}}}},
		},
		"Empty": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{}}},
			},
		},
		"DotNotationDocument": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v.foo", int32(42)}}}},
			},
		},
		"DotNotationArray": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v.0", int32(42)}}}},
			},
		},
		"Location15959": {
			pipeline:   bson.A{bson.D{{"$match", 1}}},
			resultType: emptyResult,
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatSort(t *testing.T) {
	testCases := map[string]aggregateStagesCompatTestCase{
		"AscendingID": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{{"_id", 1}}}}},
		},
		"DescendingID": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{{"_id", -1}}}}},
		},
		"AscendingValue": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v", 1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
		},
		"DescendingValue": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v", -1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
		},
		"DotNotation": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v.foo", 1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
			skip: "https://github.com/FerretDB/FerretDB/issues/2101",
		},
		"DotNotationNonExistent": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"invalid.foo", 1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
		},
		"Location15973": {
			pipeline:   bson.A{bson.D{{"$sort", 1}}},
			resultType: emptyResult,
		},
		"Location15975": {
			pipeline:   bson.A{bson.D{{"$sort", bson.D{{"_id", 0}}}}},
			resultType: emptyResult,
		},
		"Location15976": {
			pipeline:   bson.A{bson.D{{"$sort", bson.D{}}}},
			resultType: emptyResult,
		},
	}

	testAggregateStagesCompat(t, testCases)
}
