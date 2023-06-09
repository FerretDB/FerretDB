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

func TestIndexesDropCommandErrors(t *testing.T) {
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
			skip:   "https://github.com/FerretDB/FerretDB/issues/2311",
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

func TestCreateIndexesInvalidSpec(t *testing.T) {
	setup.SkipForTigrisWithReason(t, "Indexes are not supported for Tigris")

	t.Parallel()

	for name, tc := range map[string]struct {
		indexes    any
		err        *mongo.CommandError
		altMessage string
		skip       string
	}{
		"EmptyIndexes": {
			indexes: bson.A{},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Must specify at least one index to create",
			},
		},
		"NilIndexes": {
			indexes: nil,
			err: &mongo.CommandError{
				Code:    10065,
				Name:    "Location10065",
				Message: "invalid parameter: expected an object (indexes)",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2311",
		},
		"InvalidType": {
			indexes: 42,
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'createIndexes.indexes' is the wrong type 'int', expected type 'array'",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2311",
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
				{"createIndexes", collection.Name()},
				{"indexes", tc.indexes},
			}

			var res bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&res)

			require.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestDropIndexesInvalidCollection(t *testing.T) {
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
				Message: "ns not found TestIndexesDropInvalidCollection-NonExistentCollection.non-existent",
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
			skip: "https://github.com/FerretDB/FerretDB/issues/2311",
		},
		"NilCollection": {
			collectionName: nil,
			indexName:      "index",
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type null",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2311",
		},
		"EmptyCollection": {
			collectionName: "",
			indexName:      "index",
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Invalid namespace specified 'TestIndexesDropInvalidCollection-EmptyCollection.'",
			},
			skip: "https://github.com/FerretDB/FerretDB/issues/2311",
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
