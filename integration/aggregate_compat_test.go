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
	"golang.org/x/exp/slices"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// aggregateStagesCompatTestCase describes aggregation stages compatibility test case.
type aggregateStagesCompatTestCase struct {
	pipeline       bson.A                   // required, unspecified $sort appends bson.D{{"$sort", bson.D{{"_id", 1}}}} for non empty pipeline.
	resultType     compatTestCaseResultType // defaults to nonEmptyResult
	resultPushdown bool                     // defaults to false

	skip string // skip test for all handlers, must have issue number mentioned
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

			t.Parallel()

			pipeline := tc.pipeline
			require.NotNil(t, pipeline, "pipeline should be set")

			// to make results deterministic
			hasSortStage := slices.ContainsFunc(pipeline, func(el interface{}) bool {
				stage, _ := el.(bson.D)
				return stage.Map()["$sort"] != nil
			})
			if !hasSortStage && len(pipeline) > 0 {
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
		shareddata.BigDoubles,
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
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
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
		"DotNotationID": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.foo"},
			}}}},
			skip: "https://github.com/FerretDB/FerretDB/issues/2166",
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
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatGroupDotNotation(t *testing.T) {
	// Providers Composites, ArrayAndDocuments and Mixed
	// cannot be used due to sorting difference.
	// FerretDB always sorts empty array is less than null.
	// In compat, for `.sort()` an empty array is less than null.
	// In compat, for aggregation `$sort` null is less than an empty array.
	// https://github.com/FerretDB/FerretDB/issues/2276

	providers := []shareddata.Provider{
		shareddata.Scalars,

		shareddata.Doubles,
		shareddata.BigDoubles,
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

		shareddata.ArrayStrings,
		shareddata.ArrayDoubles,
		shareddata.ArrayInt32s,
		shareddata.ArrayRegexes,
		shareddata.ArrayDocuments,

		// shareddata.Mixed,
		// shareddata.ArrayAndDocuments,
	}

	testCases := map[string]aggregateStagesCompatTestCase{
		"DocDotNotation": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.foo"},
			}}}},
		},
		"ArrayDotNotation": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.0"},
			}}}},
		},
		"ArrayDocDotNotation": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.0.foo"},
			}}}},
		},
		"NestedDotNotation": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.0.foo.0.bar"},
			}}}},
		},
		"NonExistentDotNotation": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$non.existent"},
			}}}},
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatGroupDocDotNotation(t *testing.T) {
	// Providers Composites and Mixed cannot be used due to sorting difference.
	// FerretDB always sorts empty array is less than null.
	// In compat, for `.sort()` an empty array is less than null.
	// In compat, for aggregation `$sort` null is less than an empty array.
	// https://github.com/FerretDB/FerretDB/issues/2276

	providers := []shareddata.Provider{
		shareddata.Scalars,

		shareddata.Doubles,
		shareddata.BigDoubles,
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

		shareddata.ArrayStrings,
		shareddata.ArrayDoubles,
		shareddata.ArrayInt32s,
		shareddata.ArrayRegexes,
		shareddata.ArrayDocuments,

		// shareddata.Mixed,
		shareddata.ArrayAndDocuments,
	}

	testCases := map[string]aggregateStagesCompatTestCase{
		"DocDotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", "$v.foo"}}}},
			},
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatGroupCount(t *testing.T) {
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

func TestAggregateCompatMatch(t *testing.T) {
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
		},
		"String": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", "foo"}}}},
			},
			resultPushdown: true,
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
	}

	testAggregateStagesCompat(t, testCases)
}
