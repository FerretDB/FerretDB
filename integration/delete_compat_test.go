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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
)

// deleteCompatTestCase describes delete compatibility test case.
type deleteCompatTestCase struct {
	filters    []bson.D                 // required
	ordered    bool                     // defaults to false
	resultType compatTestCaseResultType // defaults to nonEmptyResult
}

func TestDeleteCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]deleteCompatTestCase{
		"Empty": {
			filters:    []bson.D{},
			resultType: emptyResult,
		},

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
		"DotNotation": {
			filters: []bson.D{
				{{"v.foo.bar", "baz"}},
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
			resultType: emptyResult,
		},
		"UnorderedAllErrors": {
			filters: []bson.D{
				{{"v", bson.D{{"$all", 9}}}},
				{{"v", bson.D{{"$eq", 9}}}},
				{{"v", bson.D{{"$all", 9}}}},
			},
			resultType: emptyResult,
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

			var nonEmptyResults bool
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
						targetErr = UnsetRaw(t, targetErr)
						compatErr = UnsetRaw(t, compatErr)
						assert.Equal(t, compatErr, targetErr)
					} else {
						require.NoError(t, compatErr, "compat error; target returned no error")
					}

					if pointer.Get(targetRes).DeletedCount > 0 || pointer.Get(compatRes).DeletedCount > 0 {
						nonEmptyResults = true
					}

					assert.Equal(t, compatRes, targetRes)

					targetDocs := FindAll(t, ctx, targetCollection)
					compatDocs := FindAll(t, ctx, compatCollection)

					t.Logf("Compat (expected) IDs: %v", CollectIDs(t, compatDocs))
					t.Logf("Target (actual)   IDs: %v", CollectIDs(t, targetDocs))
					AssertEqualDocumentsSlice(t, compatDocs, targetDocs)
				})
			}

			switch tc.resultType {
			case nonEmptyResult:
				assert.True(t, nonEmptyResults, "expected non-empty results (some documents should be deleted)")
			case emptyResult:
				assert.False(t, nonEmptyResults, "expected empty results (no documents should be deleted)")
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}
