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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// aggregateStagesCompatTestCase describes aggregation stages compatibility test case.
type aggregateStagesCompatTestCase struct {
	pipeline       bson.A                   // required, unspecified $sort appends bson.D{{"$sort", bson.D{{"_id", 1}}}} for non empty pipeline.
	resultType     compatTestCaseResultType // defaults to nonEmptyResult
	resultPushdown bool                     // defaults to false

	skip          string // skip test for all handlers, must have issue number mentioned
	skipForTigris string // skip test for Tigris handler, must have issue number mentioned
}

// testAggregateStagesCompat tests aggregation stages compatibility test cases with all providers.
func testAggregateStagesCompat(t *testing.T, testCases map[string]aggregateStagesCompatTestCase) {
	t.Helper()

	testAggregateStagesCompatWithProviders(t, shareddata.AllProviders(), testCases)
}

// testAggregateStagesCompatWithProviders tests aggregation stages compatibility test cases with given providers.
func testAggregateStagesCompatWithProviders(t *testing.T, providers shareddata.Providers, testCases map[string]aggregateStagesCompatTestCase) {
	t.Helper()

	require.NotEmpty(t, providers)

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers: providers,
	})
	ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			if tc.skipForTigris != "" {
				setup.SkipForTigrisWithReason(t, tc.skipForTigris)
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

			if !hasSortStage && len(pipeline) > 0 {
				// add sort stage to sort by _id because compat and target
				// would be ordered differently otherwise.
				pipeline = append(pipeline, bson.D{{"$sort", bson.D{{"_id", 1}}}})
			}

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

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
	command    bson.D                   // required
	resultType compatTestCaseResultType // defaults to nonEmptyResult

	skip string // skip test for all handlers, must have issue number mentioned
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

					// error messages are intentionally not compared
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
		"InvalidStage": {
			command: bson.D{
				{"aggregate", "collection-name"},
				{"pipeline", bson.A{"$invalid-stage"}},
			},
			resultType: emptyResult,
		},
	}

	testAggregateCommandCompat(t, testCases)
}

