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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, &setup.SetupOpts{BackendOptions: &setup.BackendOpts{EnableNewAuth: true}})
	ctx, db := s.Ctx, s.Collection.Database()

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
		"EmptyPassword": {
			payload: bson.D{
				{"createUser", "empty_password_user"},
				{"roles", bson.A{}},
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
			payload: bson.D{
				{"createUser", "empty_password_user"},
				{"roles", bson.A{}},
				{"pwd", "pass\x00word"},
			},
			err: &mongo.CommandError{
				Code:    50692,
				Name:    "Location50692",
				Message: "Error preflighting normalization: U_STRINGPREP_PROHIBITED_ERROR",
			},
		},
		"BadPasswordType": {
			payload: bson.D{
				{"createUser", "empty_password_user"},
				{"roles", bson.A{}},
				{"pwd", true},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'createUser.pwd' is the wrong type 'bool', expected type 'string'",
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
		"EmptyMechanism": {
			payload: bson.D{
				{"createUser", "empty_mechanism_user"},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "mechanisms field must not be empty",
			},
		},
		"BadAuthMechanism": {
			payload: bson.D{
				{"createUser", "success_user_with_plain"},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{"PLAIN", "BAD"}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Unknown auth mechanism 'PLAIN'",
			},
			altMessage: "Unknown auth mechanism 'BAD'",
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
		"FailWithPLAIN": {
			payload: bson.D{
				{"createUser", "success_user_with_plain"},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{"PLAIN"}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Unknown auth mechanism 'PLAIN'",
			},
		},
		"SuccessWithSCRAMSHA1": {
			payload: bson.D{
				{"createUser", "success_user_with_scram_sha_1"},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{"SCRAM-SHA-1"}},
			},
			expected: bson.D{
				{"ok", float64(1)},
			},
		},
		"SuccessWithSCRAMSHA256": {
			payload: bson.D{
				{"createUser", "success_user_with_scram_sha_256"},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{"SCRAM-SHA-256"}},
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

			payload := integration.ConvertDocument(t, tc.payload)
			if payload.Has("mechanisms") {
				setup.SkipForMongoDB(t, "PLAIN credentials are only supported via LDAP (PLAIN) by MongoDB Enterprise")
			}

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

			username := must.NotFail(payload.Get("createUser"))

			var usersInfo bson.D
			err = db.RunCommand(ctx, bson.D{{"usersInfo", username}, {"showCredentials", true}}).Decode(&usersInfo)
			require.NoError(t, err)

			actualRecorded := integration.ConvertDocument(t, usersInfo)
			users := must.NotFail(actualRecorded.Get("users")).(*types.Array)
			user := must.NotFail(users.Get(0)).(*types.Document)

			uuid := must.NotFail(user.Get("userId")).(types.Binary)
			assert.Equal(t, uuid.Subtype.String(), types.BinaryUUID.String(), "uuid subtype")
			assert.Equal(t, 16, len(uuid.B), "UUID length")
			user.Remove("userId")

			if payload.Has("mechanisms") {
				payloadMechanisms := must.NotFail(payload.Get("mechanisms")).(*types.Array)

				if payloadMechanisms.Contains("PLAIN") {
					assertPlainCredentials(t, "PLAIN", must.NotFail(user.Get("credentials")).(*types.Document))
				}

				if payloadMechanisms.Contains("SCRAM-SHA-1") {
					assertSCRAMSHA1Credentials(t, "SCRAM-SHA-1", must.NotFail(user.Get("credentials")).(*types.Document))
				}

				if payloadMechanisms.Contains("SCRAM-SHA-256") {
					assertSCRAMSHA256Credentials(t, "SCRAM-SHA-256", must.NotFail(user.Get("credentials")).(*types.Document))
				}
			}

			user.Remove("mechanisms")
			user.Remove("credentials")

			expectedRec := integration.ConvertDocument(t, bson.D{
				{"_id", fmt.Sprintf("%s.%s", db.Name(), must.NotFail(payload.Get("createUser")))},
				{"user", must.NotFail(payload.Get("createUser"))},
				{"db", db.Name()},
				{"roles", bson.A{}},
			})

			testutil.AssertEqual(t, expectedRec, user)
		})
	}
}

// assertPlainCredentials checks if the credential is a valid PLAIN credential.
func assertPlainCredentials(t testtb.TB, key string, cred *types.Document) {
	t.Helper()

	require.True(t, cred.Has(key), "missing credential %q", key)

	c := must.NotFail(cred.Get(key)).(*types.Document)

	assert.Equal(t, must.NotFail(c.Get("algo")), "PBKDF2-HMAC-SHA256")
	assert.NotEmpty(t, must.NotFail(c.Get("iterationCount")))
	assert.NotEmpty(t, must.NotFail(c.Get("hash")))
	assert.NotEmpty(t, must.NotFail(c.Get("salt")))
}

// assertSCRAMSHA1Credentials checks if the credential is a valid SCRAM-SHA-1 credential.
func assertSCRAMSHA1Credentials(t testtb.TB, key string, cred *types.Document) {
	t.Helper()

	require.True(t, cred.Has(key), "missing credential %q", key)

	c := must.NotFail(cred.Get(key)).(*types.Document)

	assert.Equal(t, must.NotFail(c.Get("iterationCount")), int32(10000))
	assert.NotEmpty(t, must.NotFail(c.Get("salt")).(string))
	assert.NotEmpty(t, must.NotFail(c.Get("serverKey")).(string))
	assert.NotEmpty(t, must.NotFail(c.Get("storedKey")).(string))
}

// assertSCRAMSHA256Credentials checks if the credential is a valid SCRAM-SHA-256 credential.
func assertSCRAMSHA256Credentials(t testtb.TB, key string, cred *types.Document) {
	t.Helper()

	require.True(t, cred.Has(key), "missing credential %q", key)

	c := must.NotFail(cred.Get(key)).(*types.Document)

	assert.Equal(t, must.NotFail(c.Get("iterationCount")), int32(15000))
	assert.NotEmpty(t, must.NotFail(c.Get("salt")).(string))
	assert.NotEmpty(t, must.NotFail(c.Get("serverKey")).(string))
	assert.NotEmpty(t, must.NotFail(c.Get("storedKey")).(string))
}
