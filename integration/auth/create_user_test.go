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

package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/FerretDB/FerretDB/v2/internal/util/testutil"

	"github.com/FerretDB/FerretDB/v2/integration"
	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestCreateUserCommand(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, db := s.Ctx, s.Collection.Database()

	emptyPasswordUser := testutil.UserName(t)
	badPasswordUser := testutil.UserName(t)
	badPasswordTypeUser := testutil.UserName(t)
	existingUser := testutil.UserName(t)
	emptyMechanismUser := testutil.UserName(t)
	badMechanismUser := testutil.UserName(t)
	missingPasswordUser := testutil.UserName(t)
	successUser := testutil.UserName(t)
	successSCRAMSHA256User := testutil.UserName(t)
	withCommentUser := testutil.UserName(t)
	missingRolesUser := testutil.UserName(t)
	nullRolesUser := testutil.UserName(t)
	invalidRolesUser := testutil.UserName(t)

	testCases := map[string]struct { //nolint:vet // for readability
		payload bson.D
		user    string

		expected                      bson.D
		expectedMechanisms            bson.A
		expectedCredentialsComparable bson.D // field keys without values are compared for `salt`, `serverKey` and `storedKey` fields
		err                           *mongo.CommandError
		altMessage                    string
		failsForFerretDB              string
	}{
		"Empty": {
			payload: bson.D{
				{"createUser", ""},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			user: "",
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "User document needs 'user' field to be non-empty",
			},
			altMessage: "'createUser' is a required field.",
		},
		"EmptyPassword": {
			payload: bson.D{
				{"createUser", emptyPasswordUser},
				{"roles", bson.A{}},
				{"pwd", ""},
			},
			user: emptyPasswordUser,
			err: &mongo.CommandError{
				Code:    50687,
				Name:    "Location50687",
				Message: "Error preflighting UTF-8 conversion: U_STRING_NOT_TERMINATED_WARNING",
			},
			altMessage:       "Password cannot be empty",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/934",
		},
		"BadPasswordValue": {
			payload: bson.D{
				{"createUser", badPasswordUser},
				{"roles", bson.A{}},
				{"pwd", "pass\x00word"},
			},
			user: badPasswordUser,
			err: &mongo.CommandError{
				Code:    50692,
				Name:    "Location50692",
				Message: "Error preflighting normalization: U_STRINGPREP_PROHIBITED_ERROR",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/941",
		},
		"BadPasswordType": {
			payload: bson.D{
				{"createUser", badPasswordTypeUser},
				{"roles", bson.A{}},
				{"pwd", true},
			},
			user: badPasswordTypeUser,
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'createUser.pwd' is the wrong type 'bool', expected type 'string'",
			},
			altMessage: "BSON field 'pwd' is the wrong type 'bool', expected type 'string'",
		},
		"AlreadyExists": {
			payload: bson.D{
				{"createUser", existingUser},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			user: "should_already_exist",
			err: &mongo.CommandError{
				Code:    51003,
				Name:    "Location51003",
				Message: fmt.Sprintf("User \"%s@%s\" already exists", existingUser, db.Name()),
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/934",
		},
		"EmptyMechanism": {
			payload: bson.D{
				{"createUser", emptyMechanismUser},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{}},
			},
			user: emptyMechanismUser,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "mechanisms field must not be empty",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/913",
		},
		"BadAuthMechanism": {
			payload: bson.D{
				{"createUser", badMechanismUser},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{"BAD"}},
			},
			user: badMechanismUser,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Unknown auth mechanism 'BAD'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/913",
		},
		"MissingPwdOrExternal": {
			payload: bson.D{
				{"createUser", missingPasswordUser},
				{"roles", bson.A{}},
			},
			user: missingPasswordUser,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Must provide a 'pwd' field for all user documents, except those with '$external' as the user's source db",
			},
			altMessage: "'createUser', 'roles' and 'pwd' are required fields.",
		},
		"Success": {
			payload: bson.D{
				{"createUser", successUser},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			user:               successUser,
			expectedMechanisms: bson.A{"SCRAM-SHA-256"},
			expected: bson.D{
				{"ok", float64(1)},
			},
			expectedCredentialsComparable: bson.D{
				{"SCRAM-SHA-256", bson.D{
					{"iterationCount", int32(15000)},
					{Key: "salt"},
					{Key: "storedKey"},
					{Key: "serverKey"},
				}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB/issues/5313",
		},
		"SuccessWithSCRAMSHA256": {
			payload: bson.D{
				{"createUser", successSCRAMSHA256User},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{"SCRAM-SHA-256"}},
			},
			user:               successSCRAMSHA256User,
			expectedMechanisms: bson.A{"SCRAM-SHA-256"},
			expected: bson.D{
				{"ok", float64(1)},
			},
			expectedCredentialsComparable: bson.D{
				{"SCRAM-SHA-256", bson.D{
					{"iterationCount", int32(15000)},
					{Key: "salt"},
					{Key: "storedKey"},
					{Key: "serverKey"},
				}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB/issues/5313",
		},
		"WithComment": {
			payload: bson.D{
				{"createUser", withCommentUser},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"comment", "test string comment"},
			},
			user:               withCommentUser,
			expectedMechanisms: bson.A{"SCRAM-SHA-256"},
			expected: bson.D{
				{"ok", float64(1)},
			},
			expectedCredentialsComparable: bson.D{
				{"SCRAM-SHA-256", bson.D{
					{"iterationCount", int32(15000)},
					{Key: "salt"},
					{Key: "storedKey"},
					{Key: "serverKey"},
				}},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB/issues/5313",
		},
		"MissingRoles": {
			payload: bson.D{
				{"createUser", missingRolesUser},
				{"pwd", "password"},
			},
			user: missingRolesUser,
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: "BSON field 'createUser.roles' is missing but a required field",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/934",
		},
		"NilRoles": {
			payload: bson.D{
				{"createUser", nullRolesUser},
				{"roles", nil},
				{"pwd", "password"},
			},
			user: nullRolesUser,
			err: &mongo.CommandError{
				Code:    10065,
				Name:    "Location10065",
				Message: "invalid parameter: expected an object (roles)",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/934",
		},
		"InvalidRoles": {
			payload: bson.D{
				{"createUser", invalidRolesUser},
				{"roles", "not-array"},
				{"pwd", "password"},
			},
			user: invalidRolesUser,
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'createUser.roles' is the wrong type 'string', expected type 'array'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/934",
		},
	}

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	_ = db.RunCommand(ctx, bson.D{{"dropUser", existingUser}})

	// The subtest "AlreadyExists" tries to create the following user, which should fail with an error that the user already exists.
	// Here, we create the user for the very first time to populate the database.
	err := db.RunCommand(ctx, bson.D{
		{"createUser", existingUser},
		{"roles", bson.A{}},
		{"pwd", "password"},
	}).Err()
	require.NoError(t, err)

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			var t testing.TB = tt
			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(t, tc.failsForFerretDB)
			}

			tt.Parallel()

			// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
			if tc.user != "should_already_exist" {
				_ = db.RunCommand(ctx, bson.D{{"dropUser", tc.user}})
			}

			var res bson.D
			err := db.RunCommand(ctx, tc.payload).Decode(&res)
			if tc.err != nil {
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			require.NoError(t, err)

			integration.AssertEqualDocuments(t, tc.expected, res)

			err = db.RunCommand(ctx, bson.D{{"usersInfo", tc.user}, {"showCredentials", true}}).Decode(&res)
			require.NoError(t, err)

			var resComparable bson.D

			for _, v := range res {
				if v.Key != "users" {
					resComparable = append(resComparable, v)
					continue
				}

				var userComparable bson.D

				for _, u := range v.Value.(bson.A)[0].(bson.D) {
					switch u.Key {
					case "userId":
						uuid, ok := u.Value.(primitive.Binary)
						assert.True(t, ok, "userId is not a primitive.Binary")
						assert.Equal(t, bson.TypeBinaryUUID, uuid.Subtype, "uuid subtype")
						assert.Equal(t, 16, len(uuid.Data), "UUID length")
						userComparable = append(userComparable, bson.E{Key: u.Key})
					case "credentials":
						var credentials bson.D

						for _, c := range u.Value.(bson.D) {
							if c.Key != "SCRAM-SHA-256" {
								continue
							}

							var mechanism bson.D

							for _, m := range c.Value.(bson.D) {
								switch m.Key {
								case "salt", "serverKey", "storedKey":
									assert.NotEmpty(t, m.Value.(string))
									mechanism = append(mechanism, bson.E{Key: m.Key})
								default:
									mechanism = append(mechanism, m)
								}
							}

							credentials = append(credentials, bson.E{Key: c.Key, Value: mechanism})
						}

						userComparable = append(userComparable, bson.E{Key: u.Key, Value: credentials})
					case "mechanisms":
						var mechanismsComparable bson.A

						for _, m := range u.Value.(bson.A) {
							if m != "SCRAM-SHA-256" {
								continue
							}

							mechanismsComparable = append(mechanismsComparable, m)
						}

						userComparable = append(userComparable, bson.E{Key: u.Key, Value: mechanismsComparable})
					default:
						userComparable = append(userComparable, u)
					}
				}

				resComparable = append(resComparable, bson.E{Key: v.Key, Value: bson.A{userComparable}})
			}

			expectedUsersInfo := bson.D{
				{
					"users", bson.A{
						bson.D{
							{"_id", fmt.Sprintf("%s.%s", db.Name(), tc.user)},
							{Key: "userId"},
							{"user", tc.user},
							{"db", db.Name()},
							{"credentials", tc.expectedCredentialsComparable},
							{"roles", bson.A{}},
							{"mechanisms", tc.expectedMechanisms},
						},
					},
				},
				{"ok", float64(1)},
			}

			integration.AssertEqualDocuments(t, expectedUsersInfo, resComparable)
		})
	}
}
