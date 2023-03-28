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

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// queryCompatTestCase describes query compatibility test case.
type queryCompatTestCase struct {
	filter         bson.D                   // required
	sort           bson.D                   // defaults to `bson.D{{"_id", 1}}`
	limit          *int64                   // defaults to nil to leave unset
	optSkip        *int64                   // defaults to nil to leave unset
	projection     bson.D                   // nil for leaving projection unset
	resultType     compatTestCaseResultType // defaults to nonEmptyResult
	resultPushdown bool                     // defaults to false

	skip          string // skip test for all handlers, must have issue number mentioned
	skipForTigris string // skip test for Tigris
}

// testQueryCompat tests query compatibility test cases.
func testQueryCompat(t *testing.T, testCases map[string]queryCompatTestCase) {
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

			t.Parallel()

			filter := tc.filter
			require.NotNil(t, filter, "filter should be set")

			opts := options.Find()

			opts.SetSort(tc.sort)
			if tc.sort == nil {
				opts.SetSort(bson.D{{"_id", 1}})
			}

			if tc.limit != nil {
				opts.SetLimit(*tc.limit)
			}

			if tc.optSkip != nil {
				opts.SetSkip(*tc.optSkip)
			}

			if tc.projection != nil {
				opts.SetProjection(tc.projection)
			}

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

					targetCursor, targetErr := targetCollection.Find(ctx, filter, opts)
					compatCursor, compatErr := compatCollection.Find(ctx, filter, opts)

					if targetCursor != nil {
						defer targetCursor.Close(ctx)
					}
					if compatCursor != nil {
						defer compatCursor.Close(ctx)
					}

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					var targetRes, compatRes []bson.D
					require.NoError(t, targetCursor.All(ctx, &targetRes))
					require.NoError(t, compatCursor.All(ctx, &compatRes))

					t.Logf("Compat (expected) IDs: %v", CollectIDs(t, compatRes))
					t.Logf("Target (actual)   IDs: %v", CollectIDs(t, targetRes))
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

func TestQueryCompatFilter(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Empty": {
			filter: bson.D{},
		},
		"IDString": {
			filter:         bson.D{{"_id", "string"}},
			resultPushdown: true,
		},
		"IDObjectID": {
			filter:         bson.D{{"_id", primitive.NilObjectID}},
			resultPushdown: true,
		},
		"ObjectID": {
			filter:         bson.D{{"v", primitive.NilObjectID}},
			resultPushdown: true,
		},
		"UnknownFilterOperator": {
			filter:     bson.D{{"v", bson.D{{"$someUnknownOperator", 42}}}},
			resultType: emptyResult,
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryCompatSort(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Asc": {
			filter: bson.D{},
			sort:   bson.D{{"v", 1}, {"_id", 1}},
		},
		"Desc": {
			filter: bson.D{},
			sort:   bson.D{{"v", -1}, {"_id", 1}},
		},

		"Bad": {
			filter:     bson.D{},
			sort:       bson.D{{"v", 13}},
			resultType: emptyResult,
		},
		"BadZero": {
			filter:     bson.D{},
			sort:       bson.D{{"v", 0}},
			resultType: emptyResult,
		},
		"BadNull": {
			filter:     bson.D{},
			sort:       bson.D{{"v", nil}},
			resultType: emptyResult,
		},

		"DotNotation": {
			filter: bson.D{},
			sort:   bson.D{{"v.foo", 1}, {"_id", 1}},
		},
		"DotNotationIndex": {
			filter: bson.D{},
			sort:   bson.D{{"v.0", 1}, {"_id", 1}},
		},
		"DotNotationNonExistent": {
			filter: bson.D{},
			sort:   bson.D{{"invalid.foo", 1}, {"_id", 1}},
		},
		"DotNotationMissingField": {
			filter:     bson.D{},
			sort:       bson.D{{"v..foo", 1}, {"_id", 1}},
			resultType: emptyResult,
		},

		"BadDollarStart": {
			filter:     bson.D{},
			sort:       bson.D{{"$v.foo", 1}},
			resultType: emptyResult,

			skip: "https://github.com/FerretDB/FerretDB/issues/2259",
		},
		"BadDollarMid": {
			filter:     bson.D{},
			sort:       bson.D{{"v.$foo.bar", 1}},
			resultType: emptyResult,

			skip: "https://github.com/FerretDB/FerretDB/issues/2259",
		},
		"BadDollarEnd": {
			filter:     bson.D{},
			sort:       bson.D{{"v.$foo", 1}},
			resultType: emptyResult,

			skip: "https://github.com/FerretDB/FerretDB/issues/2259",
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryCompatLimit(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Simple": {
			filter: bson.D{},
			limit:  pointer.ToInt64(1),
		},
		"AlmostAll": {
			filter: bson.D{},
			limit:  pointer.ToInt64(int64(len(shareddata.Strings.Docs()) - 1)),
		},
		"All": {
			filter: bson.D{},
			limit:  pointer.ToInt64(int64(len(shareddata.Strings.Docs()))),
		},
		"More": {
			filter: bson.D{},
			limit:  pointer.ToInt64(int64(len(shareddata.Strings.Docs()) + 1)),
		},
		"Big": {
			filter: bson.D{},
			limit:  pointer.ToInt64(1000),
		},
		"Zero": {
			filter: bson.D{},
			limit:  pointer.ToInt64(0),
		},
		"SingleBatch": {
			// The meaning of negative limits is redefined by the Go driver:
			// > A negative limit specifies that the resulting documents should be returned in a single batch.
			// On the wire, "limit" can't be negative.
			// TODO https://github.com/FerretDB/FerretDB/issues/2255
			filter: bson.D{},
			limit:  pointer.ToInt64(-1),
		},
	}

	testQueryCompat(t, testCases)
}

func TestQueryCompatSkip(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryCompatTestCase{
		"Simple": {
			filter:  bson.D{},
			optSkip: pointer.ToInt64(1),
		},
		"AlmostAll": {
			filter:  bson.D{},
			optSkip: pointer.ToInt64(int64(len(shareddata.Strings.Docs()) - 1)),
		},
		"All": {
			filter:  bson.D{},
			optSkip: pointer.ToInt64(int64(len(shareddata.Strings.Docs()))),
		},
		"More": {
			filter:  bson.D{},
			optSkip: pointer.ToInt64(int64(len(shareddata.Strings.Docs()) + 1)),
		},
		"Big": {
			filter:     bson.D{},
			optSkip:    pointer.ToInt64(1000),
			resultType: emptyResult,
		},
		"Zero": {
			filter:  bson.D{},
			optSkip: pointer.ToInt64(0),
		},
		"Bad": {
			filter:     bson.D{},
			optSkip:    pointer.ToInt64(-1),
			resultType: emptyResult,
		},
	}

	testQueryCompat(t, testCases)
}
