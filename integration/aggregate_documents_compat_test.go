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
	"time"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

// aggregateStagesCompatTestCase describes aggregation stages compatibility test case.
type aggregateStagesCompatTestCase struct {
	pipeline bson.A         // required, unspecified $sort appends bson.D{{"$sort", bson.D{{"_id", 1}}}} for non empty pipeline.
	maxTime  *time.Duration // optional, leave nil for unset maxTime

	resultType     compatTestCaseResultType // defaults to nonEmptyResult
	resultPushdown bool                     // defaults to false
	skip           string                   // skip test for all handlers, must have issue number mentioned
	failsForSQLite string                   // optional, if set, the case is expected to fail for SQLite due to given issue
}

// testAggregateStagesCompat tests aggregation stages compatibility test cases with all providers.
func testAggregateStagesCompat(t *testing.T, testCases map[string]aggregateStagesCompatTestCase) {
	t.Helper()

	testAggregateStagesCompatWithProviders(t, shareddata.AllProviders(), testCases)
}

// testAggregateStagesCompatWithProviders tests aggregation stages compatibility test cases with given providers.
func testAggregateStagesCompatWithProviders(tt *testing.T, providers shareddata.Providers, testCases map[string]aggregateStagesCompatTestCase) {
	tt.Helper()

	require.NotEmpty(tt, providers)

	s := setup.SetupCompatWithOpts(tt, &setup.SetupCompatOpts{
		Providers: providers,
	})
	ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

	for name, tc := range testCases {
		name, tc := name, tc
		tt.Run(name, func(tt *testing.T) {
			tt.Helper()

			if tc.skip != "" {
				tt.Skip(tc.skip)
			}

			tt.Parallel()

			pipeline := tc.pipeline
			require.NotNil(tt, pipeline, "pipeline should be set")

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

			if !hasSortStage && len(pipeline) > 0 {
				// add sort stage to sort by _id because compat and target
				// would be ordered differently otherwise.
				pipeline = append(pipeline, bson.D{{"$sort", bson.D{{"_id", 1}}}})
			}

			opts := options.Aggregate()

			if tc.maxTime != nil {
				opts.SetMaxTime(*tc.maxTime)
			}

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				tt.Run(targetCollection.Name(), func(tt *testing.T) {
					tt.Helper()

					var t testtb.TB = tt //nolint:vet // that's intentional
					if tc.failsForSQLite != "" {
						t = setup.FailsForSQLite(tt, tc.failsForSQLite)
					}

					explainCommand := bson.D{{"explain", bson.D{
						{"aggregate", targetCollection.Name()},
						{"pipeline", pipeline},
					}}}
					var explainRes bson.D
					require.NoError(t, targetCollection.Database().RunCommand(ctx, explainCommand).Decode(&explainRes))

					var msg string
					if setup.IsPushdownDisabled() {
						tc.resultPushdown = false
						msg = "Query pushdown is disabled, but target resulted with pushdown"
					}

					assert.Equal(t, tc.resultPushdown, explainRes.Map()["pushdown"], msg)

					targetCursor, targetErr := targetCollection.Aggregate(ctx, pipeline, opts)
					compatCursor, compatErr := compatCollection.Aggregate(ctx, pipeline, opts)

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

					AssertEqualDocumentsSlice(t, compatRes, targetRes)

					if len(targetRes) > 0 || len(compatRes) > 0 {
						nonEmptyResults = true
					}
				})
			}

			if tc.failsForSQLite != "" && setup.IsSQLite(tt) {
				return
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(tt, nonEmptyResults, "expected non-empty results")
			case emptyResult:
				assert.False(tt, nonEmptyResults, "expected empty results")
			default:
				tt.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

// aggregateCommandCompatTestCase describes aggregate compatibility test case.
type aggregateCommandCompatTestCase struct {
	command    bson.D                   // required
	resultType compatTestCaseResultType // defaults to nonEmptyResult

	skip           string // skip test for all handlers, must have issue number mentioned
	failsForSQLite string
}

// testAggregateCommandCompat tests aggregate pipeline compatibility test cases using one collection.
// Use testAggregateStagesCompat for testing stages of aggregation.
func testAggregateCommandCompat(tt *testing.T, testCases map[string]aggregateCommandCompatTestCase) {
	tt.Helper()

	s := setup.SetupCompatWithOpts(tt, &setup.SetupCompatOpts{
		// Use a provider that works for all handlers.
		Providers: []shareddata.Provider{shareddata.Int32s},
	})

	ctx := s.Ctx
	targetCollection := s.TargetCollections[0]
	compatCollection := s.CompatCollections[0]

	for name, tc := range testCases {
		name, tc := name, tc
		tt.Run(name, func(tt *testing.T) {
			tt.Helper()

			if tc.skip != "" {
				tt.Skip(tc.skip)
			}

			tt.Parallel()

			command := tc.command
			require.NotNil(tt, command, "command should be set")

			var nonEmptyResults bool

			tt.Run(targetCollection.Name(), func(tt *testing.T) {
				tt.Helper()

				var t testtb.TB = tt //nolint:vet // that's intentional
				if tc.failsForSQLite != "" {
					t = setup.FailsForSQLite(tt, tc.failsForSQLite)
				}

				var targetRes, compatRes bson.D
				targetErr := targetCollection.Database().RunCommand(ctx, command).Decode(&targetRes)
				compatErr := compatCollection.Database().RunCommand(ctx, command).Decode(&compatRes)

				if targetErr != nil {
					t.Logf("Target error: %v", targetErr)
					t.Logf("Compat error: %v", compatErr)

					if _, ok := targetErr.(mongo.CommandError); ok { //nolint:errorlint // do not inspect error chain
						// error messages are intentionally not compared
						AssertMatchesCommandError(t, compatErr, targetErr)
					} else {
						// driver sent an error
						require.Equal(t, compatErr, targetErr)
					}

					return
				}
				require.NoError(t, compatErr, "compat error; target returned no error")
				AssertEqualDocuments(t, compatRes, targetRes)

				if len(targetRes) > 0 || len(compatRes) > 0 {
					nonEmptyResults = true
				}
			})

			if tc.failsForSQLite != "" && setup.IsSQLite(tt) {
				return
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(tt, nonEmptyResults, "expected non-empty results")
			case emptyResult:
				assert.False(tt, nonEmptyResults, "expected empty results")
			default:
				tt.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}

func TestAggregateCommandCompat(t *testing.T) {
	t.Parallel()

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
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"PipelineTypeMismatch": {
			command: bson.D{
				{"aggregate", "collection-name"},
				{"pipeline", 1},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"StageTypeMismatch": {
			command: bson.D{
				{"aggregate", "collection-name"},
				{"pipeline", bson.A{1}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidStage": {
			command: bson.D{
				{"aggregate", "collection-name"},
				{"pipeline", bson.A{"$invalid-stage"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MaxTimeMSDoubleWholeNumber": {
			command: bson.D{
				{"aggregate", "collection-name"},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", float64(1000)},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateCommandCompat(t, testCases)
}

func TestAggregateCompatOptions(t *testing.T) {
	t.Parallel()

	providers := []shareddata.Provider{
		// one provider is sufficient to test aggregate options
		shareddata.Unsets,
	}

	testCases := map[string]aggregateStagesCompatTestCase{
		"MaxTimeZero": {
			pipeline:       bson.A{},
			maxTime:        pointer.ToDuration(time.Duration(0)),
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MaxTime": {
			pipeline:       bson.A{},
			maxTime:        pointer.ToDuration(time.Second),
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatStages(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"MatchAndCount": {
			pipeline: bson.A{
				// when $match is the first stage, pushdown is done.
				bson.D{{"$match", bson.D{{"v", 42}}}},
				bson.D{{"$count", "v"}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			resultPushdown: true,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"CountAndMatch": {
			pipeline: bson.A{
				// when $match is the second stage, no pushdown is done.
				bson.D{{"$count", "v"}},
				bson.D{{"$match", bson.D{{"v", 1}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatEmptyPipeline(t *testing.T) {
	t.Parallel()

	providers := []shareddata.Provider{
		// for testing empty pipeline use a collection with single document,
		// because sorting will not matter.
		shareddata.Unsets,
	}

	testCases := map[string]aggregateStagesCompatTestCase{
		"Empty": {
			pipeline:       bson.A{},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatCount(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"Value": {
			pipeline:       bson.A{bson.D{{"$count", "v"}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExistent": {
			pipeline:       bson.A{bson.D{{"$count", "nonexistent"}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"CountGroupID": {
			pipeline:       bson.A{bson.D{{"$count", "_id"}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"CountNonString": {
			pipeline:       bson.A{bson.D{{"$count", 1}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"CountEmpty": {
			pipeline:       bson.A{bson.D{{"$count", ""}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"CountBadValue": {
			pipeline:       bson.A{bson.D{{"$count", "v.foo"}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"CountBadPrefix": {
			pipeline:       bson.A{bson.D{{"$count", "$foo"}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatGroupDeterministicCollections(t *testing.T) {
	t.Parallel()

	// Composites, ArrayStrings, ArrayInt32s, ArrayAndDocuments and Mixed are not included
	// because the order in compat and target can be not deterministic.
	// Aggregation assigns BSON array to output _id, and an array with
	// descending sort use the greatest element for comparison causing
	// multiple documents with the same greatest element the same order,
	// so compat and target results in different order.
	// https://github.com/FerretDB/FerretDB/issues/2185

	providers := shareddata.AllProviders().Remove(shareddata.Composites, shareddata.ArrayStrings, shareddata.ArrayInt32s, shareddata.ArrayAndDocuments, shareddata.Mixed)
	testCases := map[string]aggregateStagesCompatTestCase{
		"DistinctValue": {
			pipeline: bson.A{
				// sort to assure the same type of values (while grouping 2 types with the same value,
				// the first type in collection is chosen)
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
				}}},
				// sort descending order, so ArrayDoubles has deterministic order.
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Distinct": {
			pipeline: bson.A{
				// sort collection to ensure the order is consistent
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					// set first _id of the collection as group's unique value
					{"unique", bson.D{{"$first", "$_id"}}},
				}}},
				// ensure output is ordered by the _id of the collection, not _id of the group
				// because _id of group can be an array
				bson.D{{"$sort", bson.D{{"unique", 1}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2185",
		},
		"CountValue": {
			pipeline: bson.A{
				// sort to assure the same type of values (while grouping 2 types with the same value,
				// the first type in collection is chosen)
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"count", bson.D{{"$count", bson.D{}}}},
				}}},
				// sort descending order, so ArrayDoubles has deterministic order.
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},

		"LimitAfter": {
			pipeline: bson.A{
				// sort to assure the same type of values (while grouping 2 types with the same value,
				// the first type in collection is chosen)
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$group", bson.D{{"_id", "$v"}}}},
				// sort descending order, so ArrayDoubles has deterministic order.
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 5}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"LimitBefore": {
			pipeline: bson.A{
				// sort to assure the same type of values (while grouping 2 types with the same value,
				// the first type in collection is chosen)
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 5}},
				bson.D{{"$group", bson.D{{"_id", "$v"}}}},
				// sort descending order, so ArrayDoubles has deterministic order.
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"SkipAfter": {
			pipeline: bson.A{
				// sort to assure the same type of values (while grouping 2 types with the same value,
				// the first type in collection is chosen)
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$group", bson.D{{"_id", "$v"}}}},
				// sort descending order, so ArrayDoubles has deterministic order.
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", 2}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"SkipBefore": {
			pipeline: bson.A{
				// the first type in collection is chosen)
				// sort to assure the same type of values (while grouping 2 types with the same value,
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", 2}},
				bson.D{{"$group", bson.D{{"_id", "$v"}}}},
				// sort descending order, so ArrayDoubles has deterministic order.
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatGroup(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"NullID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DistinctID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$_id"},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IDExpression": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", bson.D{{"v", "$v"}}},
			}}}},
			skip:           "https://github.com/FerretDB/FerretDB/issues/2165",
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExistentID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$invalid"},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExpressionID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "v"},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonStringID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", bson.A{}},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"OperatorNameAsExpression": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$add"},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyPath": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$"},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyVariable": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$"},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidVariable$": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$$"},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidVariable$s": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$$s"},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExistingVariable": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$s"},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"SystemVariable": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$NOW"},
			}}}},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/2275",
		},
		"GroupInvalidFields": {
			pipeline:       bson.A{bson.D{{"$group", 1}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyGroup": {
			pipeline:       bson.A{bson.D{{"$group", bson.D{}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MissingID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"bla", 1},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidAccumulator": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"v", 1},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyAccumulator": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"v", bson.D{}},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"GroupMultipleAccumulator": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"v", bson.D{{"$count", "v"}, {"$count", "v"}}},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"GroupInvalidAccumulator": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"v", bson.D{{"invalid", "v"}}},
			}}}},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/2123",
		},
		"IDType": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", bson.D{{"$type", "$v"}}},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IDSum": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$sum", "$v"}}}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IDSumNonExistentField": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$sum", "$non-existent"}}}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IDSumInvalid": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$sum", "$"}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IDSumRecursiveInvalid": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$sum", bson.D{{"$sum", "$"}}}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatGroupExpressionDottedFields(t *testing.T) {
	t.Parallel()

	// TODO Use all providers after fixing $sort problem:  https://github.com/FerretDB/FerretDB/issues/2276.
	//
	// Currently, providers Composites, DocumentsDeeplyNested, ArrayAndDocuments and Mixed
	// cannot be used due to sorting difference.
	// FerretDB always sorts empty array is less than null.
	// In compat, for `.sort()` an empty array is less than null.
	// In compat, for aggregation `$sort` null is less than an empty array.
	providers := shareddata.AllProviders().Remove(shareddata.Mixed, shareddata.Composites, shareddata.DocumentsDeeplyNested, shareddata.ArrayAndDocuments)

	testCases := map[string]aggregateStagesCompatTestCase{
		"NestedInDocument": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.foo"},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DeeplyNested": { // Expect non-empty results for DocumentsDeeplyNested provider
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.a.b.c"},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayIndex": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.0"},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NestedInArray": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.0.foo"},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExistentParent": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$non.existent"},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExistentChild": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.non.existent"},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatGroupExpressionDottedFieldsDocs(t *testing.T) {
	t.Parallel()

	// TODO Merge the current function with TestAggregateCompatGroupExpressionDottedFields
	// and use all providers when $sort problem is fixed:
	// https://github.com/FerretDB/FerretDB/issues/2276

	providers := []shareddata.Provider{
		shareddata.DocumentsDeeplyNested,
	}

	testCases := map[string]aggregateStagesCompatTestCase{
		"DeeplyNested": { // Expect non-empty results for DocumentsDeeplyNested provider
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.a.b.c"},
			}}}},
			skip: "https://github.com/FerretDB/FerretDB/issues/2276",
			// compat results [ { _id: null }, { _id: 12 }, { _id: { d: 123 } }, { _id: [ 1, 2 ] } ]
			// target results [ { _id: null }, { _id: [ 1, 2 ] }, { _id: 12 }, { _id: { d: 123 } } ]
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatGroupCount(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"CountNull": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"CountID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$_id"},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeMismatch": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"count", bson.D{{"$count", ""}}},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonEmptyExpression": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"count", bson.D{{"$count", bson.D{{"a", 1}}}}},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExistentField": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$nonexistent"},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Duplicate": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v"},
				{"count", bson.D{{"$count", bson.D{}}}},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatLimit(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"Zero": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 0}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"One": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 1}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Five": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 5}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"StringInt": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", "5"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Double": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 4.5}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DoubleInt": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 5.0}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MaxInt64": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", math.MaxInt64}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MinInt64": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", math.MinInt64}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Negative": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", -1}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NegativeDouble": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", -2.1}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Document": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", bson.D{}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Int64Overflow": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", float64(1 << 86)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"AfterMatch": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
				bson.D{{"$limit", 1}},
			},
			resultPushdown: true, // $sort and $match are first two stages
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"BeforeMatch": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 1}},
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NoSortAfterMatch": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
				bson.D{{"$limit", 100}},
			},
			resultPushdown: true,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NoSortBeforeMatch": {
			pipeline: bson.A{
				bson.D{{"$limit", 100}},
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatGroupSum(t *testing.T) {
	t.Parallel()

	providers := shareddata.AllProviders().
		// skipped due to https://github.com/FerretDB/FerretDB/issues/2185.
		Remove(shareddata.Composites).
		Remove(shareddata.ArrayStrings).
		Remove(shareddata.ArrayInt32s).
		Remove(shareddata.Mixed).
		Remove(shareddata.ArrayAndDocuments).
		// TODO: handle $sum of doubles near max precision.
		// https://github.com/FerretDB/FerretDB/issues/2300
		Remove(shareddata.Doubles).
		// TODO: https://github.com/FerretDB/FerretDB/issues/2616
		Remove(shareddata.ArrayDocuments)

	testCases := map[string]aggregateStagesCompatTestCase{
		"GroupNullID": {
			pipeline: bson.A{
				// Without $sort, the sum of large values results different in compat and target.
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", nil},
					{"sum", bson.D{{"$sum", "$v"}}},
				}}},
				// Without $sort, documents are ordered not the same.
				// Descending sort is used because it is more unique than
				// ascending sort for shareddata collections.
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"GroupByID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$_id"},
					{"sum", bson.D{{"$sum", "$v"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"GroupByValue": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", "$v"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyString": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", ""}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExpression": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", nil},
					{"sum", bson.D{{"$sum", "v"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", "$non-existent"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Document": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},

				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", bson.D{}}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Array": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", bson.A{"$v", "$c"}}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Int32": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", int32(1)}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MaxInt32": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", math.MaxInt32}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NegativeInt32": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", int32(-1)}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Int64": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", int64(20)}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MaxInt64": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", math.MaxInt64}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Double": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", 43.7}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MaxDouble": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", math.MaxFloat64}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Bool": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", true}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Duplicate": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$v"},
					{"sum", bson.D{{"$sum", "$v"}}},
					{"sum", bson.D{{"$sum", "$s"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"RecursiveOperator": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", "$_id"},
					// first $sum is accumulator operator, second $sum is operator
					{"sum", bson.D{{"$sum", bson.D{{"$sum", "$v"}}}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"RecursiveInvalid": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"sum", bson.D{{"$sum", bson.D{{"v", "$v"}}}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"RecursiveArrayInvalid": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"sum", bson.D{{"$sum", bson.D{{"$type", bson.A{"1", "2"}}}}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"RecursiveOperatorNonExistent": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", "$_id"},
					// first $sum is accumulator operator, second $sum is operator
					{"sum", bson.D{{"$sum", bson.D{{"$non-existent", "$v"}}}}},
				}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatMatch(t *testing.T) {
	t.Parallel()

	// TODO https://github.com/FerretDB/FerretDB/issues/2291
	providers := shareddata.AllProviders().Remove(shareddata.ArrayAndDocuments)

	testCases := map[string]aggregateStagesCompatTestCase{
		"ID": {
			pipeline:       bson.A{bson.D{{"$match", bson.D{{"_id", "string"}}}}},
			resultPushdown: true,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Int": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", 42}}}},
			},
			resultPushdown: true,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"String": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
			},
			resultPushdown: true,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Document": {
			pipeline:       bson.A{bson.D{{"$match", bson.D{{"v", bson.D{{"foo", int32(42)}}}}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Array": {
			pipeline:       bson.A{bson.D{{"$match", bson.D{{"v", bson.A{int32(42)}}}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Regex": {
			pipeline:       bson.A{bson.D{{"$match", bson.D{{"v", bson.D{{"$eq", primitive.Regex{Pattern: "foo", Options: "i"}}}}}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Empty": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationDocument": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v.foo", int32(42)}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationArray": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v.0", int32(42)}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MatchBadValue": {
			pipeline:       bson.A{bson.D{{"$match", 1}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"SumValue": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$sum", "$v"}}}}}},
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/414",
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatSort(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"AscendingID": {
			pipeline:       bson.A{bson.D{{"$sort", bson.D{{"_id", 1}}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DescendingID": {
			pipeline:       bson.A{bson.D{{"$sort", bson.D{{"_id", -1}}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"AscendingValue": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v", 1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DescendingValue": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v", -1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"AscendingValueDescendingID": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v", 1},
				{"_id", -1},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DescendingValueDescendingID": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v", -1},
				{"_id", -1},
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},

		"DotNotationIndex": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v.0", 1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationNonExistent": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"invalid.foo", 1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationMissingField": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v..foo", 1},
			}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},

		"SortBadExpression": {
			pipeline:       bson.A{bson.D{{"$sort", 1}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"SortBadOrder": {
			pipeline:       bson.A{bson.D{{"$sort", bson.D{{"_id", 0}}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"SortMissingKey": {
			pipeline:       bson.A{bson.D{{"$sort", bson.D{}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"BadDollarStart": {
			pipeline:       bson.A{bson.D{{"$sort", bson.D{{"$v.foo", 1}}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatSortDotNotation(t *testing.T) {
	t.Parallel()

	providers := shareddata.AllProviders().
		// TODO: https://github.com/FerretDB/FerretDB/issues/2617
		Remove(shareddata.ArrayDocuments)

	testCases := map[string]aggregateStagesCompatTestCase{
		"DotNotation": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v.foo", 1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatUnwind(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"Simple": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$v"}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$non-existent"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Invalid": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "invalid"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$v.foo"}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationNonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$v.non-existent"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayDotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$v.0"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayDotNotationKey": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$v.0.foo"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Null": {
			pipeline:       bson.A{bson.D{{"$unwind", nil}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$_id"}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NameAsExpression": {
			pipeline:       bson.A{bson.D{{"$unwind", "$add"}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyPath": {
			pipeline:       bson.A{bson.D{{"$unwind", "$"}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyVariable": {
			pipeline:       bson.A{bson.D{{"$unwind", "$$"}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidVariable$": {
			pipeline:       bson.A{bson.D{{"$unwind", "$$$"}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidVariable$s": {
			pipeline:       bson.A{bson.D{{"$unwind", "$$$s"}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NonExistingVariable": {
			pipeline:       bson.A{bson.D{{"$unwind", "$$s"}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"SystemVariable": {
			pipeline:       bson.A{bson.D{{"$unwind", "$$NOW"}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Empty": {
			pipeline:       bson.A{bson.D{{"$unwind", ""}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Number": {
			pipeline:       bson.A{bson.D{{"$unwind", 42}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Array": {
			pipeline:       bson.A{bson.D{{"$unwind", bson.A{"$v"}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatSkip(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"Document": {
			pipeline:       bson.A{bson.D{{"$skip", bson.D{}}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Zero": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", int32(0)}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"One": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", int32(1)}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"SkipAll": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", int32(1000)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"StringInt": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", "1"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NegativeValue": {
			pipeline:       bson.A{bson.D{{"$skip", int32(-1)}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"NegativeDouble": {
			pipeline:       bson.A{bson.D{{"$skip", -3.2}}},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MaxInt64": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", math.MaxInt64}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MinInt64": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", math.MinInt64}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Int64Overflow": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", float64(1 << 86)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"AfterMatch": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
				bson.D{{"$skip", int32(1)}},
			},
			resultPushdown: true, // $match after $sort can be pushed down
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"BeforeMatch": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", int32(1)}},
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatProject(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"InvalidType": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", "invalid"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ZeroOperators": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"v", bson.D{}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TwoOperators": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"v", bson.D{{"$type", "foo"}, {"$sum", 1}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DollarSignField": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"$v", 1}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Include1Field": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"v", int32(1)}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Exclude1Field": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"v", int64(0)}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IncludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", 1.42}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ExcludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", false}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Include2Fields": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"foo", 1.24}, {"bar", true}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Exclude2Fields": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},

				bson.D{{"$project", bson.D{{"foo", int32(0)}, {"bar", false}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Include1FieldExclude1Field": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},

				bson.D{{"$project", bson.D{{"foo", int32(0)}, {"bar", true}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Exclude1FieldInclude1Field": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},

				bson.D{{"$project", bson.D{{"foo", int32(1)}, {"bar", false}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IncludeFieldExcludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", false}, {"v", true}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ExcludeFieldIncludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", true}, {"v", false}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ExcludeFieldExcludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", false}, {"v", false}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IncludeFieldIncludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", true}, {"v", true}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Assign1Field": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"foo", primitive.NewObjectID()}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"AssignID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Assign1FieldIncludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", true}, {"foo", primitive.NewDateTimeFromTime(time.Unix(0, 0))}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Assign2FieldsIncludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", true}, {"foo", nil}, {"bar", "qux"}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Assign1FieldExcludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", false}, {"foo", primitive.Regex{Pattern: "^fo"}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationInclude": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.foo", true},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationIncludeTwo": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.foo", true},
					{"v.array", true},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationExclude": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.foo", false},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationExcludeTwo": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.foo", false},
					{"v.array", false},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationExcludeSecondLevel": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.array.42", false},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationIncludeExclude": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.foo", true},
					{"v.array.42", false},
				}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyDocument": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"foo", bson.D{}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Document": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"foo", bson.D{{"v", "foo"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IDDocument": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", bson.D{{"v", "foo"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IDType": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", bson.D{{"$type", "$v"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DocumentAndValue": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{
					{"foo", bson.D{{"v", "foo"}}},
					{"v", 1},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Type": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", "$v"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeNonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", "$foo"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeDotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", "$v.foo"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeRecursive": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{{"$type", "$v"}}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeRecursiveNonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{{"$non-existent", "$v"}}}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeRecursiveInvalid": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{{"v", "$v"}}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeRecursiveArrayInvalid": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{{"$type", bson.A{"1", "2"}}}}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},

		"TypeInt": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", int32(42)}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeLong": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", int64(42)}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeString": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", "42"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeDocument": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{{"foo", "bar"}}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeEmpty": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeArraySingleItem": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.A{int32(42)}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeArray": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.A{"foo", "bar"}}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeNestedArray": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.A{bson.A{"foo", "bar"}}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeObjectID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", primitive.NewObjectID()}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeBool": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", true}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ProjectManyOperators": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"$type", "foo"}, {"$op", "foo"}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"SumValue": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$v"}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatProjectSum(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"Value": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$v"}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotation": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$v.foo"}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayDotNotation": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$v.0.foo"}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Int": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", int32(2)}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Long": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", int64(3)}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Double": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", float64(4)}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyString": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", ""}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayEmptyVariable": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$"}}}},
				}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayValue": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v"}}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayTwoValues": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v", "$v"}}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayValueInt": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v", int32(1)}}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayIntLongDoubleStringBool": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{int32(2), int64(3), float64(4), "not-expression", true}}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"RecursiveValue": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sumsum", bson.D{{"$sum", bson.D{{"$sum", "$v"}}}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"RecursiveArrayValue": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sumsum", bson.D{{"$sum", bson.D{{"$sum", bson.A{"$v"}}}}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayValueRecursiveInt": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v", bson.D{{"$sum", int32(2)}}}}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayValueAndRecursiveValue": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v", bson.D{{"$sum", "$v"}}}}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayValueAndRecursiveArray": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v", bson.D{{"$sum", bson.A{"$v"}}}}}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Type": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sumtype", bson.D{{"$sum", bson.D{{"$type", "$v"}}}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"RecursiveEmptyVariable": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.D{{"$sum", "$$$"}}}}},
				}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MultipleRecursiveEmptyVariable": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.D{{"$sum", bson.D{{"$sum", "$$$"}}}}}}},
				}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatAddFields(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"InvalidTypeString": {
			pipeline: bson.A{
				bson.D{{"$addFields", "invalid"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeBool": {
			pipeline: bson.A{
				bson.D{{"$addFields", false}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeArray": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.A{}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeInt32": {
			pipeline: bson.A{
				bson.D{{"$addFields", int32(1)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeInt64": {
			pipeline: bson.A{
				bson.D{{"$addFields", int64(1)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeFloat32": {
			pipeline: bson.A{
				bson.D{{"$addFields", float32(1)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeFloat64": {
			pipeline: bson.A{
				bson.D{{"$addFields", float64(1)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeNull": {
			pipeline: bson.A{
				bson.D{{"$addFields", nil}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Include1Field": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField", int32(1)}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Include2Fields": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField1", int32(1)}, {"newField2", int32(2)}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Include2Stages": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField1", int32(1)}}}},
				bson.D{{"$addFields", bson.D{{"newField2", int32(2)}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IncludeDocument": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField", bson.D{{"doc", int32(1)}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IncludeNestedDocument": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField", bson.D{{"doc", bson.D{{"nested", int32(1)}}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IncludeArray": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField", bson.A{bson.D{{"elem", int32(1)}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"UnsupportedExpressionObject": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField1", bson.D{{"$sum", 1}}}}}},
			},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/1413",
		},
		"UnsupportedExpressionArray": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField1", bson.A{bson.D{{"$sum", 1}}}}}}},
			},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/1413",
		},

		"InvalidOperator": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"value", bson.D{{"$invalid-operator", "foo"}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Type": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "$v"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeNonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "$foo"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeDotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "$v.foo"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeRecursive": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.D{{"$type", "$v"}}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeRecursiveNonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.D{{"$non-existent", "$v"}}}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeRecursiveInvalid": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.D{{"v", "$v"}}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},

		"TypeInt": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", int32(42)}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeLong": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", int64(42)}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeString": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "42"}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeDocument": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.D{{"foo", "bar"}}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeEmpty": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.D{}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MultipleOperators": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "foo"}, {"$operator", "foo"}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MultipleOperatorFirst": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "foo"}, {"not-operator", "foo"}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"MultipleOperatorLast": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"not-operator", "foo"}, {"$type", "foo"}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeArraySingleItem": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.A{int32(42)}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeArray": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.A{"foo", "bar"}}}}}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeNestedArray": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.A{bson.A{"foo", "bar"}}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeObjectID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", primitive.NewObjectID()}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"TypeBool": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", true}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"SumValue": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{
					{"sum", bson.D{{"$sum", "$v"}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"InvalidTypeString": {
			pipeline: bson.A{
				bson.D{{"$set", "invalid"}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeBool": {
			pipeline: bson.A{
				bson.D{{"$set", false}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeArray": {
			pipeline: bson.A{
				bson.D{{"$set", bson.A{}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeInt32": {
			pipeline: bson.A{
				bson.D{{"$set", int32(1)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeInt64": {
			pipeline: bson.A{
				bson.D{{"$set", int64(1)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeFloat32": {
			pipeline: bson.A{
				bson.D{{"$set", float32(1)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeFloat64": {
			pipeline: bson.A{
				bson.D{{"$set", float64(1)}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeNull": {
			pipeline: bson.A{
				bson.D{{"$set", nil}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Include1Field": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField", int32(1)}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Include2Fields": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField1", int32(1)}, {"newField2", int32(2)}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Include2Stages": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField1", int32(1)}}}},
				bson.D{{"$set", bson.D{{"newField2", int32(2)}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IncludeDocument": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField", bson.D{{"doc", int32(1)}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IncludeNestedDocument": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField", bson.D{{"doc", bson.D{{"nested", int32(1)}}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"IncludeArray": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField", bson.A{bson.D{{"elem", int32(1)}}}}}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"UnsupportedExpressionObject": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField1", bson.D{{"$sum", 1}}}}}},
			},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/1413",
		},
		"UnsupportedExpressionArray": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField1", bson.A{bson.D{{"$sum", 1}}}}}}},
			},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/1413",
		},
		"SumValue": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{
					{"sum", bson.D{{"$sum", "$v"}}},
				}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}
	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatUnset(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"InvalidTypeArray": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{1, 2, 3}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyArray": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"EmptyString": {
			pipeline: bson.A{
				bson.D{{"$unset", ""}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"ArrayWithEmptyString": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{""}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidTypeArrayWithDifferentTypes": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"v", 42, false}}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"InvalidType": {
			pipeline: bson.A{
				bson.D{{"$unset", false}},
			},
			resultType:     emptyResult,
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Unset1Field": {
			pipeline: bson.A{
				bson.D{{"$unset", "v"}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"UnsetID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unset", "_id"}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"Unset2Fields": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"_id", "v"}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationUnset": {
			pipeline: bson.A{
				bson.D{{"$unset", "v.foo"}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationUnsetTwo": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"v.foo", "v.array"}}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
		"DotNotationUnsetSecondLevel": {
			pipeline: bson.A{
				bson.D{{"$unset", "v.array.42"}},
			},
			failsForSQLite: "https://github.com/FerretDB/FerretDB/issues/3148",
		},
	}
	testAggregateStagesCompat(t, testCases)
}
