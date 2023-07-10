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
)

func TestFindAndModifyEmptyCollectionName(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"EmptyCollectionName": {
			err: &mongo.CommandError{
				Code:    73,
				Message: "Invalid namespace specified 'TestFindAndModifyEmptyCollectionName-EmptyCollectionName.'",
				Name:    "InvalidNamespace",
			},
			altMessage: "Invalid namespace specified 'TestFindAndModifyEmptyCollectionName-EmptyCollectionName.'",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.err, "err must not be nil")

			ctx, collection := setup.Setup(t, shareddata.Doubles)

			var res bson.D
			err := collection.Database().RunCommand(ctx, bson.D{{"findAndModify", ""}}).Decode(&res)

			assert.Nil(t, res)
			AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
		})
	}
}

func TestFindAndModifyCommandErrors(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // it is used for test only
		command  bson.D              // required, command to run
		provider shareddata.Provider // optional, default uses shareddata.ArrayDocuments

		err        *mongo.CommandError // required
		altMessage string              // optional, alternative error message
		skip       string              // optional, skip test with a specified reason
	}{
		"UpsertAndRemove": {
			command: bson.D{
				{"upsert", true},
				{"remove", true},
			},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "Cannot specify both upsert=true and remove=true ",
			},
			altMessage: "Cannot specify both upsert=true and remove=true",
		},
		"BadSortType": {
			command: bson.D{
				{"update", bson.D{}},
				{"sort", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'findAndModify.sort' is the wrong type 'string', expected type 'object'",
			},
			altMessage: "BSON field 'findAndModify.sort' is the wrong type 'string', expected type 'object'",
		},
		"BadRemoveType": {
			command: bson.D{
				{"query", bson.D{}},
				{"remove", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'findAndModify.remove' is the wrong type 'string', expected types '[bool, long, int, decimal, double']",
			},
			altMessage: "BSON field 'findAndModify.remove' is the wrong type 'string', expected type 'bool'",
		},
		"BadNewType": {
			command: bson.D{
				{"query", bson.D{}},
				{"new", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'findAndModify.new' is the wrong type 'string', expected types '[bool, long, int, decimal, double']",
			},
			altMessage: "BSON field 'new' is the wrong type 'string', expected types '[bool, long, int, decimal, double]'",
		},
		"BadUpsertType": {
			command: bson.D{
				{"query", bson.D{}},
				{"upsert", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'findAndModify.upsert' is the wrong type 'string', expected types '[bool, long, int, decimal, double']",
			},
			altMessage: "BSON field 'findAndModify.upsert' is the wrong type 'string', expected type 'bool'",
		},
		"SetUnsuitableValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$set", bson.D{{"v.foo", "foo"}}}}},
			},
			err: &mongo.CommandError{
				Code: 28,
				Name: "PathNotViable",
				Message: "Plan executor error during findAndModify :: caused by :: Cannot create field 'foo' " +
					"in element {v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
			},
			altMessage: "Cannot create field 'foo' in element " +
				"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
		},
		"SetImmutableID": {
			command: bson.D{
				{"update", bson.D{{"$set", bson.D{{"_id", "non-existent"}}}}},
			},
			err: &mongo.CommandError{
				Code: 66,
				Name: "ImmutableField",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"Performing an update on the path '_id' would modify the immutable field '_id'",
			},
			altMessage: "Performing an update on the path '_id' would modify the immutable field '_id'",
		},
		"RenameEmptyFieldName": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$rename", bson.D{{"", "v"}}}}},
			},
			err: &mongo.CommandError{
				Code:    56,
				Name:    "EmptyFieldName",
				Message: "An empty update path is not valid.",
			},
		},
		"RenameEmptyPath": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$rename", bson.D{{"v.", "v"}}}}},
			},
			err: &mongo.CommandError{
				Code:    56,
				Name:    "EmptyFieldName",
				Message: "The update path 'v.' contains an empty field name, which is not allowed.",
			},
		},
		"RenameArrayInvalidIndex": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$rename", bson.D{{"v.-1", "f"}}}}},
			},
			err: &mongo.CommandError{
				Code: 28,
				Name: "PathNotViable",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"cannot use the part (v of v.-1) to traverse the element " +
					"({v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]})",
			},
			altMessage: "cannot use path 'v.-1' to traverse the document",
		},
		"RenameUnsuitableValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$rename", bson.D{{"v.0.foo.0.bar.z", "f"}}}}},
			},
			err: &mongo.CommandError{
				Code: 28,
				Name: "PathNotViable",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"cannot use the part (bar of v.0.foo.0.bar.z) to traverse the element ({bar: \"hello\"})",
			},
			altMessage: "types.getByPath: can't access string by path \"z\"",
		},
		"IncTypeMismatch": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$inc", bson.D{{"v", "string"}}}}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Cannot increment with non-numeric argument: {v: \"string\"}",
			},
		},
		"IncUnsuitableValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$inc", bson.D{{"v.foo", 1}}}}},
			},
			err: &mongo.CommandError{
				Code: 28,
				Name: "PathNotViable",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"Cannot create field 'foo' in element " +
					"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
			},
			altMessage: "Cannot create field 'foo' in element " +
				"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
		},
		"IncNonNumeric": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$inc", bson.D{{"v.0.foo.0.bar", 1}}}}},
			},
			err: &mongo.CommandError{
				Code: 14,
				Name: "TypeMismatch",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"Cannot apply $inc to a value of non-numeric type. " +
					"{_id: \"array-documents-nested\"} has the field 'bar' of non-numeric type string",
			},
			altMessage: "Cannot apply $inc to a value of non-numeric type. " +
				"{_id: \"array-documents-nested\"} has the field 'bar' of non-numeric type string",
		},
		"IncInt64BadValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64-max"}}},
				{"update", bson.D{{"$inc", bson.D{{"v", math.MaxInt64}}}}},
			},
			err: &mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"Failed to apply $inc operations to current value " +
					"((NumberLong)9223372036854775807) for document {_id: \"int64-max\"}",
			},
			provider: shareddata.Int64s,
			altMessage: "Failed to apply $inc operations to current value " +
				"((NumberLong)9223372036854775807) for document {_id: \"int64-max\"}",
		},
		"IncInt32BadValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"$inc", bson.D{{"v", math.MaxInt64}}}}},
			},
			err: &mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"Failed to apply $inc operations to current value " +
					"((NumberInt)42) for document {_id: \"int32\"}",
			},
			provider: shareddata.Int32s,
			altMessage: "Failed to apply $inc operations to current value " +
				"((NumberInt)42) for document {_id: \"int32\"}",
		},
		"MaxUnsuitableValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$max", bson.D{{"v.foo", 1}}}}},
			},
			err: &mongo.CommandError{
				Code: 28,
				Name: "PathNotViable",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"Cannot create field 'foo' in element " +
					"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
			},
			altMessage: "Cannot create field 'foo' in element " +
				"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
		},
		"MinUnsuitableValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$min", bson.D{{"v.foo", 1}}}}},
			},
			err: &mongo.CommandError{
				Code: 28,
				Name: "PathNotViable",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"Cannot create field 'foo' in element " +
					"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
			},
			altMessage: "Cannot create field 'foo' in element " +
				"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
		},
		"MulTypeMismatch": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$mul", bson.D{{"v", "string"}}}}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Cannot multiply with non-numeric argument: {v: \"string\"}",
			},
		},
		"MulTypeMismatchNonExistent": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$mul", bson.D{{"non-existent", "string"}}}}},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Cannot multiply with non-numeric argument: {non-existent: \"string\"}",
			},
		},
		"MulUnsuitableValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$mul", bson.D{{"v.foo", 1}}}}},
			},
			err: &mongo.CommandError{
				Code: 28,
				Name: "PathNotViable",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"Cannot create field 'foo' in element " +
					"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
			},
			altMessage: "Cannot create field 'foo' in element " +
				"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
		},
		"MulNonNumeric": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$mul", bson.D{{"v.0.foo.0.bar", 1}}}}},
			},
			err: &mongo.CommandError{
				Code: 14,
				Name: "TypeMismatch",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"Cannot apply $mul to a value of non-numeric type. " +
					"{_id: \"array-documents-nested\"} has the field 'bar' of non-numeric type string",
			},
			altMessage: "Cannot apply $mul to a value of non-numeric type. " +
				"{_id: \"array-documents-nested\"} has the field 'bar' of non-numeric type string",
		},
		"MulInt64BadValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "int64-max"}}},
				{"update", bson.D{{"$mul", bson.D{{"v", math.MaxInt64}}}}},
			},
			err: &mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: "Failed to apply $mul operations to current value " +
					"((NumberLong)9223372036854775807) for document {_id: \"int64-max\"}",
			},
			provider: shareddata.Int64s,
			altMessage: "Plan executor error during findAndModify :: caused by :: " +
				"Failed to apply $mul operations to current value " +
				"((NumberLong)9223372036854775807) for document {_id: \"int64-max\"}",
		},
		"MulInt32BadValue": {
			command: bson.D{
				{"query", bson.D{{"_id", "int32"}}},
				{"update", bson.D{{"$mul", bson.D{{"v", math.MaxInt64}}}}},
			},
			err: &mongo.CommandError{
				Code: 2,
				Name: "BadValue",
				Message: "Plan executor error during findAndModify :: caused by :: " +
					"Failed to apply $mul operations to current value " +
					"((NumberInt)42) for document {_id: \"int32\"}",
			},
			provider: shareddata.Int32s,
			altMessage: "Failed to apply $mul operations to current value " +
				"((NumberInt)42) for document {_id: \"int32\"}",
		},
		"MulEmptyPath": {
			command: bson.D{
				{"query", bson.D{{"_id", "array-documents-nested"}}},
				{"update", bson.D{{"$mul", bson.D{{"v.", "v"}}}}},
			},
			err: &mongo.CommandError{
				Code:    56,
				Name:    "EmptyFieldName",
				Message: "The update path 'v.' contains an empty field name, which is not allowed.",
			},
		},
		"ConflictCollision": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"update", bson.D{
					{"$set", bson.D{{"v", "val"}}},
					{"$min", bson.D{{"v.foo", "val"}}},
				}},
			},
			err: &mongo.CommandError{
				Code:    40,
				Name:    "ConflictingUpdateOperators",
				Message: "Updating the path 'v.foo' would create a conflict at 'v'",
			},
			altMessage: "Updating the path 'v' would create a conflict at 'v'",
		},
		"ConflictOverwrite": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"update", bson.D{
					{"$set", bson.D{{"v.foo", "val"}}},
					{"$min", bson.D{{"v", "val"}}},
				}},
			},
			err: &mongo.CommandError{
				Code:    40,
				Name:    "ConflictingUpdateOperators",
				Message: "Updating the path 'v' would create a conflict at 'v'",
			},
			altMessage: "Updating the path 'v.foo' would create a conflict at 'v.foo'",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.command, "command must not be nil")
			require.NotNil(t, tc.err, "err must not be nil")

			provider := tc.provider
			if provider == nil {
				provider = shareddata.ArrayDocuments
			}

			ctx, collection := setup.Setup(t, provider)

			command := bson.D{{"findAndModify", collection.Name()}}
			command = append(command, tc.command...)
			if command.Map()["sort"] == nil {
				command = append(command, bson.D{{"sort", bson.D{{"_id", 1}}}}...)
			}

			var actual bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&actual)
			if tc.altMessage != "" {
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			AssertEqualCommandError(t, *tc.err, err)
		})
	}
}

