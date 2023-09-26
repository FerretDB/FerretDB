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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestRenameCollectionCompat(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                []shareddata.Provider{shareddata.DocumentsDocuments, shareddata.Bools},
		AddNonExistentCollection: true,
	})

	ctx, targetCollection, compatCollection := s.Ctx, s.TargetCollections[0], s.CompatCollections[0]

	targetDB := targetCollection.Database()
	compatDB := compatCollection.Database()

	require.Equal(t, compatDB.Name(), targetDB.Name())
	dbName := targetDB.Name()

	require.Equal(t, compatCollection.Name(), targetCollection.Name())
	cName := targetCollection.Name()

	to := dbName + ".newCollection"
	from := dbName + "." + cName

	targetDBConnect := targetCollection.Database().Client().Database("admin")
	compatDBConnect := compatCollection.Database().Client().Database("admin")

	var targetRes bson.D
	targetCommand := bson.D{{"renameCollection", from}, {"to", to}}
	targetErr := targetDBConnect.RunCommand(ctx, targetCommand).Decode(&targetRes)
	require.NoError(t, targetErr, "compat error; target returned no error")

	var compatRes bson.D
	compatCommand := bson.D{{"renameCollection", from}, {"to", to}}
	compatErr := compatDBConnect.RunCommand(ctx, compatCommand).Decode(&compatRes)
	require.NoError(t, compatErr, "compat error; target returned no error")

	// Collection lists after rename must be the same
	targetNames, err := targetDB.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	compatNames, err := compatDB.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	assert.ElementsMatch(t, targetNames, compatNames)

	// Recreation of collection with the old name must be possible
	err = targetDB.CreateCollection(ctx, cName)
	require.NoError(t, err)

	err = compatDB.CreateCollection(ctx, cName)
	require.NoError(t, err)

	// Collection lists after recreation must be the same
	targetNames, err = targetDB.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	compatNames, err = compatDB.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	assert.ElementsMatch(t, targetNames, compatNames)

	// Rename one more time
	targetCommand = bson.D{{"renameCollection", from}, {"to", to + "_new"}}
	targetErr = targetDBConnect.RunCommand(ctx, targetCommand).Decode(&targetRes)
	require.NoError(t, targetErr)

	compatCommand = bson.D{{"renameCollection", from}, {"to", to + "_new"}}
	compatErr = compatDBConnect.RunCommand(ctx, compatCommand).Decode(&compatRes)
	require.NoError(t, compatErr)

	// Rename back
	targetCommand = bson.D{{"renameCollection", to + "_new"}, {"to", from}}
	targetErr = targetDBConnect.RunCommand(ctx, targetCommand).Decode(&targetRes)
	require.NoError(t, targetErr)

	compatCommand = bson.D{{"renameCollection", to + "_new"}, {"to", from}}
	compatErr = compatDBConnect.RunCommand(ctx, compatCommand).Decode(&compatRes)
	require.NoError(t, compatErr)

	// Collection lists after all the manipulations must be the same
	targetNames, err = targetDB.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	compatNames, err = compatDB.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	assert.ElementsMatch(t, targetNames, compatNames)
}

func TestRenameCollectionCompatErrors(t *testing.T) {
	t.Parallel()

	s := setup.SetupCompatWithOpts(t, &setup.SetupCompatOpts{
		Providers:                []shareddata.Provider{shareddata.DocumentsDocuments, shareddata.Bools},
		AddNonExistentCollection: true,
	})

	ctx, targetCollection, compatCollection := s.Ctx, s.TargetCollections[0], s.CompatCollections[0]
	targetCollectionExists, compatCollectionExists := s.TargetCollections[1], s.CompatCollections[1]

	targetDB := targetCollection.Database()
	compatDB := compatCollection.Database()

	require.Equal(t, compatDB.Name(), targetDB.Name())
	dbName := targetDB.Name()

	require.Equal(t, compatCollection.Name(), targetCollection.Name())
	cName := targetCollection.Name()

	require.Equal(t, compatCollectionExists.Name(), targetCollectionExists.Name())
	cExistingName := targetCollectionExists.Name()

	targetDBConnect := targetCollection.Database().Client().Database("admin")
	compatDBConnect := compatCollection.Database().Client().Database("admin")

	for name, tc := range map[string]struct {
		from any
		to   any
	}{
		"NilFrom": {
			from: nil,
			to:   dbName + ".newCollection",
		},
		"NilTo": {
			from: dbName + "." + cName,
			to:   nil,
		},
		"BadTypeFrom": {
			from: int32(42),
			to:   dbName + ".newCollection",
		},
		"BadTypeTo": {
			from: dbName + "." + cName,
			to:   int32(42),
		},
		"EmptyFrom": {
			from: "",
			to:   dbName + ".newCollection",
		},
		"EmptyTo": {
			from: dbName + "." + cName,
			to:   "",
		},
		"EmptyDBFrom": {
			from: "." + cName,
			to:   dbName + ".newCollection",
		},
		"EmptyDBTo": {
			from: dbName + "." + cName,
			to:   ".newCollection",
		},
		"EmptyCollectionFrom": {
			from: dbName + ".",
			to:   dbName + ".newCollection",
		},
		"EmptyCollectionTo": {
			from: dbName + "." + cName,
			to:   dbName + ".",
		},
		"NonExistentDB": {
			from: "nonExistentDB." + cName,
			to:   "nonExistentDB.newCollection",
		},
		"NonExistentCollectionFrom": {
			from: dbName + ".nonExistentCollection",
			to:   dbName + ".newCollection",
		},
		"CollectionToAlreadyExists": {
			from: dbName + "." + cName,
			to:   dbName + "." + cExistingName,
		},
		"SameNamespace": {
			from: dbName + "." + cName,
			to:   dbName + "." + cName,
		},
		"InvalidNameTo": {
			from: dbName + "." + cName,
			to:   dbName + ".new$Collection",
		},
		"LongNameTo": {
			from: dbName + "." + cName,
			to:   dbName + "." + strings.Repeat("aB", 150),
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			var targetRes bson.D
			targetCommand := bson.D{{"renameCollection", tc.from}, {"to", tc.to}}
			targetErr := targetDBConnect.RunCommand(ctx, targetCommand).Decode(&targetRes)

			var compatRes bson.D
			compatCommand := bson.D{{"renameCollection", tc.from}, {"to", tc.to}}
			compatErr := compatDBConnect.RunCommand(ctx, compatCommand).Decode(&compatRes)

			t.Logf("Target error: %v", targetErr)
			t.Logf("Compat error: %v", compatErr)

			// error messages are intentionally not compared
			AssertMatchesCommandError(t, compatErr, targetErr)
		})
	}
}
