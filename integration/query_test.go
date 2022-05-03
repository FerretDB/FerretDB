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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration/shareddata"
)

func TestUnknownFilterOperator(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars)

	filter := bson.D{{"value", bson.D{{"$someUnknownOperator", 42}}}}
	errExpected := mongo.CommandError{Code: 2, Name: "BadValue", Message: "unknown operator: $someUnknownOperator"}
	_, err := collection.Find(ctx, filter)
	AssertEqualError(t, errExpected, err)
}

func TestQueryCount(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command  any
		response int32
	}{
		"CountAllDocuments": {
			command:  bson.D{{"count", collection.Name()}},
			response: 43,
		},
		"CountExactlyOneDocument": {
			command: bson.D{
				{"count", collection.Name()},
				{"query", bson.D{{"value", true}}},
			},
			response: 1,
		},
		"CountArrays": {
			command: bson.D{
				{"count", collection.Name()},
				{"query", bson.D{{"value", bson.D{{"$type", "array"}}}}},
			},
			response: 4,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.NoError(t, err)

			m := actual.Map()

			assert.Equal(t, 1.0, m["ok"])

			keys := CollectKeys(t, actual)
			assert.Contains(t, keys, "n")
			assert.Equal(t, tc.response, m["n"])
		})
	}
}

func TestQueryBadFindType(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command bson.D
		err     *mongo.CommandError
	}{
		"Document": {
			command: bson.D{
				{"find", bson.D{}},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type object",
			},
		},
		"Array": {
			command: bson.D{
				{"find", primitive.A{}},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type array",
			},
		},
		"Double": {
			command: bson.D{
				{"find", 3.14},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type double",
			},
		},
		"DoubleWhole": {
			command: bson.D{
				{"find", 42.0},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type double",
			},
		},
		"Binary": {
			command: bson.D{
				{"find", primitive.Binary{}},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type binData",
			},
		},
		"ObjectID": {
			command: bson.D{
				{"find", primitive.ObjectID{}},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type objectId",
			},
		},
		"Bool": {
			command: bson.D{
				{"find", true},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type bool",
			},
		},
		"Date": {
			command: bson.D{
				{"find", time.Now()},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type date",
			},
		},
		"Null": {
			command: bson.D{
				{"find", nil},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type null",
			},
		},
		"Regex": {
			command: bson.D{
				{"find", primitive.Regex{Pattern: "/foo/"}},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type regex",
			},
		},
		"Int": {
			command: bson.D{
				{"find", int32(42)},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type int",
			},
		},
		"Timestamp": {
			command: bson.D{
				{"find", primitive.Timestamp{}},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type timestamp",
			},
		},
		"Long": {
			command: bson.D{
				{"find", int64(42)},
				{"projection", bson.D{{"value", "some"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "collection name has invalid type long",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.Error(t, err)
			AssertEqualError(t, *tc.err, err)
		})
	}
}

func TestQueryBadSortType(t *testing.T) {
	t.Parallel()
	ctx, collection := setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command bson.D
		err     *mongo.CommandError
		message string
	}{
		"BadSortTypeDouble": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"value", "some"}}},
				{"sort", 42.13},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Expected field sortto be of type object",
			},
			message: "Expected field sort to be of type object",
		},
		"BadSortType": {
			command: bson.D{
				{"find", collection.Name()},
				{"projection", bson.D{{"value", "some"}}},
				{"sort", "123"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "Expected field sortto be of type object",
			},
			message: "Expected field sort to be of type object",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var actual bson.D
			err := collection.Database().RunCommand(ctx, tc.command).Decode(&actual)
			require.Error(t, err)
			AssertEqualErrorMessage(t, *tc.err, tc.message, err)
		})
	}
}
