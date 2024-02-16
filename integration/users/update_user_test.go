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

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx := s.Ctx

	db, _ := createUserTestRunnerUser(t, s)

	testCases := map[string]struct { //nolint:vet // for readability
		createPayload bson.D
		updatePayload bson.D

		expected   bson.D
		err        *mongo.CommandError
		altMessage string

		skipForMongoDB string
	}{
		"MissingFields": {
			createPayload: bson.D{
				{"createUser", "missing_fields"},
				{"roles", bson.A{}},
				{"pwd", "pass123654"},
			},
			updatePayload: bson.D{
				{"updateUser", "missing_fields"},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Must specify at least one field to update in updateUser",
			},
		},
		"UserNotFound": {
			createPayload: bson.D{
				{"createUser", "do_not_use"},
				{"roles", bson.A{}},
				{"pwd", "pass123654"},
			},
			updatePayload: bson.D{
				{"updateUser", "not_found"},
				{"pwd", "password"},
			},
			err: &mongo.CommandError{
				Code:    11,
				Name:    "UserNotFound",
				Message: "User not_found@TestUpdateUser not found",
			},
		},
		"EmptyUsername": {
			createPayload: bson.D{
				{"createUser", "not_empty_username"},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			updatePayload: bson.D{
				{"updateUser", ""},
				{"pwd", "anewpassword"},
			},
			err: &mongo.CommandError{
				Code:    11,
				Name:    "UserNotFound",
				Message: "User @TestUpdateUser not found",
			},
		},
		"EmptyPassword": {
			createPayload: bson.D{
				{"createUser", "a_user_bad_password"},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			updatePayload: bson.D{
				{"updateUser", "a_user_bad_password"},
				{"pwd", ""},
			},
			err: &mongo.CommandError{
				Code:    50687,
				Name:    "Location50687",
				Message: "Error preflighting UTF-8 conversion: U_STRING_NOT_TERMINATED_WARNING",
			},
			altMessage: "Password cannot be empty",
		},
		"BadPasswordValue": {
			createPayload: bson.D{
				{"createUser", "b_user_bad_password_value"},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			updatePayload: bson.D{
				{"updateUser", "b_user_bad_password_value"},
				{"pwd", true},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'updateUser.pwd' is the wrong type 'bool', expected type 'string'",
			},
		},
		"BadPasswordType": {
			createPayload: bson.D{
				{"createUser", "a_user_bad_password_type"},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			updatePayload: bson.D{
				{"updateUser", "a_user_bad_password_type"},
				{"pwd", true},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'updateUser.pwd' is the wrong type 'bool', expected type 'string'",
			},
			altMessage: "Password cannot be empty",
		},
		"SamePassword": {
			createPayload: bson.D{
				{"createUser", "same_password_user"},
				{"roles", bson.A{}},
				{"pwd", "donotchange"},
			},
			updatePayload: bson.D{
				{"updateUser", "same_password_user"},
				{"pwd", "donotchange"},
			},
			expected: bson.D{
				{"_id", "TestUpdateUser.same_password_user"},
				{"user", "same_password_user"},
				{"db", "TestUpdateUser"},
				{"roles", bson.A{}},
			},
		},
		"PasswordChange": {
			createPayload: bson.D{
				{"createUser", "a_user"},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			updatePayload: bson.D{
				{"updateUser", "a_user"},
				{"pwd", "anewpassword"},
			},
			expected: bson.D{
				{"_id", "TestUpdateUser.a_user"},
				{"user", "a_user"},
				{"db", "TestUpdateUser"},
				{"roles", bson.A{}},
			},
		},
		"PasswordChangeWithMechanism": {
			createPayload: bson.D{
				{"createUser", "a_user_with_mechanism"},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{"PLAIN"}},
			},
			updatePayload: bson.D{
				{"updateUser", "a_user_with_mechanism"},
				{"pwd", "anewpassword"},
			},
			expected: bson.D{
				{"_id", "TestUpdateUser.a_user_with_mechanism"},
				{"user", "a_user_with_mechanism"},
				{"db", "TestUpdateUser"},
				{"roles", bson.A{}},
			},
			skipForMongoDB: "MongoDB decommissioned support to PLAIN auth",
		},
		"PasswordChangeWithSCRAMMechanism": {
			createPayload: bson.D{
				{"createUser", "a_user_with_scram_mechanism"},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{"SCRAM-SHA-256"}},
			},
			updatePayload: bson.D{
				{"updateUser", "a_user_with_scram_mechanism"},
				{"pwd", "anewpassword"},
				{"mechanisms", bson.A{"SCRAM-SHA-256"}},
			},
			expected: bson.D{
				{"_id", "TestUpdateUser.a_user_with_scram_mechanism"},
				{"user", "a_user_with_scram_mechanism"},
				{"db", "TestUpdateUser"},
				{"roles", bson.A{}},
			},
		},
		"PasswordChangeWithBadAuthMechanism": {
			createPayload: bson.D{
				{"createUser", "a_user_with_mechanism_bad"},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			updatePayload: bson.D{
				{"updateUser", "a_user_with_mechanism_bad"},
				{"pwd", "anewpassword"},
				{"mechanisms", bson.A{"PLAIN", "BAD"}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Unknown auth mechanism 'BAD'",
			},
			skipForMongoDB: "MongoDB decommissioned support to PLAIN auth",
		},
		"PasswordChangeWithRoles": {
			createPayload: bson.D{
				{"createUser", "a_user_with_no_roles"},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			updatePayload: bson.D{
				{"updateUser", "a_user_with_no_roles"},
				{"roles", bson.A{}},
				{"pwd", "anewpassword"},
			},
			expected: bson.D{
				{"_id", "TestUpdateUser.a_user_with_no_roles"},
				{"user", "a_user_with_no_roles"},
				{"db", "TestUpdateUser"},
				{"roles", bson.A{}},
			},
		},
		"WithComment": {
			createPayload: bson.D{
				{"createUser", "another_user"},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			updatePayload: bson.D{
				{"updateUser", "another_user"},
				{"pwd", "anewpassword"},
				{"comment", "test string comment"},
			},
			expected: bson.D{
				{"_id", "TestUpdateUser.another_user"},
				{"user", "another_user"},
				{"db", "TestUpdateUser"},
				{"roles", bson.A{}},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.skipForMongoDB != "" {
				setup.SkipForMongoDB(t, tc.skipForMongoDB)
			}

			t.Parallel()

			createPayloadDoc := integration.ConvertDocument(t, tc.createPayload)

			err := db.RunCommand(ctx, tc.createPayload).Err()
			require.NoErrorf(t, err, "cannot create user: %q", tc.createPayload)

			err = db.RunCommand(ctx, tc.updatePayload).Err()
			if tc.err != nil {
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			require.NoErrorf(t, err, "cannot update user: %q", tc.updatePayload)

			user := must.NotFail(createPayloadDoc.Get("createUser")).(string)

			var res bson.D
			err = db.RunCommand(ctx, bson.D{{"usersInfo", user}}).Decode(&res)
			require.NoError(t, err)

			actual := integration.ConvertDocument(t, res)
			actualUser := must.NotFail(must.NotFail(actual.Get("users")).(*types.Array).Get(0)).(*types.Document)
			actualUser.Remove("mechanisms")

			uuid := must.NotFail(actualUser.Get("userId")).(types.Binary)
			assert.Equal(t, uuid.Subtype.String(), types.BinaryUUID.String(), "uuid subtype")
			assert.Equal(t, 16, len(uuid.B), "UUID length")
			actualUser.Remove("userId")

			expected := integration.ConvertDocument(t, tc.expected)
			testutil.AssertEqual(t, expected, actualUser)
		})
	}
}
