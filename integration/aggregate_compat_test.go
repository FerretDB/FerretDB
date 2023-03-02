package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
)

// aggregateCompatTestCase describes aggregate compatibility test case.
type aggregateCompatTestCase struct {
	pipeline   any                      // required
	resultType compatTestCaseResultType // defaults to nonEmptyResult
	skip       string                   // skip test for all handlers, must have issue number mentioned

}

// testAggregateCompat tests aggregate pipeline compatibility test cases.
func testAggregateCompat(t *testing.T, testCases map[string]aggregateCompatTestCase) {
	t.Helper()

	ctx, targetCollections, compatCollections := setup.SetupCompat(t)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			pipeline := tc.pipeline
			require.NotNil(t, pipeline, "pipeline should be set")

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					if tc.skip != "" {
						t.Skip(tc.skip)
					}

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

func TestAggregateCompatSort(t *testing.T) {
	testCases := map[string]aggregateCompatTestCase{
		"AscendingID": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{{"_id", 1}}}}},
		},
		"DescendingID": {
			pipeline: bson.A{bson.D{{"$sort", bson.D{{"_id", -1}}}}},
		},
		"InvalidSortOrder": {
			pipeline:   bson.A{bson.D{{"$sort", bson.D{{"_id", 0}}}}},
			resultType: emptyResult,
		},
		"Location15976": {
			pipeline:   bson.A{bson.D{{"$sort", bson.D{}}}},
			resultType: emptyResult,
		},
		"Location15973": {
			pipeline:   bson.A{bson.D{{"$sort", 1}}},
			resultType: emptyResult,
		},
		"Sort": {
			pipeline:   bson.A{bson.D{{"$sort", bson.A{}}}},
			resultType: emptyResult,
		},
	}

	testAggregateCompat(t, testCases)
}
