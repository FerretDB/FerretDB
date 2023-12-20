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

package users

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	ctx, collection := setup.Setup(t)
	db := collection.Database()
	client := db.Client()
	users := client.Database("admin").Collection("system.users")

	// TODO https://github.com/FerretDB/FerretDB/issues/1492
	if setup.IsMongoDB(t) {
		assert.NoError(t, collection.Database().RunCommand(ctx, bson.D{
			{"dropAllUsersFromDatabase", 1},
		}).Err())
	} else {
		// Erase any previously saved user in the database.
		_, err := users.DeleteMany(ctx, bson.D{{"db", db.Name()}})
		require.NoError(t, err)
	}

	testCases := map[string]struct { //nolint:vet // for readability
		payload    bson.D
		err        *mongo.CommandError
		altMessage string
		expected   bson.D
	}{
		"Empty": {
			payload: bson.D{
				{"createUser", ""},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "User document needs 'user' field to be non-empty",
			},
		},
		"AlreadyExists": {
			payload: bson.D{
				{"createUser", "should_already_exist"},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			err: &mongo.CommandError{
				Code:    51003,
				Name:    "Location51003",
				Message: "User \"should_already_exist@TestCreateUser\" already exists",
			},
		},
		"MissingPwdOrExternal": {
			payload: bson.D{
				{"createUser", "mising_pwd_or_external"},
				{"roles", bson.A{}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Must provide a 'pwd' field for all user documents, except those with '$external' as the user's source db",
			},
		},
		"Success": {
			payload: bson.D{
				{"createUser", "success_user"},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			expected: bson.D{
				{"ok", float64(1)},
			},
		},
		"WithComment": {
			payload: bson.D{
				{"createUser", "with_comment_user"},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"comment", "test string comment"},
			},
			expected: bson.D{
				{"ok", float64(1)},
			},
		},
		"MissingRoles": {
			payload: bson.D{
				{"createUser", "missing_roles"},
				{"pwd", "password"},
			},
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: "BSON field 'createUser.roles' is missing but a required field",
			},
		},
	}

	// The subtest "AlreadyExists" tries to create the following user, which should fail with an error that the user already exists.
	// Here, we create the user for the very first time to populate the database.
	err := db.RunCommand(ctx, bson.D{
		{"createUser", "should_already_exist"},
		{"roles", bson.A{}},
		{"pwd", "password"},
	}).Err()
	require.NoError(t, err)

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var res bson.D
			err := db.RunCommand(ctx, tc.payload).Decode(&res)
			if tc.err != nil {
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			require.NoError(t, err)

			actual := integration.ConvertDocument(t, res)
			actual.Remove("$clusterTime")
			actual.Remove("operationTime")

			expected := integration.ConvertDocument(t, tc.expected)
			testutil.AssertEqual(t, expected, actual)

			payload := integration.ConvertDocument(t, tc.payload)
			// All users are created in the "admin" database.

			var rec bson.D
			err = users.FindOne(ctx, bson.D{{"user", must.NotFail(payload.Get("createUser"))}}).Decode(&rec)
			require.NoError(t, err, "user not found")

			actualRecorded := integration.ConvertDocument(t, rec)

			uuid := must.NotFail(actualRecorded.Get("userId")).(types.Binary)
			assert.Equal(t, uuid.Subtype.String(), types.BinaryUUID.String(), "uuid subtype")
			assert.Equal(t, 16, len(uuid.B), "UUID length")
			actualRecorded.Remove("userId")

			actualRecorded.Remove("credentials")

			expectedRec := integration.ConvertDocument(t, bson.D{
				{"_id", fmt.Sprintf("%s.%s", db.Name(), must.NotFail(payload.Get("createUser")))},
				{"user", must.NotFail(payload.Get("createUser"))},
				{"db", db.Name()},
				{"roles", bson.A{}},
			})

			testutil.AssertEqual(t, expectedRec, actualRecorded)
		})
	}
}
