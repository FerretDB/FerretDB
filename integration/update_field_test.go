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

func TestUpdateFieldSet(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		id     string // optional, defaults to empty
		update bson.D // required, used for update parameter

		res     *mongo.UpdateResult // optional, expected response from update
		findRes bson.D              // optional, expected response from find
		skip    string              // optional, skip test with a specified reason
	}{
		"ArrayNil": {
			id:      "string",
			update:  bson.D{{"$set", bson.D{{"v", bson.A{nil}}}}},
			findRes: bson.D{{"_id", "string"}, {"v", bson.A{nil}}},
			res: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 1,
				UpsertedCount: 0,
			},
		},
		"SetSameValueInt": {
			id:      "int32",
			update:  bson.D{{"$set", bson.D{{"v", int32(42)}}}},
			findRes: bson.D{{"_id", "int32"}, {"v", int32(42)}},
			res: &mongo.UpdateResult{
				MatchedCount:  1,
				ModifiedCount: 0,
				UpsertedCount: 0,
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.update, "update should be set")

			ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

			res, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)

			require.NoError(t, err)
			require.Equal(t, tc.res, res)

			var actual bson.D
			err = collection.FindOne(ctx, bson.D{{"_id", tc.id}}).Decode(&actual)
			require.NoError(t, err)
			AssertEqualDocuments(t, tc.findRes, actual)
		})
	}
}

