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

	userA := testutil.UserName(t)
	userB := testutil.UserName(t)
	userC := testutil.UserName(t)
	userD := testutil.UserName(t)
	userE := testutil.UserName(t)
	userF := testutil.UserName(t)
	userG := testutil.UserName(t)
	userH := testutil.UserName(t)
	userI := testutil.UserName(t)
	userJ := testutil.UserName(t)
	userK := testutil.UserName(t)
	userL := testutil.UserName(t)
	userM := testutil.UserName(t)

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
				{"createUser", userA},
				{"roles", bson.A{}},
				{"pwd", ""},
			},
			user: userA,
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
				{"createUser", userB},
				{"roles", bson.A{}},
				{"pwd", "pass\x00word"},
			},
			user: userB,
			err: &mongo.CommandError{
				Code:    50692,
				Name:    "Location50692",
				Message: "Error preflighting normalization: U_STRINGPREP_PROHIBITED_ERROR",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/941",
		},
		"BadPasswordType": {
			payload: bson.D{
				{"createUser", userC},
				{"roles", bson.A{}},
				{"pwd", true},
			},
			user: userC,
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'createUser.pwd' is the wrong type 'bool', expected type 'string'",
			},
			altMessage: "BSON field 'pwd' is the wrong type 'bool', expected type 'string'",
		},
		"AlreadyExists": {
			payload: bson.D{
				{"createUser", userD},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			user: "should_already_exist",
			err: &mongo.CommandError{
				Code:    51003,
				Name:    "Location51003",
				Message: fmt.Sprintf("User \"%s@%s\" already exists", userD, db.Name()),
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/934",
		},
		"EmptyMechanism": {
			payload: bson.D{
				{"createUser", userE},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{}},
			},
			user: userE,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "mechanisms field must not be empty",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/913",
		},
		"BadAuthMechanism": {
			payload: bson.D{
				{"createUser", userF},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{"BAD"}},
			},
			user: userF,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Unknown auth mechanism 'BAD'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/913",
		},
		"MissingPwdOrExternal": {
			payload: bson.D{
				{"createUser", userG},
				{"roles", bson.A{}},
			},
			user: userG,
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Must provide a 'pwd' field for all user documents, except those with '$external' as the user's source db",
			},
			altMessage: "'createUser', 'roles' and 'pwd' are required fields.",
		},
		"Success": {
			payload: bson.D{
				{"createUser", userH},
				{"roles", bson.A{}},
				{"pwd", "password"},
			},
			user:               userH,
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
				{"createUser", userI},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"mechanisms", bson.A{"SCRAM-SHA-256"}},
			},
			user:               userI,
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
				{"createUser", userJ},
				{"roles", bson.A{}},
				{"pwd", "password"},
				{"comment", "test string comment"},
			},
			user:               userJ,
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
				{"createUser", userK},
				{"pwd", "password"},
			},
			user: userK,
			err: &mongo.CommandError{
				Code:    40414,
				Name:    "Location40414",
				Message: "BSON field 'createUser.roles' is missing but a required field",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/934",
		},
		"NilRoles": {
			payload: bson.D{
				{"createUser", userL},
				{"roles", nil},
				{"pwd", "password"},
			},
			user: userL,
			err: &mongo.CommandError{
				Code:    10065,
				Name:    "Location10065",
				Message: "invalid parameter: expected an object (roles)",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/934",
		},
		"InvalidRoles": {
			payload: bson.D{
				{"createUser", userM},
				{"roles", "not-array"},
				{"pwd", "password"},
			},
			user: userM,
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'createUser.roles' is the wrong type 'string', expected type 'array'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/934",
		},
	}

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	_ = db.RunCommand(ctx, bson.D{{"dropUser", userD}})

	// The subtest "AlreadyExists" tries to create the following user, which should fail with an error that the user already exists.
	// Here, we create the user for the very first time to populate the database.
	err := db.RunCommand(ctx, bson.D{
		{"createUser", userD},
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
