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

// distinctCompatTestCase describes distinct compatibility test case.
type distinctCompatTestCase struct {
	field      string                   // required
	filter     bson.D                   // required
	resultType compatTestCaseResultType // defaults to nonEmptyResult
}

func testDistinctCompat(tt *testing.T, testCases map[string]distinctCompatTestCase) {
	tt.Helper()

	// Use shared setup because distinct queries can't modify data.
	//
	// Use read-only user.
	// TODO https://github.com/FerretDB/FerretDB/issues/1025
	s := setup.SetupCompatWithOpts(tt, &setup.SetupCompatOpts{
		Providers:                shareddata.AllProviders().Remove(shareddata.Scalars), // Remove provider with the same values with different types
		AddNonExistentCollection: true,
	})
	ctx, targetCollections, compatCollections := s.Ctx, s.TargetCollections, s.CompatCollections

	for name, tc := range testCases {
		name, tc := name, tc
		tt.Run(name, func(tt *testing.T) {
			tt.Helper()

			tt.Parallel()

			filter := tc.filter
			require.NotNil(tt, filter, "filter should be set")

			var nonEmptyResults bool
			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				tt.Run(targetCollection.Name(), func(tt *testing.T) {
					tt.Helper()

					t := setup.FailsForSQLite(tt, "https://github.com/FerretDB/FerretDB/issues/3157")

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

					assert.Equal(t, targetRes, compatRes)

					if targetRes != nil || compatRes != nil {
						nonEmptyResults = true
					}
				})
			}

			// TODO https://github.com/FerretDB/FerretDB/issues/3157
			if setup.IsSQLite(tt) {
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