func TestUpdateFieldErrors(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct { //nolint:vet // it is used for test only
		id       string              // optional, defaults to empty
		update   bson.D              // required, used for update parameter
		provider shareddata.Provider // optional, default uses shareddata.ArrayDocuments

		err        *mongo.WriteError // required, expected error from MongoDB
		altMessage string            // optional, alternative error message for FerretDB, ignored if empty
		skip       string            // optional, skip test with a specified reason
	}{
		"SetUnsuitableValue": {
			id:     "array-documents-nested",
			update: bson.D{{"$rename", bson.D{{"v.foo", "foo"}}}},
			err: &mongo.WriteError{
				Code: 28,
				Message: "cannot use the part (v of v.foo) to traverse the element " +
					"({v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]})",
			},
			altMessage: "cannot use path 'v.foo' to traverse the document",
		},
		"SetImmutableID": {
			id:     "array-documents-nested",
			update: bson.D{{"$set", bson.D{{"_id", "another-id"}}}},
			err: &mongo.WriteError{
				Code:    66,
				Message: "Performing an update on the path '_id' would modify the immutable field '_id'",
			},
		},
		"RenameEmptyFieldName": {
			id:     "array-documents-nested",
			update: bson.D{{"$rename", bson.D{{"", "v"}}}},
			err: &mongo.WriteError{
				Code:    56,
				Message: "An empty update path is not valid.",
			},
		},
		"RenameEmptyPath": {
			id:     "array-documents-nested",
			update: bson.D{{"$rename", bson.D{{"v.", "v"}}}},
			err: &mongo.WriteError{
				Code:    56,
				Message: "The update path 'v.' contains an empty field name, which is not allowed.",
			},
		},
		"RenameArrayInvalidIndex": {
			id:     "array-documents-nested",
			update: bson.D{{"$rename", bson.D{{"v.-1", "f"}}}},
			err: &mongo.WriteError{
				Code: 28,
				Message: "cannot use the part (v of v.-1) to traverse the element " +
					"({v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]})",
			},
			altMessage: "cannot use path 'v.-1' to traverse the document",
		},
		"RenameUnsuitableValue": {
			id:     "array-documents-nested",
			update: bson.D{{"$rename", bson.D{{"v.0.foo.0.bar.z", "f"}}}},
			err: &mongo.WriteError{
				Code:    28,
				Message: "cannot use the part (bar of v.0.foo.0.bar.z) to traverse the element ({bar: \"hello\"})",
			},
			altMessage: "types.getByPath: can't access string by path \"z\"",
		},
		"IncTypeMismatch": {
			id:     "array-documents-nested",
			update: bson.D{{"$inc", bson.D{{"v", "string"}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: "Cannot increment with non-numeric argument: {v: \"string\"}",
			},
		},
		"IncUnsuitableValue": {
			id:     "array-documents-nested",
			update: bson.D{{"$inc", bson.D{{"v.foo", 1}}}},
			err: &mongo.WriteError{
				Code: 28,
				Message: "Cannot create field 'foo' in element " +
					"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
			},
		},
		"IncNonNumeric": {
			id:     "array-documents-nested",
			update: bson.D{{"$inc", bson.D{{"v.0.foo.0.bar", 1}}}},
			err: &mongo.WriteError{
				Code: 14,
				Message: "Cannot apply $inc to a value of non-numeric type. " +
					"{_id: \"array-documents-nested\"} has the field 'bar' of non-numeric type string",
			},
		},
		"IncInt64BadValue": {
			id:     "int64-max",
			update: bson.D{{"$inc", bson.D{{"v", math.MaxInt64}}}},
			err: &mongo.WriteError{
				Code: 2,
				Message: "Failed to apply $inc operations to current value " +
					"((NumberLong)9223372036854775807) for document {_id: \"int64-max\"}",
			},
			provider: shareddata.Int64s,
		},
		"IncInt32BadValue": {
			id:     "int32",
			update: bson.D{{"$inc", bson.D{{"v", math.MaxInt64}}}},
			err: &mongo.WriteError{
				Code: 2,
				Message: "Failed to apply $inc operations to current value " +
					"((NumberInt)42) for document {_id: \"int32\"}",
			},
			provider: shareddata.Int32s,
		},
		"MaxUnsuitableValue": {
			id:     "array-documents-nested",
			update: bson.D{{"$max", bson.D{{"v.foo", 1}}}},
			err: &mongo.WriteError{
				Code: 28,
				Message: "Cannot create field 'foo' in element " +
					"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
			},
		},
		"MinUnsuitableValue": {
			id:     "array-documents-nested",
			update: bson.D{{"$min", bson.D{{"v.foo", 1}}}},
			err: &mongo.WriteError{
				Code: 28,
				Message: "Cannot create field 'foo' in element " +
					"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
			},
		},
		"MulTypeMismatch": {
			id:     "array-documents-nested",
			update: bson.D{{"$mul", bson.D{{"v", "string"}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: "Cannot multiply with non-numeric argument: {v: \"string\"}",
			},
		},
		"MulTypeMismatchNonExistent": {
			id:     "array-documents-nested",
			update: bson.D{{"$mul", bson.D{{"non-existent", "string"}}}},
			err: &mongo.WriteError{
				Code:    14,
				Message: "Cannot multiply with non-numeric argument: {non-existent: \"string\"}",
			},
		},
		"MulUnsuitableValue": {
			id:     "array-documents-nested",
			update: bson.D{{"$mul", bson.D{{"v.foo", 1}}}},
			err: &mongo.WriteError{
				Code: 28,
				Message: "Cannot create field 'foo' in element " +
					"{v: [ { foo: [ { bar: \"hello\" }, { bar: \"world\" } ] } ]}",
			},
		},
		"MulNonNumeric": {
			id:     "array-documents-nested",
			update: bson.D{{"$mul", bson.D{{"v.0.foo.0.bar", 1}}}},
			err: &mongo.WriteError{
				Code: 14,
				Message: "Cannot apply $mul to a value of non-numeric type. " +
					"{_id: \"array-documents-nested\"} has the field 'bar' of non-numeric type string",
			},
		},
		"MulInt64BadValue": {
			id:     "int64-max",
			update: bson.D{{"$mul", bson.D{{"v", math.MaxInt64}}}},
			err: &mongo.WriteError{
				Code: 2,
				Message: "Failed to apply $mul operations to current value " +
					"((NumberLong)9223372036854775807) for document {_id: \"int64-max\"}",
			},
			provider: shareddata.Int64s,
		},
		"MulInt32BadValue": {
			id:     "int32",
			update: bson.D{{"$mul", bson.D{{"v", math.MaxInt64}}}},
			err: &mongo.WriteError{
				Code: 2,
				Message: "Failed to apply $mul operations to current value " +
					"((NumberInt)42) for document {_id: \"int32\"}",
			},
			provider: shareddata.Int32s,
		},
		"MulEmptyPath": {
			id:     "array-documents-nested",
			update: bson.D{{"$mul", bson.D{{"v.", "v"}}}},
			err: &mongo.WriteError{
				Code:    56,
				Message: "The update path 'v.' contains an empty field name, which is not allowed.",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.update, "update should be set")
			require.NotNil(t, tc.err, "err should be set")

			provider := tc.provider
			if provider == nil {
				provider = shareddata.ArrayDocuments
			}

			ctx, collection := setup.Setup(t, provider)

			res, err := collection.UpdateOne(ctx, bson.D{{"_id", tc.id}}, tc.update)

			assert.Nil(t, res)
			AssertEqualAltWriteError(t, *tc.err, tc.altMessage, err)
		})
	}
}