func TestFindAndModifyCommandUpsert(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		command         bson.D // required, command to run
		lastErrorObject bson.D
	}{
		"UpsertNoSuchDocumentNoIdInQuery": {
			command: bson.D{
				{"query", bson.D{{
					"$and",
					bson.A{
						bson.D{{"v", bson.D{{"$gt", 0}}}},
						bson.D{{"v", bson.D{{"$lt", 0}}}},
					},
				}}},
				{"update", bson.D{{"$set", bson.D{{"v", 43.13}}}}},
				{"upsert", true},
			},
			lastErrorObject: bson.D{
				{"n", int32(1)},
				{"updatedExisting", false},
			},
		},
		"UpsertExpressionKey": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"upsert", true},
				{"update", bson.D{{"v", "replaced"}}},
			},
			lastErrorObject: bson.D{
				{"n", int32(1)},
				{"updatedExisting", false},
			},
		},
		"UpsertDocumentKey": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"key", "val"}}}}},
				{"upsert", true},
				{"update", bson.D{{"v", "replaced"}}},
			},
			lastErrorObject: bson.D{
				{"n", int32(1)},
				{"updatedExisting", false},
			},
		},
		"ExistsFalse": {
			command: bson.D{
				{"query", bson.D{{"_id", bson.D{{"$exists", false}}}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"v", "foo"}}}}},
			},
			lastErrorObject: bson.D{
				{"n", int32(1)},
				{"updatedExisting", false},
			},
		},
		"NonExistentExistsTrue": {
			command: bson.D{
				{"query", bson.D{{"non-existent", bson.D{{"$exists", true}}}}},
				{"upsert", true},
				{"update", bson.D{{"$set", bson.D{{"v", "foo"}}}}},
			},
			lastErrorObject: bson.D{
				{"n", int32(1)},
				{"updatedExisting", false},
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.NotNil(t, tc.command, "command must not be nil")

			ctx, collection := setup.Setup(t, shareddata.Doubles)

			command := append(bson.D{{"findAndModify", collection.Name()}}, tc.command...)

			var actual bson.D
			err := collection.Database().RunCommand(ctx, command).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()
			assert.Equal(t, float64(1), m["ok"])

			leb, ok := m["lastErrorObject"].(bson.D)
			if !ok {
				t.Fatal(actual)
			}

			// TODO: add document comparison here. Skip _id check as it always would different.
			for _, v := range leb {
				if v.Key == "upserted" {
					continue
				}
				assert.Equal(t, tc.lastErrorObject.Map()[v.Key], v.Value)
			}
		})
	}
}

func TestFindAndModifyNonExistingCollection(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t)

	var actual bson.D
	err := collection.FindOneAndUpdate(
		ctx, bson.D{}, bson.D{{"$set", bson.E{"foo", "bar"}}},
	).Decode(&actual)

	assert.Equal(t, mongo.ErrNoDocuments, err)
	assert.Nil(t, actual)
}
