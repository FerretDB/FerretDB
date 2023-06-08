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

func TestDistinctErrors(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		command  any // required, command to run
		collName any // optional, defaults to coll.Name()
		filter   any // required

		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"StringFilter": {
			command: "a",
			filter:  "a",
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.query' is the wrong type 'string', expected type 'object'",
			},
			altMessage: "BSON field 'distinct.query' is the wrong type 'string', expected type 'object'",
		},
		"EmptyCollection": {
			command:  "a",
			filter:   bson.D{},
			collName: "",
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "Invalid namespace specified 'TestDistinctErrors.'",
			},
		},
		"CollectionTypeObject": {
			command:  "a",
			filter:   bson.D{},
			collName: bson.D{},
			err: &mongo.CommandError{
				Code:    73,
				Name:    "InvalidNamespace",
				Message: "collection name has invalid type object",
			},
		},
		"WrongTypeObject": {
			command: bson.D{},
			filter:  bson.D{},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.key' is the wrong type 'object', expected type 'string'",
			},
		},
		"WrongTypeArray": {
			command: bson.A{},
			filter:  bson.D{},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.key' is the wrong type 'array', expected type 'string'",
			},
		},
		"WrongTypeNumber": {
			command: int32(1),
			filter:  bson.D{},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'distinct.key' is the wrong type 'int', expected type 'string'",
			},
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.command, "command must not be nil")
			require.NotNil(t, tc.filter, "filter must not be nil")

			var collName any = coll.Name()
			if tc.collName != nil {
				collName = tc.collName
			}

			command := bson.D{{"distinct", collName}, {"key", tc.command}, {"query", tc.filter}}

			var res bson.D
			err := coll.Database().RunCommand(ctx, command).Decode(res)
			if tc.err != nil {
				assert.Nil(t, res)
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDistinctDuplicates(t *testing.T) {
	t.Parallel()

	ctx, coll := setup.Setup(t)

	docs := []any{
		bson.D{{"v", int64(42)}},
		bson.D{{"v", float64(42)}},
		bson.D{{"v", "42"}},
	}

	expected := []any{int64(42), "42"}

	_, err := coll.InsertMany(ctx, docs)
	require.NoError(t, err)

	distinct, err := coll.Distinct(ctx, "v", bson.D{})
	require.NoError(t, err)

	// We can't check the exact data types because they might be different.
	// For example, if expected is [int64(42), "42"] and distinct is [float64(42), "42"],
	// we consider them equal. If different documents use different types to store the same value
	// in the same field, it's hard to predict what type will be returned by distinct.
	// This is why we use assert.EqualValues instead of assert.Equal.
	assert.EqualValues(t, expected, distinct)
}
