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

// TestDeleteSimple checks simple cases of doc deletion.
func TestDeleteSimple(t *testing.T) {
	t.Parallel()
	ctx, collection := setup.Setup(t, shareddata.Scalars, shareddata.Composites)

	for name, tc := range map[string]struct {
		collection    string
		expectedCount int64
	}{
		"DeleteOne": {
			collection:    collection.Name(),
			expectedCount: 1,
		},
		"DeleteFromNonExistingCollection": {
			collection:    "doesnotexist",
			expectedCount: 0,
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cursor, err := collection.Database().Collection(tc.collection).DeleteOne(ctx, bson.D{})
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCount, cursor.DeletedCount)
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		deletes bson.A // required, set to deletes parameter

		err        *mongo.CommandError // optional, expected error from MongoDB
		altMessage string              // optional, alternative error message for FerretDB, ignored if empty
		skip       string              // optional, skip test with a specified reason
	}{
		"QueryNotSet": {
			deletes: bson.A{bson.D{}},
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: "BSON field 'delete.deletes.q' is missing but a required field",
			},
		},
		"NotSet": {
			deletes: bson.A{bson.D{{"q", bson.D{{"v", "foo"}}}}},
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: "BSON field 'delete.deletes.limit' is missing but a required field",
			},
		},
		"ValidFloat": {
			deletes: bson.A{bson.D{{"q", bson.D{{"v", "foo"}}}, {"limit", 1.00}}},
		},
		"ValidString": {
			deletes: bson.A{bson.D{{"q", bson.D{{"v", "foo"}}}, {"limit", "1"}}},
			skip:    "https://github.com/FerretDB/FerretDB/issues/1089",
		},
		"InvalidFloat": {
			deletes: bson.A{bson.D{{"q", bson.D{{"v", "foo"}}}, {"limit", 42.13}}},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "The limit field in delete objects must be 0 or 1. Got 42.13",
			},
			altMessage: "The 'delete.deletes.limit' field must be 0 or 1. Got 42.13",
		},
		"InvalidInt": {
			deletes: bson.A{bson.D{{"q", bson.D{{"v", "foo"}}}, {"limit", 100}}},
			err: &mongo.CommandError{
				Code:    9,
				Name:    "FailedToParse",
				Message: "The limit field in delete objects must be 0 or 1. Got 100",
			},
			altMessage: "The 'delete.deletes.limit' field must be 0 or 1. Got 100",
		},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skip != "" {
				t.Skip(tc.skip)
			}

			t.Parallel()

			require.NotNil(t, tc.deletes, "deletes must not be nil")

			ctx, collection := setup.Setup(t)

			var res bson.D
			err := collection.Database().RunCommand(ctx, bson.D{
				{"delete", collection.Name()},
				{"deletes", tc.deletes},
			}).Decode(&res)
			if tc.err != nil {
				assert.Nil(t, res)
				AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)

				return
			}

			require.NoError(t, err)
		})
	}
}
