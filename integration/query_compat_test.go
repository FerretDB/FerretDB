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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

// queryCompatTestCase describes query compatibility test case.
type queryCompatTestCase struct {
	filter     bson.D                   // required
	sort       bson.D                   // defaults to `bson.D{{"_id", 1}}`
	resultType compatTestCaseResultType // defaults to nonEmptyResult
}

// testQueryCompat tests query compatibility test cases.
func testQueryCompat(t *testing.T, testCases map[string]queryCompatTestCase) {
	t.Helper()

	providers := []shareddata.Provider{
		shareddata.FixedScalars,
		shareddata.Scalars,
		shareddata.Composites,
	}

	// Use shared setup because find queries can't modify data.
	// TODO use read-only user https://github.com/FerretDB/FerretDB/issues/914
	ctx, targetCollection, compatCollection := setup.SetupCompat(t, providers...)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			filter := tc.filter
			require.NotNil(t, filter)

			sort := tc.sort
			if sort == nil {
				sort = bson.D{{"_id", 1}}
			}
			opts := options.Find().SetSort(sort)

			targetCursor, targetErr := targetCollection.Find(ctx, filter, opts)
			compatCursor, compatErr := compatCollection.Find(ctx, filter, opts)

			if targetCursor != nil {
				defer targetCursor.Close(ctx)
			}
			if compatCursor != nil {
				defer compatCursor.Close(ctx)
			}

			if targetErr != nil {
				targetErr = UnsetRaw(t, targetErr)
				compatErr = UnsetRaw(t, compatErr)
				assert.Equal(t, errorResult, tc.resultType)
				assert.Equal(t, compatErr, targetErr)
				return
			}
			require.NoError(t, compatErr)

			var targetRes, compatRes []bson.D
			require.NoError(t, targetCursor.All(ctx, &targetRes))
			require.NoError(t, compatCursor.All(ctx, &compatRes))

			t.Logf("Compat (expected) IDs: %v", CollectIDs(t, compatRes))
			t.Logf("Target (actual)   IDs: %v", CollectIDs(t, targetRes))

			switch tc.resultType {
			case nonEmptyResult:
				assert.NotEmpty(t, compatRes)
				assert.NotEmpty(t, targetRes)
				AssertEqualDocumentsSlice(t, compatRes, targetRes)
			case emptyResult:
				assert.Empty(t, compatRes)
				assert.Empty(t, targetRes)
			case errorResult:
				fallthrough
			default:
				t.Fatalf("unknown result type %v", tc.resultType)
			}
		})
	}
}
