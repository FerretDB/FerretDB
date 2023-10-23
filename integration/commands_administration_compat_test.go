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
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
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
		},
		"scaleObject": {
			scale:      bson.D{{"a", 1}},
			resultType: emptyResult,
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

			if targetErr != nil {
				t.Logf("Target error: %v", targetErr)
				t.Logf("Compat error: %v", compatErr)

				// error messages are intentionally not compared
				AssertMatchesCommandError(t, compatErr, targetErr)

				return
			}
			require.NoError(t, compatErr, "compat error; target returned no error")

			targetDoc := ConvertDocument(t, targetRes)
			compatDoc := ConvertDocument(t, compatRes)

			targetFactor := must.NotFail(targetDoc.Get("scaleFactor"))
			compatFactor := must.NotFail(compatDoc.Get("scaleFactor"))

			assert.Equal(t, compatFactor, targetFactor)
		})
	}
}

func TestCommandsAdministrationCompatCollStatsCappedCollection(t *testing.T) {
	t.Skip("https://github.com/FerretDB/FerretDB/issues/2447")

	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                []shareddata.Provider{},
		AddNonExistentCollection: true,
	})

	ctx, targetCollection, compatCollection := s.Ctx, s.TargetCollections[0], s.CompatCollections[0]

	for name, tc := range map[string]struct { //nolint:vet // for readability
		sizeInBytes  int64 // also sets capped true if it is greater than zero
		maxDocuments int64 // maxDocuments is set if sizeInBytes is greater than zero
	}{
		"Size": {
			sizeInBytes: 1000,
		},
		"MaxDocuments": {
			sizeInBytes:  1000,
			maxDocuments: 10,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cName := testutil.CollectionName(t) + name
			opts := options.CreateCollection()

			if tc.sizeInBytes > 0 {
				opts.SetCapped(true)
				opts.SetSizeInBytes(tc.sizeInBytes)

				if tc.maxDocuments > 0 {
					opts.SetMaxDocuments(tc.maxDocuments)
				}
			}

			targetErr := targetCollection.Database().CreateCollection(ctx, cName, opts)
			require.NoError(t, targetErr)

			compatErr := compatCollection.Database().CreateCollection(ctx, cName, opts)
			require.NoError(t, compatErr)

			require.Equal(t, compatCollection.Name(), targetCollection.Name())
			command := bson.D{{"collStats", targetCollection.Name()}}

			var targetRes bson.D
			targetErr = targetCollection.Database().RunCommand(ctx, command).Decode(&targetRes)
			require.NoError(t, targetErr)

			var compatRes bson.D
			targetErr = compatCollection.Database().RunCommand(ctx, command).Decode(&targetRes)
			require.NoError(t, targetErr)

			targetDoc := ConvertDocument(t, targetRes)
			compatDoc := ConvertDocument(t, compatRes)

			assert.Equal(t, must.NotFail(compatDoc.Get("capped")), must.NotFail(targetDoc.Get("capped")))
			assert.Equal(t, must.NotFail(compatDoc.Get("max")), must.NotFail(targetDoc.Get("max")))
			assert.Equal(t, must.NotFail(compatDoc.Get("maxSize")), must.NotFail(targetDoc.Get("maxSize")))
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

			if targetErr != nil {
				t.Logf("Target error: %v", targetErr)
				t.Logf("Compat error: %v", compatErr)

				// error messages are intentionally not compared
				AssertMatchesCommandError(t, compatErr, targetErr)

				return
			}
			require.NoError(t, compatErr, "compat error; target returned no error")

			targetDoc := ConvertDocument(t, targetRes)
			compatDoc := ConvertDocument(t, compatRes)

			targetFactor := must.NotFail(targetDoc.Get("scaleFactor"))
			compatFactor := must.NotFail(compatDoc.Get("scaleFactor"))

			assert.Equal(t, compatFactor, targetFactor)
		})
	}
}

func TestCommandsAdministrationCompatDBStatsFreeStorage(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                []shareddata.Provider{shareddata.DocumentsDocuments},
		AddNonExistentCollection: true,
	})

	ctx, targetCollection, compatCollection := s.Ctx, s.TargetCollections[0], s.CompatCollections[0]

	for name, tc := range map[string]struct { //nolint:vet // for readability
		command bson.D // required, command to run
		skip    string // optional, skip test with a specified reason
	}{
		"Unset": {
			command: bson.D{{"dbStats", int32(1)}},
		},
		"Int32Zero": {
			command: bson.D{{"dbStats", int32(1)}, {"freeStorage", int32(0)}},
		},
		"Int32One": {
			command: bson.D{{"dbStats", int32(1)}, {"freeStorage", int32(1)}},
		},
		"Int32Negative": {
			command: bson.D{{"dbStats", int32(1)}, {"freeStorage", int32(-1)}},
		},
		"True": {
			command: bson.D{{"dbStats", int32(1)}, {"freeStorage", true}},
		},
		"False": {
			command: bson.D{{"dbStats", int32(1)}, {"freeStorage", false}},
		},
		"Nil": {
			command: bson.D{{"dbStats", int32(1)}, {"freeStorage", nil}},
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var targetRes bson.D
			targetErr := targetCollection.Database().RunCommand(ctx, tc.command).Decode(&targetRes)

			var compatRes bson.D
			compatErr := compatCollection.Database().RunCommand(ctx, tc.command).Decode(&compatRes)

			if targetErr != nil {
				t.Logf("Target error: %v", targetErr)
				t.Logf("Compat error: %v", compatErr)

				// error messages are intentionally not compared
				AssertMatchesCommandError(t, compatErr, targetErr)

				return
			}
			require.NoError(t, compatErr, "compat error; target returned no error")

			targetDoc := ConvertDocument(t, targetRes)
			compatDoc := ConvertDocument(t, compatRes)

			assert.Equal(t, compatDoc.Has("freeStorageSize"), targetDoc.Has("freeStorageSize"))
			assert.Equal(t, compatDoc.Has("totalFreeStorageSize"), targetDoc.Has("totalFreeStorageSize"))
		})
	}
}
