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

// distinctCompatTestCase describes count compatibility test case.
type distinctCompatTestCase struct {
	field      string                   // required
	skip       string                   // optional
	filter     bson.D                   // required
	resultType compatTestCaseResultType // defaults to nonEmptyResult
}

func testDistinctCompat(t *testing.T, testCases map[string]distinctCompatTestCase) {
	t.Helper()

	// Use shared setup because distinct queries can't modify data.
	// TODO Use read-only user. https://github.com/FerretDB/FerretDB/issues/1025
	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                shareddata.AllProviders(),
		AddNonExistentCollection: true,
	})
	ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			if tc.skip != "" {
				t.Skip(t, tc.skip)
			}

			t.Parallel()

			filter := tc.filter
			require.NotNil(t, filter, "filter should be set")

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					targetRes, targetErr := targetCollection.Distinct(ctx, tc.field, tc.filter)
					compatRes, compatErr := compatCollection.Distinct(ctx, tc.field, tc.filter)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						targetErr = UnsetRaw(t, targetErr)
						compatErr = UnsetRaw(t, compatErr)
						assert.Equal(t, compatErr, targetErr)
						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					t.Logf("Compat (expected) result: %v", compatRes)
					t.Logf("Target (actual)   result: %v", targetRes)

					require.Equal(t, len(compatRes), len(targetRes))

					// If compat result is an empty array, the target results must be empty array too
					// (not nil or something else).
					if len(compatRes) == 0 {
						assert.Equal(t, compatRes, targetRes)
					}

					// We can't check the exact data types because they might be different.
					// For example, if targetRes is [float64(0), int32(1)] and compatRes is [int64(0), int64(1)],
					// we consider them equal. If different documents use different types to store the same value
					// in the same field, it's hard to predict what type will be returned by distinct.
					// This is why we iterate through results and use assert.EqualValues instead of assert.Equal.
					for i := range compatRes {
						assert.EqualValues(t, compatRes[i], targetRes[i])
					}

					if targetRes != nil || compatRes != nil {
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

func TestDistinctCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]distinctCompatTestCase{
		"EmptyField": {
			field:      "",
			filter:     bson.D{},
			resultType: emptyResult,
		},
		"IDAny": {
			field:  "_id",
			filter: bson.D{},
		},
		"IDString": {
			field:  "_id",
			filter: bson.D{{"_id", "string"}},
		},
		"IDNotExists": {
			field:  "_id",
			filter: bson.D{{"_id", "count-id-not-exists"}},
		},
		"VArray": {
			field:  "v",
			filter: bson.D{{"v", bson.D{{"$type", "array"}}}},
		},
		"VAny": {
			field:  "v",
			filter: bson.D{},
		},
		"NonExistentField": {
			field:  "field-not-exists",
			filter: bson.D{},
		},
		"DotNotation": {
			field:  "v.foo",
			filter: bson.D{},
		},
		"DotNotationArray": {
			field:  "v.array.0",
			filter: bson.D{},
		},
		"DotNotationArrayFirstLevel": {
			field:  "v.0.foo",
			filter: bson.D{},
		},
	}

	testDistinctCompat(t, testCases)
}
