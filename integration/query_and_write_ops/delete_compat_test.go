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

package query_and_write_ops

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
)

// deleteCompatTestCase describes delete compatibility test case.
type deleteCompatTestCase struct {
	filters    []bson.D                             // required
	ordered    bool                                 // defaults to false
	resultType integration.CompatTestCaseResultType // defaults to NonEmptyResult
}

func TestDeleteCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]deleteCompatTestCase{
		"One": {
			filters: []bson.D{
				{{"v", int32(42)}},
			},
		},
		"All": {
			filters: []bson.D{
				{},
			},
		},

		"Two": {
			filters: []bson.D{
				{{"v", int32(42)}},
				{{"v", int32(0)}},
			},
		},
		"TwoAll": {
			filters: []bson.D{
				{{"v", int32(42)}},
				{},
			},
		},
		"TwoAllOrdered": {
			filters: []bson.D{
				{{"v", "foo"}},
				{},
			},
			ordered: true,
		},

		"OrderedError": {
			filters: []bson.D{
				{{"v", "foo"}},
				{{"v", bson.D{{"$all", 9}}}},
				{{"v", float32(42.13)}},
			},
			ordered: true,
		},
		"UnorderedError": {
			filters: []bson.D{
				{{"v", "foo"}},
				{{"v", bson.D{{"$all", 9}}}},
				{{"v", float32(42.13)}},
			},
		},

		"OrderedTwoErrors": {
			filters: []bson.D{
				{{"v", "foo"}},
				{{"v", bson.D{{"$all", 9}}}},
				{{"v", float32(42.13)}},
				{{"v", bson.D{{"$eq", 9}}}},
			},
			ordered: true,
		},
		"UnorderedTwoErrors": {
			filters: []bson.D{
				{{"v", "foo"}},
				{{"v", bson.D{{"$all", 9}}}},
				{{"v", float32(42.13)}},
				{{"v", bson.D{{"$eq", 9}}}},
			},
		},

		"OrderedAllErrors": {
			filters: []bson.D{
				{{"v", bson.D{{"$all", 9}}}},
				{{"v", bson.D{{"$eq", 9}}}},
				{{"v", bson.D{{"$all", 9}}}},
			},
			ordered:    true,
			resultType: integration.EmptyResult,
		},
		"UnorderedAllErrors": {
			filters: []bson.D{
				{{"v", bson.D{{"$all", 9}}}},
				{{"v", bson.D{{"$eq", 9}}}},
				{{"v", bson.D{{"$all", 9}}}},
			},
			resultType: integration.EmptyResult,
		},
	}

	testDeleteCompat(t, testCases)
}

// testDeleteCompat tests delete compatibility test cases.
func testDeleteCompat(t *testing.T, testCases map[string]deleteCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			// Use per-test setup because deletes modify data set.
			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			filters := tc.filters
			require.NotNil(t, filters)

			opts := options.BulkWrite().SetOrdered(tc.ordered)

			var NonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					models := make([]mongo.WriteModel, len(filters))
					for i, q := range filters {
						models[i] = mongo.NewDeleteManyModel().SetFilter(q)
					}

					targetRes, targetErr := targetCollection.BulkWrite(ctx, models, opts)
					compatRes, compatErr := compatCollection.BulkWrite(ctx, models, opts)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)

						// error messages are intentionally not compared
						integration.AssertMatchesBulkException(t, compatErr, targetErr)
					} else { // we have to check the results in case of error because some documents may be deleted
						require.NoError(t, compatErr, "compat error; target returned no error")
					}

					if pointer.Get(targetRes).DeletedCount > 0 || pointer.Get(compatRes).DeletedCount > 0 {
						NonEmptyResults = true
					}

					t.Logf("Compat (expected) result: %v", compatRes)
					t.Logf("Target (actual)   result: %v", targetRes)
					assert.Equal(t, compatRes, targetRes)

					targetDocs := integration.FindAll(t, ctx, targetCollection)
					compatDocs := integration.FindAll(t, ctx, compatCollection)

					t.Logf("Compat (expected) IDs: %v", integration.CollectIDs(t, compatDocs))
					t.Logf("Target (actual)   IDs: %v", integration.CollectIDs(t, targetDocs))
					integration.AssertEqualDocumentsSlice(t, compatDocs, targetDocs)
				})
			}

			switch tc.resultType {
			case integration.NonEmptyResult:
				assert.True(t, NonEmptyResults, "expected non-empty results (some documents should be deleted)")
			case integration.EmptyResult:
				assert.False(t, NonEmptyResults, "expected empty results (no documents should be deleted)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}
