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
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
)

// aggregateStagesCompatTestCase describes aggregation stages compatibility test case.
type aggregateStagesCompatTestCase struct {
	pipeline bson.A         // required, unspecified $sort appends bson.D{{"$sort", bson.D{{"_id", 1}}}} for non empty pipeline.
	maxTime  *time.Duration // optional, leave nil for unset maxTime

	resultType       compatTestCaseResultType // defaults to nonEmptyResult
	skip             string                   // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1086
	failsForFerretDB string
	failsProviders   []shareddata.Provider // use only if failsForFerretDB is set, defaults to all providers
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

			if !hasSortStage && len(pipeline) > 0 {
				// add sort stage to sort by _id because compat and target
				// would be ordered differently otherwise.
				pipeline = append(pipeline, bson.D{{"$sort", bson.D{{"_id", 1}}}})
			}

			opts := options.Aggregate()

			if tc.maxTime != nil {
				opts.SetMaxTime(*tc.maxTime)
			}

			failsProviders := make([]string, len(tc.failsProviders))
			for i, p := range tc.failsProviders {
				failsProviders[i] = p.Name()
			}

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]

				t.Run(targetCollection.Name(), func(tt *testing.T) {
					tt.Helper()

					// a workaround to get provider name by using the part after last `_`,
					// e.g. `Doubles` from `TestQueryArrayCompatElemMatch_Doubles`
					str := strings.Split(targetCollection.Name(), "_")
					providerName := str[len(str)-1]

					failsForCollection := len(tc.failsProviders) == 0 || slices.Contains(failsProviders, providerName)

					var t testing.TB = tt

					if tc.failsForFerretDB != "" && failsForCollection {
						t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
					}

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

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results")
			case emptyResult:
				if tc.failsForFerretDB != "" {
					return
				}

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

	skip             string // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/1086
	failsForFerretDB string
}

// testAggregateCommandCompat tests aggregate pipeline compatibility test cases using one collection.
// Use testAggregateStagesCompat for testing stages of aggregation.
func testAggregateCommandCompat(t *testing.T, testCases map[string]aggregateCommandCompatTestCase) {
	t.Helper()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers: []shareddata.Provider{shareddata.Int32s},
	})

	ctx := s.Ctx
	targetCollection := s.TargetCollections[0]
	compatCollection := s.CompatCollections[0]

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			command := tc.command
			require.NotNil(t, command, "command should be set")

			var nonEmptyResults bool

			t.Run(targetCollection.Name(), func(tt *testing.T) {
				tt.Helper()

				var t testing.TB = tt

				if tc.failsForFerretDB != "" {
					t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
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
		"MaxTimeMSDoubleWholeNumber": {
			command: bson.D{
				{"aggregate", "collection-name"},
				{"pipeline", bson.A{}},
				{"cursor", bson.D{}},
				{"maxTimeMS", float64(1000)},
			},
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
			pipeline: bson.A{},
			maxTime:  pointer.ToDuration(time.Duration(0)),
		},
		"MaxTime": {
			pipeline: bson.A{},
			maxTime:  pointer.ToDuration(time.Second),
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatStages(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"MatchAndCount": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"v", 42}}}},
				bson.D{{"$count", "v"}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/396",
			failsProviders: []shareddata.Provider{
				shareddata.OverflowVergeDoubles,
				shareddata.Strings,
				shareddata.Binaries,
				shareddata.ObjectIDs,
				shareddata.Bools,
				shareddata.DateTimes,
				shareddata.Nulls,
				shareddata.Regexes,
				shareddata.Timestamps,
				shareddata.Unsets,
				shareddata.ObjectIDKeys,
				shareddata.PostgresEdgeCases,
				shareddata.DocumentsDoubles,
				shareddata.DocumentsStrings,
				shareddata.DocumentsDocuments,
				shareddata.DocumentsDeeplyNested,
				shareddata.ArrayStrings,
				shareddata.ArrayDoubles,
				shareddata.ArrayRegexes,
				shareddata.ArrayDocuments,
				shareddata.Mixed,
				shareddata.ArrayAndDocuments,
			},
		},
		"CountAndMatch": {
			pipeline: bson.A{
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
			pipeline:         bson.A{bson.D{{"$count", "_id"}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
		"CountNonString": {
			pipeline:   bson.A{bson.D{{"$count", 1}}},
			resultType: emptyResult,
		},
		"CountEmpty": {
			pipeline:         bson.A{bson.D{{"$count", ""}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
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

func TestAggregateCompatGroupNonArrayProviders(t *testing.T) {
	t.Parallel()

	// Composites, ArrayStrings, ArrayInt32s, ArrayAndDocuments and Mixed are not included
	// because the order in compat and target can be not deterministic.
	// Aggregation assigns BSON array to output _id, and an array with
	// descending sort use the greatest element for comparison causing
	// multiple documents with the same greatest element the same order,
	// so compat and target results in different order.
	// https://github.com/FerretDB/FerretDB/issues/2185
	//
	// The Decimal128s are not included because they make tests flaky.
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/383

	providers := shareddata.AllProviders().Remove(
		shareddata.Composites, shareddata.Mixed,
		shareddata.ArrayStrings, shareddata.ArrayInt32s, shareddata.ArrayAndDocuments,
		shareddata.Decimal128s,
	)
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
				shareddata.ArrayDocuments, // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/385
			},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
				shareddata.ArrayDocuments, // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/385
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/385",
			failsProviders: []shareddata.Provider{
				shareddata.ArrayDocuments,
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/385",
			failsProviders: []shareddata.Provider{
				shareddata.ArrayDocuments,
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
				shareddata.ArrayDocuments, // TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/385
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
			},
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatGroup(t *testing.T) {
	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/383
	t.Skip("https://github.com/FerretDB/FerretDB-DocumentDB/issues/383")

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
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"v", "$v"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
			},
		},
		"IDExpressionNested": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"nested", bson.D{{"v", "$v"}}}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/388",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
				shareddata.Unsets,
				shareddata.Mixed,
			},
		},
		"IDExpressionDotNotation": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"v", "$v.foo"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/388",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
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
				shareddata.ObjectIDKeys,
				shareddata.DocumentsDoubles,
				shareddata.DocumentsStrings,
				shareddata.Mixed,
			},
		},
		"IDExpressionNonExistentField": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"missing", "$non-existent"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
		},
		"IDExpressionFields": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{
						{"v", "$v"},
						{"foo", "$v.foo"},
					}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/388",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
				shareddata.Unsets,
				shareddata.Composites,
				shareddata.DocumentsDoubles,
				shareddata.DocumentsStrings,
				shareddata.DocumentsDocuments,
				shareddata.DocumentsDeeplyNested,
				shareddata.PostgresEdgeCases,
				shareddata.Mixed,
			},
		},
		"IDExpressionNonExistentFields": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{
						{"missing1", "$non-existent1"},
						{"missing2", "$non-existent2"},
					}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/388",
		},
		"IDExpressionAndOperator": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{
						{"v", "$v"},
						{"sum", bson.D{{"$sum", "$v"}}},
					}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/388",
			failsProviders: []shareddata.Provider{
				shareddata.Unsets,
				shareddata.Mixed,
				shareddata.Scalars,
			},
		},
		"IDInvalidExpressionAndOperator": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{
						{"v", "$"},
						{"sum", bson.D{{"$sum", "$v"}}},
					}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"IDExpressionAndInvalidOperator": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{
						{"v", "$v"},
						{"sum", bson.D{{"$sum", "$v"}, {"second", "field"}}},
					}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/389",
		},
		"IDDocument": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"v", "v"}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
		},
		"IDNestedDocument": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{{"v", bson.D{{"nested", 1}}}}},
				}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"EmptyVariable": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$$"},
			}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
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
			skip:       "https://github.com/FerretDB/FerretDB/issues/2275",
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/389",
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/389",
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
				{"_id", bson.D{{"$type", "$v"}}},
				{"count", bson.D{{"$count", bson.D{}}}},
			}}}},
		},
		"IDSum": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$sum", "$v"}}}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
				shareddata.Doubles,
			},
		},
		"IDFieldSum": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", bson.D{{"sum", bson.D{{"$sum", "$v"}}}}}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
				shareddata.Doubles,
			},
		},
		"IDNestedFieldSum": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", bson.D{{"nested", bson.D{{"sum", bson.D{{"$sum", "$v"}}}}}}}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders: []shareddata.Provider{
				shareddata.Scalars,
				shareddata.Doubles,
			},
		},
		"IDSumNonExistentField": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$sum", "$non-existent"}}}}}},
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
			},
		},
		"IDSumInvalid": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$sum", "$"}}}}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"IDSumRecursiveInvalid": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"_id", bson.D{{"$sum", bson.D{{"$sum", "$"}}}}}}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatGroupExpressionDotNotation(t *testing.T) {
	t.Parallel()

	// Use all providers after fixing $sort problem:
	// TODO https://github.com/FerretDB/FerretDB/issues/2276
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/394",
			failsProviders: shareddata.Providers{
				shareddata.Scalars,
				shareddata.Doubles,
				shareddata.Decimal128s,
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
				shareddata.ObjectIDKeys,
				shareddata.DocumentsDoubles,
				shareddata.DocumentsStrings,
			},
		},
		"DeeplyNested": { // Expect non-empty results for DocumentsDeeplyNested provider
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.a.b.c"},
			}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/394",
			failsProviders: shareddata.Providers{
				shareddata.Scalars,
				shareddata.Doubles,
				shareddata.Decimal128s,
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
				shareddata.ObjectIDKeys,
				shareddata.DocumentsDoubles,
				shareddata.DocumentsStrings,
			},
		},
		"ArrayIndex": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.0"},
			}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/394",
			failsProviders: shareddata.Providers{
				shareddata.Scalars,
				shareddata.Doubles,
				shareddata.Decimal128s,
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
				shareddata.ObjectIDKeys,
				shareddata.DocumentsDoubles,
				shareddata.DocumentsStrings,
			},
		},
		"NestedInArray": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.0.foo"},
			}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/394",
			failsProviders: shareddata.Providers{
				shareddata.Scalars,
				shareddata.Doubles,
				shareddata.Decimal128s,
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
				shareddata.ObjectIDKeys,
				shareddata.DocumentsDoubles,
				shareddata.DocumentsStrings,
			},
		},
		"NonExistentChild": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", "$v.non.existent"},
			}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/394",
			failsProviders: shareddata.Providers{
				shareddata.Scalars,
				shareddata.Doubles,
				shareddata.Decimal128s,
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
				shareddata.ObjectIDKeys,
				shareddata.DocumentsDoubles,
				shareddata.DocumentsStrings,
			},
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
}

