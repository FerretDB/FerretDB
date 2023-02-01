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
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

type queryGetMoreCompatTestCase struct {
	sort      bson.D // defaults to bson.D{{"_id", 1}}
	batchSize int    // defaults to 0
	limit     int    // defaults to 0
}

func testGetMoreCompat(t *testing.T, testCases map[string]queryGetMoreCompatTestCase) {
	t.Helper()

	res := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers: []shareddata.Provider{
			shareddata.Int32BigAmounts,
		},
	})

	ctx, targetCollections, compatCollections := res.Ctx, res.TargetCollections, res.CompatCollections

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Helper()

			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					sort := tc.sort
					if sort == nil {
						sort = bson.D{{"_id", 1}}
					}
					opts := options.Find().SetSort(sort)

					var batchSize int32
					if tc.batchSize != 0 {
						batchSize = int32(tc.batchSize)
					}
					opts = opts.SetBatchSize(batchSize)

					var limit int64
					if tc.limit != 0 {
						limit = int64(tc.limit)
					}
					opts = opts.SetLimit(limit)

					targetResult, targetErr := targetCollection.Find(ctx, bson.D{}, opts)
					compatResult, compatErr := compatCollection.Find(ctx, bson.D{}, opts)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}

					require.NoError(t, targetErr)
					require.NoError(t, compatErr, "compat error; target returned no error")

					// Retrieve all documents from the cursor.
					// Driver will call getMore until the cursor is exhausted.
					var targetRes, compatRes []bson.D
					require.NoError(t, targetResult.All(ctx, &targetRes))
					require.NoError(t, compatResult.All(ctx, &compatRes))

					assert.Equal(t, len(compatRes), len(targetRes), "result length mismatch")
				})
			}
		})
	}
}

func TestGetMoreCompat(t *testing.T) {
	testCases := map[string]queryGetMoreCompatTestCase{
		"id": {
			batchSize: 200,
		},
		"getMoreWithLimitLessThanBatch": {
			batchSize: 200,
			limit:     100,
		},
		"getMoreWithLimitGreaterThanBatch": {
			batchSize: 200,
			limit:     300,
		},
	}

	testGetMoreCompat(t, testCases)
}

type queryGetMoreErrorsCompatTestCase struct {
	id         any    // defaults to id fetched from running find command
	altMessage string // expects different error altMessage on target
	command    bson.D // defaults to empty
	err        bool   // expects error if it is set to true
}

func testGetMoreCompatErrors(t *testing.T, testCases map[string]queryGetMoreErrorsCompatTestCase) {
	t.Helper()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers: []shareddata.Provider{shareddata.Int32BigAmounts},
	})

	// We expect to have only one collection as the result of setup.
	require.Len(t, s.TargetCollections, 1)
	require.Len(t, s.CompatCollections, 1)

	targetCollection := s.TargetCollections[0]
	compatCollection := s.CompatCollections[0]

	ctx := s.Ctx

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			targetID := tc.id
			compatID := tc.id

			if tc.id == nil {
				targetID = getCursorIDAndFirstBatch(t, ctx, targetCollection)
				compatID = getCursorIDAndFirstBatch(t, ctx, compatCollection)
			}
			targetCommand := bson.D{{"getMore", targetID}, {"collection", targetCollection.Name()}}
			targetCommand = append(targetCommand, tc.command...)
			compatCommand := bson.D{{"getMore", compatID}, {"collection", compatCollection.Name()}}
			compatCommand = append(compatCommand, tc.command...)

			var targetResult, compatResult bson.D
			targetErr := targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetResult)
			compatErr := compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatResult)
			if tc.err {
				var compatCommandErr mongo.CommandError
				if !errors.As(compatErr, &compatCommandErr) {
					t.Fatalf("Expected error of type %T, got %T", compatCommandErr, compatErr)
				}
				compatCommandErr.Raw = nil

				if strings.Contains(compatCommandErr.Message, "Cannot run getMore on cursor") {
					t.Skip("TODO: https://github.com/FerretDB/FerretDB/issues/1807")
				}

				AssertEqualAltError(t, compatCommandErr, tc.altMessage, targetErr)

				return
			}

			require.NoError(t, targetErr)
			require.NoError(t, compatErr)

			targetDoc := ConvertDocument(t, targetResult)
			compatDoc := ConvertDocument(t, compatResult)

			targetCursor := must.NotFail(targetDoc.Get("cursor")).(*types.Document)
			compatCursor := must.NotFail(compatDoc.Get("cursor")).(*types.Document)

			targetNextBatch := must.NotFail(targetCursor.Get("nextBatch")).(*types.Array)
			compatNextBatch := must.NotFail(compatCursor.Get("nextBatch")).(*types.Array)

			require.Equal(t, compatNextBatch.Len(), targetNextBatch.Len(), "result length mismatch")

			targetNS, err := targetCursor.Get("ns")
			require.NoError(t, err)

			compatNS, err := compatCursor.Get("ns")
			require.NoError(t, err)

			require.Equal(t, targetNS, compatNS, "ns mismatch")
		})
	}
}

func TestGetMoreErrorsCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]queryGetMoreErrorsCompatTestCase{
		"InvalidCursorID": {
			id:         int64(2),
			err:        true,
			altMessage: "cursor id 2 not found",
		},
		"CursorIdInt32": {
			id:  int32(1),
			err: true,
		},
		"CursorIDNegative": {
			id:  int64(-1),
			err: true,
		},
		"BatchSizeDocument": {
			command: bson.D{
				{"batchSize", bson.D{}},
			},
			err:        true,
			altMessage: "BSON field 'batchSize' is the wrong type 'object', expected type 'long'",
		},
		"BatchSizeArray": {
			command: bson.D{
				{"batchSize", bson.A{}},
			},
			err:        true,
			altMessage: "BSON field 'batchSize' is the wrong type 'array', expected type 'long'",
		},
		"BatchSizeNegative": {
			command: bson.D{
				{"batchSize", int64(-1)},
			},
			err:        true,
			altMessage: "cursor id -1 not found",
		},
		"BatchSizeResponse": {
			command: bson.D{
				{"batchSize", int64(300)},
			},
		},
	}

	testGetMoreCompatErrors(t, testCases)
}
