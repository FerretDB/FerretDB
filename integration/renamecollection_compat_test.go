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
	"go.mongodb.org/mongo-driver/mongo"

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
	targetCollectionExists, compatCollectionExists := s.TargetCollections[1], s.CompatCollections[1]

	targetDB := targetCollection.Database()
	compatDB := compatCollection.Database()

	for name, tc := range map[string]struct {
		targetNSFrom any
		compatNSFrom any
		targetNSTo any
		compatNSTo any
		resultType compatTestCaseResultType
	} {
		"Valid": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo: targetDB.Name() + ".newCollection",
			compatNSTo: targetDB.Name() + ".newCollection",
		},
		"NilFrom": {
			targetNSFrom: nil,
			compatNSFrom: nil,
			targetNSTo: targetDB.Name() + ".newCollection",
			compatNSTo: compatDB.Name() + ".newCollection",
			resultType: emptyResult,
		},
		"NilTo": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo: nil,
			compatNSTo: nil,
			resultType: emptyResult,
		},
		"EmptyFrom": {
			targetNSFrom: "",
			compatNSFrom: "",
			targetNSTo: targetDB.Name() + ".newCollection",
			compatNSTo: compatDB.Name() + ".newCollection",
			resultType: emptyResult,
		},
		"EmptyTo": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo: "",
			compatNSTo: "",
			resultType: emptyResult,
		},
		"EmptyDBFrom": {

		},
		"EmptyDBTo": {

		},
		"EmptyCollectionFrom": {

		},
		"EmptyCollectionTo": {

		},
		"NonExistentDB": {

		},
		"NonExistentCollectionFrom": {

		},
		"CollectionToAlreadyExists": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo: targetDB.Name() + "." + targetCollectionExists.Name(),
			compatNSTo: targetDB.Name() + "." + compatCollectionExists.Name(),
			resultType: emptyResult,
		},
		"SameNamespace": {
			targetNSFrom: targetDB.Name() + "." + targetCollection.Name(),
			compatNSFrom: compatDB.Name() + "." + compatCollection.Name(),
			targetNSTo: targetDB.Name() + "." + targetCollection.Name(),
			compatNSTo: targetDB.Name() + "." + compatCollection.Name(),
			resultType: emptyResult,
		},

	} {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Helper()

			t.Parallel()

			var targetRes bson.D
			targetCommand := bson.D{{"renameCollection",tc.targetNSFrom}, {"to", tc.targetNSTo}}
			targetErr := targetDB.RunCommand(ctx, targetCommand).Decode(&targetRes)

			var compatRes bson.D
			compatCommand := bson.D{{"renameCollection",tc.compatNSFrom}, {"to", tc.compatNSTo}}
			compatErr := compatDB.RunCommand(ctx, compatCommand).Decode(&compatRes)

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
		}
	}
}

