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

func TestDropIndexesCommandErrors(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes are not supported for Tigris")

	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // for readability
		toCreate []mongo.IndexModel // optional, if set, create the given indexes before drop is called
		toDrop   any                // required, index to drop
		command  bson.D             // optional, if set it runs this command instead of dropping toDrop

		err        *mongo.CommandError // required, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"InvalidType": {
			toDrop: true,
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'dropIndexes.index' is the wrong type 'bool', expected types '[string, object']",
			},
			altMessage: `BSON field 'dropIndexes.index' is the wrong type 'bool', expected types '[string, object]'`,
		},
		"MultipleIndexesByKey": {
			toCreate: []mongo.IndexModel{
				{Keys: bson.D{{"v", -1}}},
				{Keys: bson.D{{"v.foo", -1}}},
			},
			toDrop: bson.A{bson.D{{"v", -1}}, bson.D{{"v.foo", -1}}},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'dropIndexes.index' is the wrong type 'array', expected types '[string']",
			},
			altMessage: `BSON field 'dropIndexes.index' is the wrong type 'array', expected types '[string, object]'`,
		},
		"NonExistentMultipleIndexes": {
			err: &mongo.CommandError{
				Code:    27,
				Name:    "IndexNotFound",
				Message: "index not found with name [non-existent]",
			},
			toDrop: bson.A{"non-existent", "invalid"},
		},
		"InvalidMultipleIndexType": {
			toDrop: bson.A{1},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'dropIndexes.index' is the wrong type 'array', expected types '[string']",
			},
			altMessage: `BSON field 'dropIndexes.index' is the wrong type 'array', expected types '[string, object]'`,
		},
		"InvalidDocumentIndex": {
			toDrop: bson.D{{"invalid", "invalid"}},
			err: &mongo.CommandError{
				Code:    27,
				Name:    "IndexNotFound",
				Message: "can't find index with key: { invalid: \"invalid\" }",
			},
		},
		"NonExistentKey": {
			toDrop: bson.D{{"non-existent", 1}},
			err: &mongo.CommandError{
				Code:    27,
				Name:    "IndexNotFound",
				Message: "can't find index with key: { non-existent: 1 }",
			},
		},
		"DocumentIndexID": {
			toDrop: bson.D{{"_id", 1}},
			err: &mongo.CommandError{
				Code:    72,
				Name:    "InvalidOptions",
				Message: "cannot drop _id index",
			},
		},
		"MissingIndexField": {
			command: bson.D{
				{"dropIndexes", "collection"},
			},
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: "BSON field 'dropIndexes.index' is missing but a required field",
			},
		},
		"NonExistentDescendingID": {
			toDrop: bson.D{{"_id", -1}},
			err: &mongo.CommandError{
				Code:    27,
				Name:    "IndexNotFound",
				Message: "can't find index with key: { _id: -1 }",
			},
		},
		"NonExistentMultipleKeyIndex": {
			toDrop: bson.D{
				{"non-existent1", -1},
				{"non-existent2", -1},
			},
			err: &mongo.CommandError{
				Code:    27,
				Name:    "IndexNotFound",
				Message: "can't find index with key: { non-existent1: -1, non-existent2: -1 }",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			if tc.command != nil {
				require.Nil(t, tc.toDrop, "toDrop must be nil when using command")
			} else {
				require.NotNil(t, tc.toDrop, "toDrop must not be nil")
			}

			require.NotNil(t, tc.err, "err must not be nil")

			s := setup.SetupWithOpts(t, &setup.SetupOpts{
				Providers: []shareddata.Provider{shareddata.Composites},
			})
			ctx, collection := s.Ctx, s.Collection

			if tc.toCreate != nil {
				_, err := collection.Indexes().CreateMany(ctx, tc.toCreate)
				require.NoError(t, err)
			}

			command := bson.D{
				{"dropIndexes", collection.Name()},
				{"index", tc.toDrop},
			}

			if tc.command != nil {
				command = tc.command
			}

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)

			assert.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestCreateIndexesCommandInvalidSpec(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes are not supported for Tigris")

	t.Parallel()

	for name, tc := range map[string]struct {
		indexes        any  // optional
		missingIndexes bool // optional, if set indexes must be nil

		err        *mongo.CommandError // required, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"EmptyIndexes": {
			indexes: bson.A{},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Must specify at least one index to create",
			},
		},
		"MissingIndexes": {
			missingIndexes: true,
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: "BSON field 'createIndexes.indexes' is missing but a required field",
			},
		},
		"NilIndexes": {
			indexes: nil,
			err: &mongo.CommandError{
				Code:    10065,
				Name:    "Location10065",
				Message: "invalid parameter: expected an object (indexes)",
			},
		},
		"InvalidTypeObject": {
			indexes: bson.D{},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'createIndexes.indexes' is the wrong type 'object', expected type 'array'",
			},
			altMessage: "BSON field 'createIndexes.indexes' is the wrong type, expected type 'array'",
		},
		"InvalidTypeInt": {
			indexes: 42,
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'createIndexes.indexes' is the wrong type 'int', expected type 'array'",
			},
			altMessage: "BSON field 'createIndexes.indexes' is the wrong type, expected type 'array'",
		},
		"InvalidTypeArrayString": {
			indexes: bson.A{"invalid"},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'createIndexes.indexes.0' is the wrong type 'string', expected type 'object'",
			},
			altMessage: "BSON field 'createIndexes.indexes.0' is the wrong type, expected type 'object'",
		},
		"IDIndex": {
			indexes: bson.A{
				bson.D{
					{"key", bson.D{{"_id", 1}}},
					{"name", "_id_"},
					{"unique", true},
				},
			},
			err: &mongo.CommandError{
				Code: 197,
				Name: "InvalidIndexSpecificationOption",
				Message: `The field 'unique' is not valid for an _id index specification.` +
					` Specification: { key: { _id: 1 }, name: "_id_", unique: true, v: 2 }`,
			},
		},
		"MissingName": {
			indexes: bson.A{
				bson.D{
					{"key", bson.D{{"v", 1}}},
				},
			},
			err: &mongo.CommandError{
				Code: 9,
				Name: "FailedToParse",
				Message: `Error in specification { key: { v: 1 } } :: caused by :: ` +
					`The 'name' field is a required property of an index specification`,
			},
		},
		"EmptyName": {
			indexes: bson.A{
				bson.D{
					{"key", bson.D{{"v", -1}}},
					{"name", ""},
				},
			},
			err: &mongo.CommandError{
				Code:    67,
				Name:    "CannotCreateIndex",
				Message: `Error in specification { key: { v: -1 }, name: "", v: 2 } :: caused by :: index name cannot be empty`,
			},
			altMessage: `Error in specification { key: { v: -1 }, name: "" } :: caused by :: index name cannot be empty`,
		},
		"MissingKey": {
			indexes: bson.A{
				bson.D{},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: `Error in specification {} :: caused by :: The 'key' field is a required property of an index specification`,
			},
		},
		"UniqueTypeDocument": {
			indexes: bson.A{
				bson.D{
					{"key", bson.D{{"v", 1}}},
					{"name", "unique_index"},
					{"unique", bson.D{}},
				},
			},
			err: &mongo.CommandError{
				Code: 14,
				Name: "TypeMismatch",
				Message: `Error in specification { key: { v: 1 }, name: "unique_index", unique: {} } ` +
					`:: caused by :: The field 'unique has value unique: {}, which is not convertible to bool`,
			},
			altMessage: `Error in specification { key: { v: 1 }, name: "unique_index", unique: {  } } ` +
				`:: caused by :: The field 'unique' has value unique: {  }, which is not convertible to bool`,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			if tc.missingIndexes {
				require.Nil(t, tc.indexes, "indexes must be nil if missingIndexes is true")
			}

			provider := shareddata.ArrayDocuments // one provider is enough to check for errors
			ctx, collection := setup.Setup(t, provider)

			var rest bson.D

			if !tc.missingIndexes {
				rest = append(rest, bson.E{Key: "indexes", Value: tc.indexes})
			}

			command := append(bson.D{
				{"createIndexes", collection.Name()},
			},
				rest...,
			)

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)

			require.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestCreateIndexesCommandInvalidCollection(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes are not supported for Tigris")

	t.Parallel()

	for name, tc := range map[string]struct {
		collectionName any
		indexes        any
		err            *mongo.CommandError
		altMessage     string
		skip           string
	}{
		"InvalidTypeCollection": {
			collectionName: 42,
			indexes: bson.A{
				bson.D{
					{"key", bson.D{{"v", 1}}},
					{"name", "v_1"},
				},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type int",
			},
			altMessage: "required parameter \"createIndexes\" has type int32 (expected string)",
		},
		"NilCollection": {
			collectionName: nil,
			indexes: bson.A{
				bson.D{
					{"key", bson.D{{"v", 1}}},
					{"name", "v_1"},
				},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type null",
			},
			altMessage: "required parameter \"createIndexes\" has type types.NullType (expected string)",
		},
		"EmptyCollection": {
			collectionName: "",
			indexes: bson.A{
				bson.D{
					{"key", bson.D{{"v", 1}}},
					{"name", "v_1"},
				},
			},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Invalid namespace specified 'TestCreateIndexesCommandInvalidCollection-EmptyCollection.'",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			provider := shareddata.ArrayDocuments // one provider is enough to check for errors
			ctx, collection := setup.Setup(t, provider)

			command := bson.D{
				{"createIndexes", tc.collectionName},
				{"indexes", tc.indexes},
			}

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)

			require.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestDropIndexesCommandInvalidCollection(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes are not supported for Tigris")

	t.Parallel()

	for name, tc := range map[string]struct {
		collectionName any
		indexName      any
		err            *mongo.CommandError
		altMessage     string
		skip           string
	}{
		"NonExistentCollection": {
			collectionName: "non-existent",
			indexName:      "index",
			err: &mongo.CommandError{
				Code:    26,
				Name:    "NamespaceNotFound",
				Message: "ns not found TestDropIndexesCommandInvalidCollection-NonExistentCollection.non-existent",
			},
		},
		"InvalidTypeCollection": {
			collectionName: 42,
			indexName:      "index",
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type int",
			},
			altMessage: "required parameter \"dropIndexes\" has type int32 (expected string)",
		},
		"NilCollection": {
			collectionName: nil,
			indexName:      "index",
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type null",
			},
			altMessage: "required parameter \"dropIndexes\" has type types.NullType (expected string)",
		},
		"EmptyCollection": {
			collectionName: "",
			indexName:      "index",
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Invalid namespace specified 'TestDropIndexesCommandInvalidCollection-EmptyCollection.'",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			provider := shareddata.ArrayDocuments // one provider is enough to check for errors
			ctx, collection := setup.Setup(t, provider)

			command := bson.D{
				{"dropIndexes", tc.collectionName},
				{"index", tc.indexName},
			}

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)

			require.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}