func TestAggregateCompatGroupExpressionNestedDotNotation(t *testing.T) {
	t.Parallel()

	// Merge the current function with TestAggregateCompatGroupExpressionDottedFields
	// and use all providers when $sort problem is fixed:
	// TODO https://github.com/FerretDB/FerretDB/issues/2276

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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
		},
		"NonEmptyExpression": {
			pipeline: bson.A{bson.D{{"$group", bson.D{
				{"_id", nil},
				{"count", bson.D{{"$count", bson.D{{"a", 1}}}}},
			}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
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
		Remove(shareddata.Composites).
		Remove(shareddata.ArrayStrings).
		Remove(shareddata.ArrayInt32s).
		Remove(shareddata.Mixed).
		Remove(shareddata.ArrayAndDocuments).
		// Handle $sum of doubles near max precision.
		// TODO https://github.com/FerretDB/FerretDB/issues/2300
		Remove(shareddata.Doubles).
		// TODO https://github.com/FerretDB/FerretDB/issues/2616
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars, shareddata.Int64s, shareddata.Int32s},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/383",
			failsProviders:   []shareddata.Provider{shareddata.Scalars},
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/390",
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
		},
		"RecursiveInvalid": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"sum", bson.D{{"$sum", bson.D{{"v", "$v"}}}}}}}},
			},
			resultType: emptyResult,
		},
		"RecursiveArrayInvalid": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{{"sum", bson.D{{"$sum", bson.D{{"$type", bson.A{"1", "2"}}}}}}}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/389",
		},
		"RecursiveOperatorNonExistent": {
			pipeline: bson.A{
				bson.D{{"$group", bson.D{
					{"_id", "$_id"},
					// first $sum is accumulator operator, second $sum is operator
					{"sum", bson.D{{"$sum", bson.D{{"$non-existent", "$v"}}}}},
				}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/389",
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
		"MatchBadValue": {
			pipeline:   bson.A{bson.D{{"$match", 1}}},
			resultType: emptyResult,
		},
		"SumValue": {
			pipeline: bson.A{
				bson.D{{"$match", bson.D{{"$expr", bson.D{{"$sum", "$v"}}}}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/362",
			failsProviders:   []shareddata.Provider{shareddata.Decimal128s, shareddata.Doubles, shareddata.Int64s, shareddata.Scalars},
		},
	}

	testAggregateStagesCompatWithProviders(t, providers, testCases)
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/355",
			failsProviders:   []shareddata.Provider{shareddata.ArrayStrings, shareddata.Composites},
		},
		"DescendingValue": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v", -1},
				{"_id", 1}, // sort by _id when v is the same.
			}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/355",
			failsProviders:   []shareddata.Provider{shareddata.ArrayStrings, shareddata.Composites, shareddata.Mixed},
		},
		"AscendingValueDescendingID": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{
				{"v", 1},
				{"_id", -1},
			}}}},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/355",
			failsProviders:   []shareddata.Provider{shareddata.ArrayStrings, shareddata.Composites, shareddata.Mixed},
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
		},

		"SortBadExpression": {
			pipeline:         bson.A{bson.D{{"$sort", 1}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
		},
		"SortBadOrder": {
			pipeline:         bson.A{bson.D{{"$sort", bson.D{{"_id", 0}}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
		},
		"SortMissingKey": {
			pipeline:   bson.A{bson.D{{"$sort", bson.D{}}}},
			resultType: emptyResult,
		},
		"BadDollarStart": {
			pipeline:         bson.A{bson.D{{"$sort", bson.D{{"$v.foo", 1}}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/354",
		},
	}

	testAggregateStagesCompat(t, testCases)
}

func TestAggregateCompatSortDotNotation(t *testing.T) {
	t.Parallel()

	providers := shareddata.AllProviders().
		// TODO https://github.com/FerretDB/FerretDB/issues/2617
		Remove(shareddata.ArrayDocuments)

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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/367",
			failsProviders: shareddata.Providers{
				shareddata.ArrayStrings,
				shareddata.ArrayAndDocuments,
				shareddata.ArrayDoubles,
				shareddata.ArrayDocuments,
				shareddata.ArrayStrings,
				shareddata.ArrayInt32s,
				shareddata.ArrayInt64s,
				shareddata.ArrayRegexes,
				shareddata.Composites,
			},
		},
		"ArrayDotNotationKey": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", 1}}}},
				bson.D{{"$unwind", "$v.0.foo"}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/367",
			failsProviders: shareddata.Providers{
				shareddata.ArrayAndDocuments,
				shareddata.ArrayDocuments,
			},
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
			pipeline:         bson.A{bson.D{{"$unwind", "$"}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
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
			pipeline:         bson.A{bson.D{{"$skip", bson.D{}}}},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/368",
		},
		"ZeroOperators": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"v", bson.D{}}}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"TwoOperators": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"v", bson.D{{"$type", "foo"}, {"$sum", 1}}}}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/368",
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/368",
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
		"EmptyDocument": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"foo", bson.D{}}}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"Document": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"foo", bson.D{{"v", "foo"}}}}}},
			},
		},
		"IDDocument": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", bson.D{{"v", "foo"}}}}}},
			},
		},
		"IDType": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"_id", bson.D{{"$type", "$v"}}}}}},
			},
		},
		"DocumentAndValue": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{
					{"foo", bson.D{{"v", "foo"}}},
					{"v", 1},
				}}},
			},
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/368",
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
		"TypeEmpty": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$project", bson.D{{"type", bson.D{{"$type", bson.D{}}}}}}},
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

