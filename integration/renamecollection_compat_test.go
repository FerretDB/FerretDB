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

func TestRenameCollectionCompat(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Command renameCollection is not supported for Tigris")

	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                []shareddata.Provider{shareddata.DocumentsDocuments, shareddata.Bools},
		AddNonExistentCollection: true,
	})

	ctx, targetCollection, compatCollection := s.Ctx, s.TargetCollections[0], s.CompatCollections[0]
	targetCollectionExists, compatCollectionExists := s.TargetCollections[1], s.CompatCollections[1]

	targetDB := targetCollection.Database()
	compatDB := compatCollection.Database()

	// Rename collection should be performed while connecting to the admin database.
	targetDBConnect := targetCollection.Database().Client().Database("admin")
	compatDBConnect := compatCollection.Database().Client().Database("admin")

	for name, tc := range map[string]struct {
		targetNSFrom any
		compatNSFrom any
		targetNSTo   any
		compatNSTo   any
		resultType   compatTestCaseResultType
	}{
		"Valid": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo:   targetDB.Name() + ".newCollection",
			compatNSTo:   targetDB.Name() + ".newCollection",
		},
		"NilFrom": {
			targetNSFrom: nil,
			compatNSFrom: nil,
			targetNSTo:   targetDB.Name() + ".newCollection",
			compatNSTo:   compatDB.Name() + ".newCollection",
			resultType:   emptyResult,
		},
		"NilTo": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo:   nil,
			compatNSTo:   nil,
			resultType:   emptyResult,
		},
		"BadTypeFrom": {
			targetNSFrom: int32(42),
			compatNSFrom: int32(42),
			targetNSTo:   targetDB.Name() + ".newCollection",
			compatNSTo:   compatDB.Name() + ".newCollection",
			resultType:   emptyResult,
		},
		"BadTypeTo": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo:   int32(42),
			compatNSTo:   int32(42),
			resultType:   emptyResult,
		},
		"EmptyFrom": {
			targetNSFrom: "",
			compatNSFrom: "",
			targetNSTo:   targetDB.Name() + ".newCollection",
			compatNSTo:   compatDB.Name() + ".newCollection",
			resultType:   emptyResult,
		},
		"EmptyTo": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo:   "",
			compatNSTo:   "",
			resultType:   emptyResult,
		},
		"EmptyDBFrom": {
			targetNSFrom: "." + targetCollection.Name(),
			compatNSFrom: "." + compatCollection.Name(),
			targetNSTo:   targetDB.Name() + ".newCollection",
			compatNSTo:   targetDB.Name() + ".newCollection",
			resultType:   emptyResult,
		},
		"EmptyDBTo": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo:   ".newCollection",
			compatNSTo:   ".newCollection",
			resultType:   emptyResult,
		},
		"EmptyCollectionFrom": {
			targetNSFrom: targetDB.Name() + ".",
			compatNSFrom: compatDB.Name() + ".",
			targetNSTo:   targetDB.Name() + ".newCollection",
			compatNSTo:   targetDB.Name() + ".newCollection",
			resultType:   emptyResult,
		},
		"EmptyCollectionTo": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo:   targetDB.Name() + ".",
			compatNSTo:   targetDB.Name() + ".",
			resultType:   emptyResult,
		},
		"NonExistentDB": {
			targetNSFrom: "nonExistentDB." + targetCollection.Name(),
			compatNSFrom: "nonExistentDB." + compatCollection.Name(),
			targetNSTo:   "nonExistentDB.newCollection",
			compatNSTo:   "nonExistentDB.newCollection",
			resultType:   emptyResult,
		},
		"NonExistentCollectionFrom": {
			targetNSFrom: targetDB.Name() + ".nonExistentCollection",
			compatNSFrom: compatDB.Name() + ".nonExistentCollection",
			targetNSTo:   targetDB.Name() + ".newCollection",
			compatNSTo:   targetDB.Name() + ".newCollection",
			resultType:   emptyResult,
		},
		"CollectionToAlreadyExists": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo:   targetDB.Name() + "." + targetCollectionExists.Name(),
			compatNSTo:   targetDB.Name() + "." + compatCollectionExists.Name(),
			resultType:   emptyResult,
		},
		"SameNamespace": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo:   targetDB.Name() + "." + targetCollection.Name(),
			compatNSTo:   targetDB.Name() + "." + compatCollection.Name(),
			resultType:   emptyResult,
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			var targetRes bson.D
			targetCommand := bson.D{{"renameCollection", tc.targetNSFrom}, {"to", tc.targetNSTo}}
			targetErr := targetDBConnect.RunCommand(ctx, targetCommand).Decode(&targetRes)

			var compatRes bson.D
			compatCommand := bson.D{{"renameCollection", tc.compatNSFrom}, {"to", tc.compatNSTo}}
			compatErr := compatDBConnect.RunCommand(ctx, compatCommand).Decode(&compatRes)

			if tc.resultType == emptyResult {
				require.Error(t, compatErr)

				targetErr = UnsetRaw(t, targetErr)
				compatErr = UnsetRaw(t, compatErr)

				//if tc.altMessage != "" {
				//	var expectedErr mongo.CommandError
				//	require.True(t, errors.As(compatErr, &expectedErr))
				//	AssertEqualAltError(t, expectedErr, tc.altMessage, targetErr)
				//} else {
				assert.Equal(t, compatErr, targetErr)
				//	}

				return
			}

			// Collection lists after rename must be the same
			targetNames, err := targetDB.ListCollectionNames(ctx, bson.D{})
			require.NoError(t, err)

			compatNames, err := compatDB.ListCollectionNames(ctx, bson.D{})
			require.NoError(t, err)

			assert.Equal(t, targetNames, compatNames)

			// Recreation of collection with the old name must be possible
			err = targetDB.CreateCollection(ctx, targetCollection.Name())
			require.NoError(t, err)

			err = compatDB.CreateCollection(ctx, compatCollection.Name())
			require.NoError(t, err)

			// Collection lists after recreation must be the same
			targetNames, err = targetDB.ListCollectionNames(ctx, bson.D{})
			require.NoError(t, err)

			compatNames, err = compatDB.ListCollectionNames(ctx, bson.D{})
			require.NoError(t, err)

			assert.Equal(t, targetNames, compatNames)
		})
	}
}
