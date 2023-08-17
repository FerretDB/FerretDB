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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
)

type insertCommandCompatTestCase struct { //nolint:vet // for readability
	toInsert []any // required, slice of bson.D to insert
	ordered  any   // required, sets it to `ordered`

	skip string // optional, skip test with a specified reason
}

// testInsertCommandCompat tests insert compatibility test cases.
// It uses runCommand instead of insertOne or insertMany to let more parameters being used.
// Unlike testInsertCompat, it does not check inserted IDs.
func testInsertCommandCompat(t *testing.T, testCases map[string]insertCommandCompatTestCase) {
	t.Helper()

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Helper()
			t.Parallel()

			require.NotNil(t, tc.toInsert, "toInsert must not be nil")
			require.NotNil(t, tc.ordered, "ordered must not be nil")

			ctx, targetCollections, compatCollections := setup.SetupCompat(t)

			for i := range targetCollections {
				targetCollection := targetCollections[i]
				compatCollection := compatCollections[i]
				t.Run(targetCollection.Name(), func(t *testing.T) {
					t.Helper()

					// RunCommand must be used to give ability set various parameters
					// and prevent the driver from doing any validation.
					var targetRes, compatRes bson.D
					targetErr := targetCollection.Database().RunCommand(ctx, bson.D{
						{"insert", targetCollection.Name()},
						{"documents", tc.toInsert},
						{"ordered", tc.ordered},
					}).Decode(&targetRes)
					compatErr := compatCollection.Database().RunCommand(ctx, bson.D{
						{"insert", compatCollection.Name()},
						{"documents", tc.toInsert},
						{"ordered", tc.ordered},
					}).Decode(&compatRes)

					if targetErr != nil {
						t.Logf("Target error: %v", targetErr)
						t.Logf("Compat error: %v", compatErr)

						// error messages are intentionally not compared
						integration.AssertMatchesCommandError(t, compatErr, targetErr)

						return
					}
					require.NoError(t, compatErr, "compat error; target returned no error")

					t.Logf("Compat (expected) result: %v", compatRes)
					t.Logf("Target (actual)   result: %v", targetRes)
					assert.Equal(t, compatRes, targetRes)
				})
			}
		})
	}
}

func TestInsertCommandCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]insertCommandCompatTestCase{
		"InsertEmpty": {
			toInsert: []any{
				bson.D{{}},
			},
			ordered: true,
		},
	}

	testInsertCommandCompat(t, testCases)
}