func TestAggregateCompatProjectSum(t *testing.T) {
	t.Parallel()

	testCases := map[string]aggregateStagesCompatTestCase{
		"Value": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$v"}}},
				}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1078",
			failsProviders:   shareddata.Providers{shareddata.Decimal128s},
		},
		"DotNotation": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$v.foo"}}},
				}}},
			},
		},
		"ArrayDotNotation": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", "$v.0.foo"}}},
				}}},
			},
		},
		"Int": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", int32(2)}}},
				}}},
			},
		},
		"Long": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", int64(3)}}},
				}}},
			},
		},
		"Double": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", float64(4)}}},
				}}},
			},
		},
		"EmptyString": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", ""}}},
				}}},
			},
		},
		"ArrayEmptyVariable": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$"}}}},
				}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/369",
		},
		"ArrayValue": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v"}}}},
				}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1078",
			failsProviders:   shareddata.Providers{shareddata.Decimal128s},
		},
		"ArrayTwoValues": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v", "$v"}}}},
				}}},
			},
		},
		"ArrayValueInt": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v", int32(1)}}}},
				}}},
			},
		},
		"ArrayIntLongDoubleStringBool": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{int32(2), int64(3), float64(4), "not-expression", true}}}},
				}}},
			},
		},
		"RecursiveValue": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sumsum", bson.D{{"$sum", bson.D{{"$sum", "$v"}}}}},
				}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1078",
			failsProviders:   shareddata.Providers{shareddata.Decimal128s},
		},
		"RecursiveArrayValue": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sumsum", bson.D{{"$sum", bson.D{{"$sum", bson.A{"$v"}}}}}},
				}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1078",
			failsProviders:   shareddata.Providers{shareddata.Decimal128s},
		},
		"ArrayValueRecursiveInt": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v", bson.D{{"$sum", int32(2)}}}}}},
				}}},
			},
		},
		"ArrayValueAndRecursiveValue": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v", bson.D{{"$sum", "$v"}}}}}},
				}}},
			},
		},
		"ArrayValueAndRecursiveArray": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.A{"$v", bson.D{{"$sum", bson.A{"$v"}}}}}}},
				}}},
			},
		},
		"Type": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sumtype", bson.D{{"$sum", bson.D{{"$type", "$v"}}}}},
				}}},
			},
		},
		"RecursiveEmptyVariable": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.D{{"$sum", "$$$"}}}}},
				}}},
			},
			resultType: emptyResult,
		},
		"MultipleRecursiveEmptyVariable": {
			pipeline: bson.A{
				bson.D{{"$project", bson.D{
					{"sum", bson.D{{"$sum", bson.D{{"$sum", bson.D{{"$sum", "$$$"}}}}}}},
				}}},
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

		"InvalidOperator": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"value", bson.D{{"$invalid-operator", "foo"}}}}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
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
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
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
		"TypeEmpty": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", bson.D{}}}}}}},
			},
		},
		"MultipleOperators": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "foo"}, {"$operator", "foo"}}}}}},
			},
			resultType: emptyResult,
		},
		"MultipleOperatorFirst": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"$type", "foo"}, {"not-operator", "foo"}}}}}},
			},
			resultType: emptyResult,
		},
		"MultipleOperatorLast": {
			pipeline: bson.A{
				bson.D{{"$sort", bson.D{{"_id", -1}}}},
				bson.D{{"$addFields", bson.D{{"type", bson.D{{"not-operator", "foo"}, {"$type", "foo"}}}}}},
			},
			resultType:       emptyResult,
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/349",
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
			resultType: emptyResult,
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
		"SumValue": {
			pipeline: bson.A{
				bson.D{{"$addFields", bson.D{
					{"sum", bson.D{{"$sum", "$v"}}},
				}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1078",
			failsProviders:   shareddata.Providers{shareddata.Decimal128s},
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
		"SumValue": {
			pipeline: bson.A{
				bson.D{{"$set", bson.D{
					{"sum", bson.D{{"$sum", "$v"}}},
				}}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/1078",
			failsProviders:   shareddata.Providers{shareddata.Decimal128s},
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
