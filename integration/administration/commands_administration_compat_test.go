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

package administration

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/must"
)

func TestCommandsAdministrationCompatCollStatsWithScale(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                []shareddata.Provider{shareddata.DocumentsDocuments},
		AddNonExistentCollection: true,
	})

	ctx, targetCollection, compatCollection := s.Ctx, s.TargetCollections[0], s.CompatCollections[0]

	for name, tc := range map[string]struct { //nolint:vet // for readability
		scale      any
		resultType integration.CompatTestCaseResultType
	}{
		"scaleOne":           {scale: int32(1)},
		"scaleBig":           {scale: int64(1000)},
		"scaleMaxInt":        {scale: math.MaxInt64},
		"scaleZero":          {scale: int32(0), resultType: integration.EmptyResult},
		"scaleNegative":      {scale: int32(-100), resultType: integration.EmptyResult},
		"scaleFloat":         {scale: 2.8},
		"scaleFloatNegative": {scale: -2.8, resultType: integration.EmptyResult},
		"scaleMinFloat":      {scale: -math.MaxFloat64, resultType: integration.EmptyResult},
		"scaleMaxFloat":      {scale: math.MaxFloat64},
		"scaleString": {
			scale:      "1",
			resultType: integration.EmptyResult,
		},
		"scaleObject": {
			scale:      bson.D{{"a", 1}},
			resultType: integration.EmptyResult,
		},
		"scaleNull": {scale: nil},
	} {
		name, tc := name, tc

		t.Run(name, func(tt *testing.T) {
			tt.Helper()

			tt.Parallel()
			t := setup.FailsForSQLite(tt, "https://github.com/FerretDB/FerretDB/issues/2775")

			var targetRes bson.D
			targetCommand := bson.D{{"collStats", targetCollection.Name()}, {"scale", tc.scale}}
			targetErr := targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetRes)

			var compatRes bson.D
			compatCommand := bson.D{{"collStats", compatCollection.Name()}, {"scale", tc.scale}}
			compatErr := compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatRes)

			if targetErr != nil {
				t.Logf("Target error: %v", targetErr)
				t.Logf("Compat error: %v", compatErr)

				// error messages are intentionally not compared
				integration.AssertMatchesCommandError(t, compatErr, targetErr)

				return
			}
			require.NoError(t, compatErr, "compat error; target returned no error")

			targetDoc := integration.ConvertDocument(t, targetRes)
			compatDoc := integration.ConvertDocument(t, compatRes)

			targetFactor := must.NotFail(targetDoc.Get("scaleFactor"))
			compatFactor := must.NotFail(compatDoc.Get("scaleFactor"))

			assert.Equal(t, compatFactor, targetFactor)
		})
	}
}

func TestCommandsAdministrationCompatDBStatsWithScale(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                []shareddata.Provider{shareddata.DocumentsDocuments},
		AddNonExistentCollection: true,
	})

	ctx, targetCollection, compatCollection := s.Ctx, s.TargetCollections[0], s.CompatCollections[0]

	for name, tc := range map[string]struct { //nolint:vet // for readability
		scale      any
		resultType integration.CompatTestCaseResultType
	}{
		"scaleOne":   {scale: int32(1)},
		"scaleBig":   {scale: int64(1000)},
		"scaleFloat": {scale: 2.8},
		"scaleNull":  {scale: nil},
	} {
		name, tc := name, tc

		t.Run(name, func(tt *testing.T) {
			tt.Helper()

			tt.Parallel()

			t := setup.FailsForSQLite(tt, "https://github.com/FerretDB/FerretDB/issues/2775")

			var targetRes bson.D
			targetCommand := bson.D{{"dbStats", int32(1)}, {"scale", tc.scale}}
			targetErr := targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetRes)

			var compatRes bson.D
			compatCommand := bson.D{{"dbStats", int32(1)}, {"scale", tc.scale}}
			compatErr := compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatRes)

			if targetErr != nil {
				t.Logf("Target error: %v", targetErr)
				t.Logf("Compat error: %v", compatErr)

				// error messages are intentionally not compared
				integration.AssertMatchesCommandError(t, compatErr, targetErr)

				return
			}
			require.NoError(t, compatErr, "compat error; target returned no error")

			targetDoc := integration.ConvertDocument(t, targetRes)
			compatDoc := integration.ConvertDocument(t, compatRes)

			targetFactor := must.NotFail(targetDoc.Get("scaleFactor"))
			compatFactor := must.NotFail(compatDoc.Get("scaleFactor"))

			assert.Equal(t, compatFactor, targetFactor)
		})
	}
}
