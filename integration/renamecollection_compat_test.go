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
	dbName := targetDB.Name()
	cName := targetCollection.Name()

	require.Equal(t, compatDB.Name(), targetDB.Name())
	require.Equal(t, compatCollection.Name(), targetCollection.Name())

	targetDBConnect := targetCollection.Database().Client().Database("admin")
	compatDBConnect := compatCollection.Database().Client().Database("admin")

	nsFrom := dbName + "." + cName
	nsTo := dbName + ".newCollection"

	var targetRes bson.D
	targetCommand := bson.D{{"renameCollection", nsFrom}, {"to", nsTo}}
	targetErr := targetDBConnect.RunCommand(ctx, targetCommand).Decode(&targetRes)
	require.NoError(t, targetErr, "compat error; target returned no error")

	var compatRes bson.D
	compatCommand := bson.D{{"renameCollection", nsFrom}, {"to", nsTo}}
	compatErr := compatDBConnect.RunCommand(ctx, compatCommand).Decode(&compatRes)
	require.NoError(t, compatErr, "compat error; target returned no error")

	// Collection lists after rename must be the same
	targetNames, err := targetDB.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	compatNames, err := compatDB.ListCollectionNames(ctx, bson.D{})
	require.NoError(t, err)

	assert.ElementsMatch(t, targetNames, compatNames)

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

	assert.ElementsMatch(t, targetNames, compatNames)

	// Rename one more time
	targetCommand = bson.D{{"renameCollection", nsFrom}, {"to", nsTo + "_new"}}
	targetErr = targetDBConnect.RunCommand(ctx, targetCommand).Decode(&targetRes)
	require.NoError(t, targetErr)

	compatCommand = bson.D{{"renameCollection", nsFrom}, {"to", nsTo + "_new"}}
	compatErr = compatDBConnect.RunCommand(ctx, compatCommand).Decode(&compatRes)
	require.NoError(t, compatErr)

	// Rename back
	targetCommand = bson.D{{"renameCollection", nsTo + "_new"}, {"to", nsFrom}}
	targetErr = targetDBConnect.RunCommand(ctx, targetCommand).Decode(&targetRes)
	require.NoError(t, targetErr)

	compatCommand = bson.D{{"renameCollection", nsTo + "_new"}, {"to", nsFrom}}
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
	dbName := targetDB.Name()
	cName := targetCollection.Name()
	cExistingName := targetCollectionExists.Name()

	require.Equal(t, compatDB.Name(), targetDB.Name())
	require.Equal(t, compatCollection.Name(), targetCollection.Name())
	require.Equal(t, compatCollectionExists.Name(), targetCollectionExists.Name())

	targetDBConnect := targetCollection.Database().Client().Database("admin")
	compatDBConnect := compatCollection.Database().Client().Database("admin")

	for name, tc := range map[string]struct {
		nsFrom any
		nsTo   any
	}{
		"NilFrom": {
			nsFrom: nil,
			nsTo:   dbName + ".newCollection",
		},
		"NilTo": {
			nsFrom: dbName + "." + cName,
			nsTo:   nil,
		},
		"BadTypeFrom": {
			nsFrom: int32(42),
			nsTo:   dbName + ".newCollection",
		},
		"BadTypeTo": {
			nsFrom: dbName + "." + cName,
			nsTo:   int32(42),
		},
		"EmptyFrom": {
			nsFrom: "",
			nsTo:   dbName + ".newCollection",
		},
		"EmptyTo": {
			nsFrom: dbName + "." + cName,
			nsTo:   "",
		},
		"EmptyDBFrom": {
			nsFrom: "." + cName,
			nsTo:   dbName + ".newCollection",
		},
		"EmptyDBTo": {
			nsFrom: dbName + "." + cName,
			nsTo:   ".newCollection",
		},
		"EmptyCollectionFrom": {
			nsFrom: dbName + ".",
			nsTo:   dbName + ".newCollection",
		},
		"EmptyCollectionTo": {
			nsFrom: dbName + "." + cName,
			nsTo:   dbName + ".",
		},
		"NonExistentDB": {
			nsFrom: "nonExistentDB." + cName,
			nsTo:   "nonExistentDB.newCollection",
		},
		"NonExistentCollectionFrom": {
			nsFrom: dbName + ".nonExistentCollection",
			nsTo:   dbName + ".newCollection",
		},
		"CollectionToAlreadyExists": {
			nsFrom: dbName + "." + cName,
			nsTo:   dbName + "." + cExistingName,
		},
		"SameNamespace": {
			nsFrom: dbName + "." + cName,
			nsTo:   dbName + "." + cName,
		},
		"InvalidNameTo": {
			nsFrom: dbName + "." + cName,
			nsTo:   dbName + ".new$Collection",
		},
		"LongNameTo": {
			nsFrom: dbName + "." + cName,
			nsTo:   dbName + "." + strings.Repeat("aB", 150),
		},
	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			var targetRes bson.D
			targetCommand := bson.D{{"renameCollection", tc.nsFrom}, {"to", tc.nsTo}}
			targetErr := targetDBConnect.RunCommand(ctx, targetCommand).Decode(&targetRes)

			var compatRes bson.D
			compatCommand := bson.D{{"renameCollection", tc.nsFrom}, {"to", tc.nsTo}}
			compatErr := compatDBConnect.RunCommand(ctx, compatCommand).Decode(&compatRes)

			t.Logf("Target error: %v", targetErr)
			t.Logf("Compat error: %v", compatErr)

			// error messages are intentionally not compared
			AssertMatchesCommandError(t, compatErr, targetErr)
		})
	}
}