func TestAggregateCompatStages(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "https://github.com/FerretDB/FerretDB/issues/2523")

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
		},
		"CountAndMatch": {
			pipeline: bson.A{
				// when $match is the second stage, no pushdown is done.
				bson.D{{"$count", "v"}},
				bson.D{{"$match", bson.D{{"v", 1}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
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
			pipeline: bson.A{},
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatCount(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"Value": {
			pipeline: bson.A{bson.D{{"$count", "v"}}},
		},
		"NonExistent": {
			pipeline: bson.A{bson.D{{"$count", "nonexistent"}}},
		},
		"CountGroupID": {
			pipeline:   bson.A{bson.D{{"$count", "_id"}}},
			resultType: emptyResult,
		},
		"CountNonString": {
			pipeline:   bson.A{bson.D{{"$count", 1}}},
			resultType: emptyResult,
		},
		"CountEmpty": {
			pipeline:   bson.A{bson.D{{"$count", ""}}},
			resultType: emptyResult,
		},
		"CountBadValue": {
			pipeline:   bson.A{bson.D{{"$count", "v.foo"}}},
			resultType: emptyResult,
		},
		"CountBadPrefix": {
			pipeline:   bson.A{bson.D{{"$count", "$foo"}}},
			resultType: emptyResult,
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatGroupDeterministicCollections(t *testing.T) {
	t.Parallel()

	// Scalars collection is not included because aggregation groups
	// numbers of different types for $group, and this causes output
	// _id to be different number type between compat and target.
	// https://github.com/FerretDB/FerretDB/issues/2184
	//
	// Composites, ArrayStrings, ArrayInt32s and ArrayAndDocuments are not included
	// because the order in compat and target can be not deterministic.
	// Aggregation assigns BSON array to output _id, and an array with
	// descending sort use the greatest element for comparison causing
	// multiple documents with the same greatest element the same order,
	// so compat and target results in different order.
	// https://github.com/FerretDB/FerretDB/issues/2185

	providers := []shareddata.Provider{
		// shareddata.Scalars,

		shareddata.Doubles,
		shareddata.OverflowVergeDoubles,
		shareddata.SmallDoubles,
		shareddata.Strings,
		shareddata.Binaries,
		shareddata.ObjectIDs,
		shareddata.Bools,
		shareddata.DateTimes,
		shareddata.Nulls,
		shareddata.Regexes,
		shareddata.Int32s,
		shareddata.Timestamps,
		shareddata.Int64s,
		shareddata.Unsets,
		shareddata.ObjectIDKeys,

		// shareddata.Composites,
		shareddata.PostgresEdgeCases,

		shareddata.DocumentsDoubles,
		shareddata.DocumentsStrings,
		shareddata.DocumentsDocuments,

		// shareddata.ArrayStrings,
		shareddata.ArrayDoubles,
		// shareddata.ArrayInt32s,
		shareddata.ArrayRegexes,
		shareddata.ArrayDocuments,

		// shareddata.Mixed,
		// shareddata.ArrayAndDocuments,
	}

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
		},
		"DistinctID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$_id"},
			}}}},
		},
		"IDExpression": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", bson.D{{"v", "$v"}}},
			}}}},
			skip: "https://github.com/FerretDB/FerretDB/issues/2165",
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
		"EmptyPath": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$"},
			}}}},
			resultType: emptyResult,
		},
		"EmptyVariable": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$"},
			}}}},
			resultType: emptyResult,
		},
		"InvalidVariable$": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$$"},
			}}}},
			resultType: emptyResult,
		},
		"InvalidVariable$s": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$$s"},
			}}}},
			resultType: emptyResult,
		},
		"NonExistingVariable": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$s"},
			}}}},
			resultType: emptyResult,
		},
		"SystemVariable": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$NOW"},
			}}}},
			resultType: emptyResult,
			skip:       "https://github.com/FerretDB/FerretDB/issues/2275",
		},
		"GroupInvalidFields": {
			pipeline:   bson.A{bson.D{{"$group", 1}}},
			resultType: emptyResult,
		},
		"EmptyGroup": {
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
		"GroupMultipleAccumulator": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"v", bson.D{{"$count", "v"}, {"$count", "v"}}},
			}}}},
			resultType: emptyResult,
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
				{"_id", bson.D{{"$type", "_id"}}},
			}}}},
			skip: "https://github.com/FerretDB/FerretDB/issues/2679",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatGroupExpressionDottedFields(t *testing.T) {
	t.Parallel()

	// TODO Use all providers after fixing $sort problem:  https://github.com/FerretDB/FerretDB/issues/2276.
	//
	// Currently, providers Composites, DocumentsDeeplyNested and Mixed
	// cannot be used due to sorting difference.
	// FerretDB always sorts empty array is less than null.
	// In compat, for `.sort()` an empty array is less than null.
	// In compat, for aggregation `$sort` null is less than an empty array.
	providers := shareddata.AllProviders().Remove("Mixed", "Composites", "DocumentsDeeplyNested")

	testCases := map[string]aggregateStagesCompatTestCase{
		"NestedInDocument": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.foo"},
			}}}},
		},
		"DeeplyNested": { // Expect non-empty results for DocumentsDeeplyNested provider
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.a.b.c"},
			}}}},
		},
		"ArrayIndex": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.0"},
			}}}},
		},
		"NestedInArray": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.0.foo"},
			}}}},
		},
		"NonExistentParent": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$non.existent"},
			}}}},
		},
		"NonExistentChild": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.non.existent"},
			}}}},
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
		},
		"CountID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$_id"},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
		},
		"TypeMismatch": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"count", bson.D{{"$count", ""}}},
			}}}},
			resultType: emptyResult,
		},
		"NonEmptyExpression": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"count", bson.D{{"$count", bson.D{{"a", 1}}}}},
			}}}},
			resultType: emptyResult,
		},
		"NonExistentField": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$nonexistent"},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
		},
		"Duplicate": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v"},
				{"count", bson.D{{"$count", bson.D{}}}},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
			resultType: emptyResult,
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
			resultType: emptyResult,
		},
		"One": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 1}},
			},
		},
		"Five": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 5}},
			},
		},
		"StringInt": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", "5"}},
			},
			resultType: emptyResult,
		},
		"Double": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 4.5}},
			},
			resultType: emptyResult,
		},
		"DoubleInt": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 5.0}},
			},
		},
		"MaxInt64": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", math.MaxInt64}},
			},
		},
		"MinInt64": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", math.MinInt64}},
			},
			resultType: emptyResult,
		},
		"Negative": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", -1}},
			},
			resultType: emptyResult,
		},
		"NegativeDouble": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", -2.1}},
			},
			resultType: emptyResult,
		},
		"Document": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", bson.D{}}},
			},
			resultType: emptyResult,
		},
		"Int64Overflow": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", float64(1 << 86)}},
			},
			resultType: emptyResult,
		},
		"AfterMatch": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
				bson.D{{"$limit", 1}},
			},
			resultPushdown: true, // $sort and $match are first two stages
		},
		"BeforeMatch": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$limit", 1}},
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
			},
			resultType: emptyResult,
		},
		"NoSortAfterMatch": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
				bson.D{{"$limit", 100}},
			},
			resultPushdown: true,
			skipForTigris:  "TestAggregateCompatLimit",
		},
		"NoSortBeforeMatch": {
			pipeline: bson.A{
				bson.D{{"$limit", 100}},
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
			},
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatGroupSum(t *testing.T) {
	t.Parallel()

	providers := shareddata.AllProviders().
		// skipped due to https://github.com/FerretDB/FerretDB/issues/2185.
		Remove("Composites").
		Remove("ArrayStrings").
		Remove("ArrayInt32s").
		Remove("Mixed").
		Remove("ArrayAndDocuments").
		// TODO: handle $sum of doubles near max precision.
		// https://github.com/FerretDB/FerretDB/issues/2300
		Remove("Doubles").
		// TODO: https://github.com/FerretDB/FerretDB/issues/2616
		Remove("ArrayDocuments")

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
			resultType: emptyResult,
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
			resultType: emptyResult,
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
			skip: "https://github.com/FerretDB/FerretDB/issues/2694",
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatMatch(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"ID": {
			pipeline:       bson.A{bson.D{{"$match", bson.D{{"_id", "string"}}}}},
			resultPushdown: true,
		},
		"Int": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", 42}}}},
			},
			resultPushdown: true,
			skipForTigris:  "https://github.com/FerretDB/FerretDB/issues/2523",
		},
		"String": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
			},
			resultPushdown: true,
			skipForTigris:  "https://github.com/FerretDB/FerretDB/issues/2523",
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
		"MatchBadValue": {
			pipeline:   bson.A{bson.D{{"$match", 1}}},
			resultType: emptyResult,
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatSort(t *testing.T) {
	t.Parallel()

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
		"AscendingValueDescendingID": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v", 1},
				{"_id", -1},
			}}}},
		},
		"DescendingValueDescendingID": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v", -1},
				{"_id", -1},
			}}}},
		},

		"DotNotationIndex": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v.0", 1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
		},
		"DotNotationNonExistent": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"invalid.foo", 1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
		},
		"DotNotationMissingField": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v..foo", 1},
			}}}},
			resultType: emptyResult,
		},

		"SortBadExpression": {
			pipeline:   bson.A{bson.D{{"$sort", 1}}},
			resultType: emptyResult,
		},
		"SortBadOrder": {
			pipeline:   bson.A{bson.D{{"$sort", bson.D{{"_id", 0}}}}},
			resultType: emptyResult,
		},
		"SortMissingKey": {
			pipeline:   bson.A{bson.D{{"$sort", bson.D{}}}},
			resultType: emptyResult,
		},
		"BadDollarStart": {
			pipeline:   bson.A{bson.D{{"$sort", bson.D{{"$v.foo", 1}}}}},
			resultType: emptyResult,
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatSortDotNotation(t *testing.T) {
	t.Parallel()

	providers := shareddata.AllProviders().
		// TODO: https://github.com/FerretDB/FerretDB/issues/2617
		Remove("ArrayDocuments")

	testCases := map[string]aggregateStagesCompatTestCase{
		"DotNotation": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v.foo", 1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
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
		},
		"NonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$non-existent"}},
			},
			resultType: emptyResult,
		},
		"Invalid": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "invalid"}},
			},
			resultType: emptyResult,
		},
		"DotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$v.foo"}},
			},
		},
		"DotNotationNonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$v.non-existent"}},
			},
			resultType: emptyResult,
		},
		"ArrayDotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$v.0"}},
			},
			resultType: emptyResult,
		},
		"ArrayDotNotationKey": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$v.0.foo"}},
			},
			resultType: emptyResult,
		},
		"Null": {
			pipeline:   bson.A{bson.D{{"$unwind", nil}}},
			resultType: emptyResult,
		},
		"ID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$_id"}},
			},
		},
		"NameAsExpression": {
			pipeline:   bson.A{bson.D{{"$unwind", "$add"}}},
			resultType: emptyResult,
		},
		"EmptyPath": {
			pipeline:   bson.A{bson.D{{"$unwind", "$"}}},
			resultType: emptyResult,
		},
		"EmptyVariable": {
			pipeline:   bson.A{bson.D{{"$unwind", "$$"}}},
			resultType: emptyResult,
		},
		"InvalidVariable$": {
			pipeline:   bson.A{bson.D{{"$unwind", "$$$"}}},
			resultType: emptyResult,
		},
		"InvalidVariable$s": {
			pipeline:   bson.A{bson.D{{"$unwind", "$$$s"}}},
			resultType: emptyResult,
		},
		"NonExistingVariable": {
			pipeline:   bson.A{bson.D{{"$unwind", "$$s"}}},
			resultType: emptyResult,
		},
		"SystemVariable": {
			pipeline:   bson.A{bson.D{{"$unwind", "$$NOW"}}},
			resultType: emptyResult,
		},
		"Empty": {
			pipeline:   bson.A{bson.D{{"$unwind", ""}}},
			resultType: emptyResult,
		},
		"Number": {
			pipeline:   bson.A{bson.D{{"$unwind", 42}}},
			resultType: emptyResult,
		},
		"Array": {
			pipeline:   bson.A{bson.D{{"$unwind", bson.A{"$v"}}}},
			resultType: emptyResult,
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatSkip(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"Document": {
			pipeline:   bson.A{bson.D{{"$skip", bson.D{}}}},
			resultType: emptyResult,
		},
		"Zero": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", int32(0)}},
			},
		},
		"One": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", int32(1)}},
			},
		},
		"SkipAll": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", int32(1000)}},
			},
			resultType: emptyResult,
		},
		"StringInt": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", "1"}},
			},
			resultType: emptyResult,
		},
		"NegativeValue": {
			pipeline:   bson.A{bson.D{{"$skip", int32(-1)}}},
			resultType: emptyResult,
		},
		"NegativeDouble": {
			pipeline:   bson.A{bson.D{{"$skip", -3.2}}},
			resultType: emptyResult,
		},
		"MaxInt64": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", math.MaxInt64}},
			},
			resultType: emptyResult,
		},
		"MinInt64": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", math.MinInt64}},
			},
			resultType: emptyResult,
		},
		"Int64Overflow": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", float64(1 << 86)}},
			},
			resultType: emptyResult,
		},
		"AfterMatch": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
				bson.D{{"$skip", int32(1)}},
			},
			resultPushdown: true, // $match after $sort can be pushed down
		},
		"BeforeMatch": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$skip", int32(1)}},
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
			},
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
			resultType: emptyResult,
		},
		"ZeroOperators": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"v", bson.D{}}}}},
			},
			resultType: emptyResult,
		},
		"TwoOperators": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"v", bson.D{{"$type", "foo"}, {"$sum", 1}}}}}},
			},
			resultType: emptyResult,
		},
		"DollarSignField": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"$v", 1}}}},
			},
			resultType: emptyResult,
		},
		"Include1Field": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"v", int32(1)}}}},
			},
		},
		"Exclude1Field": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"v", int64(0)}}}},
			},
		},
		"IncludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", 1.42}}}},
			},
		},
		"ExcludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", false}}}},
			},
		},
		"Include2Fields": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"foo", 1.24}, {"bar", true}}}},
			},
		},
		"Exclude2Fields": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},

				bson.D{{"$project", bson.D{{"foo", int32(0)}, {"bar", false}}}},
			},
		},
		"Include1FieldExclude1Field": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},

				bson.D{{"$project", bson.D{{"foo", int32(0)}, {"bar", true}}}},
			},
			resultType: emptyResult,
		},
		"Exclude1FieldInclude1Field": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},

				bson.D{{"$project", bson.D{{"foo", int32(1)}, {"bar", false}}}},
			},
			resultType: emptyResult,
		},
		"IncludeFieldExcludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", false}, {"v", true}}}},
			},
		},
		"ExcludeFieldIncludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", true}, {"v", false}}}},
			},
		},
		"ExcludeFieldExcludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", false}, {"v", false}}}},
			},
		},
		"IncludeFieldIncludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", true}, {"v", true}}}},
			},
		},
		"Assign1Field": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"foo", primitive.NewObjectID()}}}},
			},
		},
		"AssignID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", primitive.Binary{Subtype: 0x80, Data: []byte{42, 0, 13}}}}}},
			},
		},
		"Assign1FieldIncludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", true}, {"foo", primitive.NewDateTimeFromTime(time.Unix(0, 0))}}}},
			},
		},
		"Assign2FieldsIncludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", true}, {"foo", nil}, {"bar", "qux"}}}},
			},
		},
		"Assign1FieldExcludeID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", false}, {"foo", primitive.Regex{Pattern: "^fo"}}}}},
			},
		},
		"DotNotationInclude": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.foo", true},
				}}},
			},
		},
		"DotNotationIncludeTwo": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.foo", true},
					{"v.array", true},
				}}},
			},
		},
		"DotNotationExclude": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.foo", false},
				}}},
			},
		},
		"DotNotationExcludeTwo": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.foo", false},
					{"v.array", false},
				}}},
			},
		},
		"DotNotationExcludeSecondLevel": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.array.42", false},
				}}},
			},
		},
		"DotNotationIncludeExclude": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"_id", true},
					{"v.foo", true},
					{"v.array.42", false},
				}}},
			},
			resultType: emptyResult,
		},
		"Type": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", "$v"}}}}}},
			},
		},
		"TypeNonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", "$foo"}}}}}},
			},
		},
		"TypeDotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", "$v.foo"}}}}}},
			},
		},
		"TypeRecursive": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{{"$type", "$v"}}}}}}}},
			},
		},
		"TypeRecursiveNonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{{"$non-existent", "$v"}}}}}}}},
			},
			resultType: emptyResult,
		},
		"TypeRecursiveInvalid": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{{"v", "$v"}}}}}}}},
			},
		},
		"TypeRecursiveArrayInvalid": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{{"$type", bson.A{"1", "2"}}}}}}}}},
			},
			resultType: emptyResult,
		},

		"TypeInt": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", int32(42)}}}}}},
			},
		},
		"TypeLong": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", int64(42)}}}}}},
			},
		},
		"TypeString": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", "42"}}}}}},
			},
		},
		"TypeDocument": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{{"foo", "bar"}}}}}}}},
			},
		},
		"TypeArraySingleItem": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.A{int32(42)}}}}}}},
			},
		},
		"TypeArray": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.A{"foo", "bar"}}}}}}},
			},
			resultType: emptyResult,
		},
		"TypeNestedArray": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.A{bson.A{"foo", "bar"}}}}}}}},
			},
		},
		"TypeObjectID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", primitive.NewObjectID()}}}}}},
			},
		},
		"TypeBool": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", true}}}}}},
			},
		},
		"ProjectManyOperators": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"$type", "foo"}, {"$op", "foo"}}}},
			},
			resultType: emptyResult,
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
			resultType: emptyResult,
		},
		"InvalidTypeBool": {
			pipeline: bson.A{
				bson.D{{"$addFields", false}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeArray": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.A{}}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeInt32": {
			pipeline: bson.A{
				bson.D{{"$addFields", int32(1)}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeInt64": {
			pipeline: bson.A{
				bson.D{{"$addFields", int64(1)}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeFloat32": {
			pipeline: bson.A{
				bson.D{{"$addFields", float32(1)}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeFloat64": {
			pipeline: bson.A{
				bson.D{{"$addFields", float64(1)}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeNull": {
			pipeline: bson.A{
				bson.D{{"$addFields", nil}},
			},
			resultType: emptyResult,
		},
		"Include1Field": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField", int32(1)}}}},
			},
		},
		"Include2Fields": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField1", int32(1)}, {"newField2", int32(2)}}}},
			},
		},
		"Include2Stages": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField1", int32(1)}}}},
				bson.D{{"$addFields", bson.D{{"newField2", int32(2)}}}},
			},
		},
		"IncludeDocument": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField", bson.D{{"doc", int32(1)}}}}}},
			},
		},
		"IncludeNestedDocument": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField", bson.D{{"doc", bson.D{{"nested", int32(1)}}}}}}}},
			},
		},
		"IncludeArray": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{{"newField", bson.A{bson.D{{"elem", int32(1)}}}}}}},
			},
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

		"Type": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "$v"}}}}}},
			},
		},
		"TypeNonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "$foo"}}}}}},
			},
		},
		"TypeDotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "$v.foo"}}}}}},
			},
		},
		"TypeRecursive": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.D{{"$type", "$v"}}}}}}}},
			},
		},
		"TypeRecursiveNonExistent": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.D{{"$non-existent", "$v"}}}}}}}},
			},
		},
		"TypeRecursiveInvalid": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.D{{"v", "$v"}}}}}}}},
			},
		},

		"TypeInt": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", int32(42)}}}}}},
			},
		},
		"TypeLong": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", int64(42)}}}}}},
			},
		},
		"TypeString": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "42"}}}}}},
			},
		},
		"TypeDocument": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.D{{"foo", "bar"}}}}}}}},
			},
		},
		"TypeArraySingleItem": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.A{int32(42)}}}}}}},
			},
		},
		"TypeArray": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.A{"foo", "bar"}}}}}}},
			},
		},
		"TypeNestedArray": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.A{bson.A{"foo", "bar"}}}}}}}},
			},
		},
		"TypeObjectID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", primitive.NewObjectID()}}}}}},
			},
		},
		"TypeBool": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", true}}}}}},
			},
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
			resultType: emptyResult,
		},
		"InvalidTypeBool": {
			pipeline: bson.A{
				bson.D{{"$set", false}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeArray": {
			pipeline: bson.A{
				bson.D{{"$set", bson.A{}}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeInt32": {
			pipeline: bson.A{
				bson.D{{"$set", int32(1)}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeInt64": {
			pipeline: bson.A{
				bson.D{{"$set", int64(1)}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeFloat32": {
			pipeline: bson.A{
				bson.D{{"$set", float32(1)}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeFloat64": {
			pipeline: bson.A{
				bson.D{{"$set", float64(1)}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeNull": {
			pipeline: bson.A{
				bson.D{{"$set", nil}},
			},
			resultType: emptyResult,
		},
		"Include1Field": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField", int32(1)}}}},
			},
		},
		"Include2Fields": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField1", int32(1)}, {"newField2", int32(2)}}}},
			},
		},
		"Include2Stages": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField1", int32(1)}}}},
				bson.D{{"$set", bson.D{{"newField2", int32(2)}}}},
			},
		},
		"IncludeDocument": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField", bson.D{{"doc", int32(1)}}}}}},
			},
		},
		"IncludeNestedDocument": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField", bson.D{{"doc", bson.D{{"nested", int32(1)}}}}}}}},
			},
		},
		"IncludeArray": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{{"newField", bson.A{bson.D{{"elem", int32(1)}}}}}}},
			},
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
			resultType: emptyResult,
		},
		"EmptyArray": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{}}},
			},
			resultType: emptyResult,
		},
		"EmptyString": {
			pipeline: bson.A{
				bson.D{{"$unset", ""}},
			},
			resultType: emptyResult,
		},
		"ArrayWithEmptyString": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{""}}},
			},
			resultType: emptyResult,
		},
		"InvalidTypeArrayWithDifferentTypes": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"v", 42, false}}},
			},
			resultType: emptyResult,
		},
		"InvalidType": {
			pipeline: bson.A{
				bson.D{{"$unset", false}},
			},
			resultType: emptyResult,
		},
		"Unset1Field": {
			pipeline: bson.A{
				bson.D{{"$unset", "v"}},
			},
		},
		"UnsetID": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unset", "_id"}},
			},
		},
		"Unset2Fields": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"_id", "v"}}},
			},
		},
		"DotNotationUnset": {
			pipeline: bson.A{
				bson.D{{"$unset", "v.foo"}},
			},
		},
		"DotNotationUnsetTwo": {
			pipeline: bson.A{
				bson.D{{"$unset", bson.A{"v.foo", "v.array"}}},
			},
		},
		"DotNotationUnsetSecondLevel": {
			pipeline: bson.A{
				bson.D{{"$unset", "v.array.42"}},
			},
		},
	}
	testAggregateStagesCompat(t, testCases)
}
