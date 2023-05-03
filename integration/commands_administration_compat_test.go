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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

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
		resultType compatTestCaseResultType
		altMessage string
	}{
		"scaleOne":           {scale: int32(1)},
		"scaleBig":           {scale: int64(1000)},
		"scaleMaxInt":        {scale: math.MaxInt64},
		"scaleZero":          {scale: int32(0), resultType: emptyResult},
		"scaleNegative":      {scale: int32(-100), resultType: emptyResult},
		"scaleFloat":         {scale: 2.8},
		"scaleFloatNegative": {scale: -2.8, resultType: emptyResult},
		"scaleMinFloat":      {scale: -math.MaxFloat64, resultType: emptyResult},
		"scaleMaxFloat":      {scale: math.MaxFloat64},
		"scaleString": {
			scale:      "1",
			resultType: emptyResult,
			altMessage: `BSON field 'collStats.scale' is the wrong type 'string', expected types '[long, int, decimal, double]'`,
		},
		"scaleObject": {
			scale:      bson.D{{"a", 1}},
			resultType: emptyResult,
			altMessage: `BSON field 'collStats.scale' is the wrong type 'object', expected types '[long, int, decimal, double]'`,
		},
		"scaleNull": {scale: nil},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			var targetRes bson.D
			targetCommand := bson.D{{"collStats", targetCollection.Name()}, {"scale", tc.scale}}
			targetErr := targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetRes)

			var compatRes bson.D
			compatCommand := bson.D{{"collStats", compatCollection.Name()}, {"scale", tc.scale}}
			compatErr := compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatRes)

			if tc.resultType == emptyResult {
				require.Error(t, compatErr)

				targetErr = UnsetRaw(t, targetErr)
				compatErr = UnsetRaw(t, compatErr)

				if tc.altMessage != "" {
					var expectedErr mongo.CommandError
					require.ErrorAs(t, compatErr, &expectedErr)
					AssertEqualAltError(t, expectedErr, tc.altMessage, targetErr)
				} else {
					assert.Equal(t, compatErr, targetErr)
				}

				return
			}

			require.NoError(t, compatErr)
			require.NoError(t, targetErr)

			targetDoc := ConvertDocument(t, targetRes)
			compatDoc := ConvertDocument(t, compatRes)

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
		resultType compatTestCaseResultType
		altMessage string
	}{
		"scaleOne":   {scale: int32(1)},
		"scaleBig":   {scale: int64(1000)},
		"scaleFloat": {scale: 2.8},
		"scaleNull":  {scale: nil},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			var targetRes bson.D
			targetCommand := bson.D{{"dbStats", int32(1)}, {"scale", tc.scale}}
			targetErr := targetCollection.Database().RunCommand(ctx, targetCommand).Decode(&targetRes)

			var compatRes bson.D
			compatCommand := bson.D{{"dbStats", int32(1)}, {"scale", tc.scale}}
			compatErr := compatCollection.Database().RunCommand(ctx, compatCommand).Decode(&compatRes)

			if tc.resultType == emptyResult {
				require.Error(t, compatErr)

				targetErr = UnsetRaw(t, targetErr)
				compatErr = UnsetRaw(t, compatErr)

				if tc.altMessage != "" {
					var expectedErr mongo.CommandError
					require.ErrorAs(t, compatErr, &expectedErr)
					AssertEqualAltError(t, expectedErr, tc.altMessage, targetErr)
				} else {
					assert.Equal(t, compatErr, targetErr)
				}

				return
			}

			require.NoError(t, compatErr)
			require.NoError(t, targetErr)

			targetDoc := ConvertDocument(t, targetRes)
			compatDoc := ConvertDocument(t, compatRes)

			targetFactor := must.NotFail(targetDoc.Get("scaleFactor"))
			compatFactor := must.NotFail(compatDoc.Get("scaleFactor"))

			assert.Equal(t, compatFactor, targetFactor)
		})
	}
}