/*
func TestRenameCollectionCompat(t *testing.T) {
	t.Parallel()

	// rename collection must be executed from admin database.
	adminDB := setup.SetupWithOpts(t, &setup.SetupOpts{DatabaseName: "admin"})

	maxCollectionNameLen := strings.Repeat("a", 235)
	insertCollections := []string{"foo", "buz"}

	for name, tc := range map[string]struct { //nolint:vet // for readability
		sourceCollection string
		targetCollection string
		to               string
		expected         bson.D
		err              *mongo.CommandError
		recreateOld      bool // this indicates that the old collection should be recreated
		targetNamespace  any  // optional, if set, ignores targetCollection and uses targetNamespace
		emptyNamespace   any  // optional, if set, allows to create an empty namespace
	}{
		"Rename": {
			sourceCollection: "foo",
			targetCollection: "bar",
			to:               "to",
			expected:         bson.D{{"ok", float64(1)}},
		},
		"RenameSame": {
			sourceCollection: "foo",
			targetCollection: "foo",
			to:               "to",
			err: &mongo.CommandError{
				Code:    20,
				Name:    "IllegalOperation",
				Message: `Can't rename a collection to itself`,
			},
		},
		"TargetAlreadyExists": {
			sourceCollection: "foo",
			targetCollection: "buz",
			to:               "to",
			err: &mongo.CommandError{
				Code:    48,
				Name:    "NamespaceExists",
				Message: `target namespace exists`,
			},
		},
		"RenameDuplicate": {
			sourceCollection: "foo",
			targetCollection: "buz",
			to:               "to",
			err: &mongo.CommandError{
				Code:    48,
				Name:    "NamespaceExists",
				Message: `target namespace exists`,
			},
		},
		"SourceDoesNotExist": {
			sourceCollection: "none",
			targetCollection: "bar",
			to:               "to",
			err: &mongo.CommandError{
				Code: 26,
				Name: "NamespaceNotFound",
				Message: "Source collection TestCommandsAdministrationRenameCollection-SourceDoesNotExist.none " +
					"does not exist",
			},
		},
		// this confirms that after we rename foo to bar and then recreate foo again,
		// 1. the newly inserted documents exist
		// 2. bool-false doesn't exist
		"InsertIntoOld": {
			sourceCollection: "foo",
			targetCollection: "bar",
			to:               "to",
			expected:         bson.D{{"ok", float64(1)}},
			recreateOld:      true,
		},
		"MaxCollectionName": {
			sourceCollection: "foo",
			targetCollection: maxCollectionNameLen + "a", // 236 chars
			to:               "to",
			err: &mongo.CommandError{
				Code: 73,
				Name: "InvalidNamespace",
				Message: "error with target namespace: Fully qualified namespace is too long. " +
					"Namespace: TestCommandsAdministrationRenameCollection-MaxCollectionName.aaaaa" +
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa Max: 235",
			},
		},
		"EmptyNamespace": {
			sourceCollection: "",
			targetCollection: "bar",
			to:               "to",
			emptyNamespace:   true,
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Invalid namespace specified ",
			},
		},
		"BadParamTo": {
			sourceCollection: "foo",
			targetCollection: "",
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: "BSON field 'renameCollection.to' is missing but a required field",
			},
		},
		"BadParamTypeTo": {
			sourceCollection: "foo",
			targetNamespace:  true,
			to:               "to",
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'renameCollection.to' is the wrong type 'bool', expected type 'string'",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx, collection := setup.Setup(t, shareddata.Bools)

			db := collection.Database()

			require.NotEmpty(t, insertCollections)
			for _, coll := range insertCollections {
				require.NoError(t, db.CreateCollection(ctx, coll))
			}

			var actual bson.D

			var sourceNamespace any
			sourceNamespace = fmt.Sprintf("%s.%s", db.Name(), tc.sourceCollection)

			if tc.emptyNamespace != nil {
				sourceNamespace = tc.sourceCollection
			}

			var targetNamespace any
			targetNamespace = fmt.Sprintf("%s.%s", db.Name(), tc.targetCollection)

			if tc.targetNamespace != nil {
				targetNamespace = tc.targetNamespace
			}

			cmd := bson.D{
				{"renameCollection", sourceNamespace},
			}

			if tc.to != "" {
				cmd = append(cmd, bson.E{Key: tc.to, Value: targetNamespace})
			}

			err := adminDB.Collection.Database().RunCommand(
				ctx, cmd,
			).Decode(&actual)

			if tc.err != nil {
				AssertEqualError(t, *tc.err, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)

			collections, err := db.ListCollectionNames(
				ctx,
				bson.D{},
				nil,
			)

			require.NoError(t, err)

			require.Contains(t, collections, tc.targetCollection)
			require.NotContains(t, collections, tc.sourceCollection)

			if tc.recreateOld {
				require.Equal(t, "bar", tc.targetCollection)

				// we effectively recreate foo, the old collection.
				require.NoError(t, db.CreateCollection(ctx, tc.sourceCollection))

				res, err := db.Collection(tc.sourceCollection).InsertMany(ctx, []any{
					bson.D{{"_id", 1}},
					bson.D{{"_id", 2}},
				})

				require.NoError(t, err)

				expected := mongo.InsertManyResult{InsertedIDs: []any{int32(1), int32(2)}}
				require.Equal(t, &expected, res)

				var v any
				err = db.Collection(tc.targetCollection).FindOne(
					ctx, bson.D{{"bool-false", false}},
				).Decode(&v)

				require.ErrorIs(t, err, mongo.ErrNoDocuments)
			}
		})
	}
}
*/
