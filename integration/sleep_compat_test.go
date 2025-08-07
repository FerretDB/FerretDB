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
	"time"

	"github.com/FerretDB/FerretDB/v2/integration/setup"
	"github.com/FerretDB/FerretDB/v2/integration/shareddata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

type sleepCompatTestCase struct {
	request bson.D
}

func TestSleepCompat(t *testing.T) {
	t.Parallel()

	testCases := map[string]sleepCompatTestCase{
		"Millis": {
			request: bson.D{
				{"sleep", int32(1)},
				{"millis", int32(500)},
			},
		},
		"Secs": {
			request: bson.D{
				{"sleep", int32(1)},
				{"secs", int32(1)},
			},
		},
		"Default": {
			request: bson.D{
				{"sleep", int32(1)},
			},
		},
	}

	testSleepCompat(t, testCases)
}

func testSleepCompat(t *testing.T, testCases map[string]sleepCompatTestCase) {
	t.Helper()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers: shareddata.Providers{shareddata.Bools},
	})

	ctx := s.Ctx
	targetDB := s.TargetCollections[0].Database().Client().Database("admin")
	compatDB := s.CompatCollections[0].Database().Client().Database("admin")

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			var targetRes, compatRes bson.D

			timeBefore := time.Now()
			targetErr := targetDB.RunCommand(ctx, tc.request).Decode(&targetRes)

			targetDuration := time.Since(timeBefore)

			timeBefore = time.Now()
			compatErr := compatDB.RunCommand(ctx, tc.request).Decode(&compatRes)

			compatDuration := time.Since(timeBefore)

			if targetErr != nil {
				t.Logf("Target error: %v", targetErr)
				t.Logf("Compat error: %v", compatErr)

				targetErr = UnsetRaw(t, targetErr)
				compatErr = UnsetRaw(t, compatErr)
				assert.Equal(t, compatErr, targetErr)
				return
			}
			require.NoError(t, compatErr, "compat error; target returned no error")

			t.Logf("Compat (expected) result: %v", compatRes)
			t.Logf("Target (actual)   result: %v", targetRes)

			AssertEqualDocuments(t, compatRes, targetRes)

			assert.InDelta(t, compatDuration.Milliseconds(), targetDuration.Milliseconds(), 100, "Compat and target sleep durations should be approximately equal")
		})
	}

}
