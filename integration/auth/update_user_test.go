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

func TestUpdateUserCommand(t *testing.T) {
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

	testCases := map[string]struct { //nolint:vet // for readability
		username      string
		password      string
		mechanisms    bson.A
		updatePayload bson.D

		expected         bson.D
		err              *mongo.CommandError
		altMessage       string
		failsForFerretDB string
	}{
		"MissingFields": {
			username: userA,
			password: "pass123654",
			updatePayload: bson.D{
				{"updateUser", userA},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Must specify at least one field to update in updateUser",
			},
			altMessage: "Password cannot be empty.",
		},
		"UserNotFound": {
			username: "do_not_use",
			password: "pass123654",
			updatePayload: bson.D{
				{"updateUser", userB},
				{"pwd", "password"},
			},
			err: &mongo.CommandError{
				Code:    11,
				Name:    "UserNotFound",
				Message: fmt.Sprintf("User %s@%s not found", userB, db.Name()),
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/945",
		},
		"EmptyUsername": {
			username: "not_empty_username",
			password: "password",
			updatePayload: bson.D{
				{"updateUser", ""},
				{"pwd", "anewpassword"},
			},
			err: &mongo.CommandError{
				Code:    11,
				Name:    "UserNotFound",
				Message: fmt.Sprintf("User @%s not found", db.Name()),
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/945",
		},
		"EmptyPassword": {
			username: userC,
			password: "password",
			updatePayload: bson.D{
				{"updateUser", userC},
				{"pwd", ""},
			},
			err: &mongo.CommandError{
				Code:    50687,
				Name:    "Location50687",
				Message: "Error preflighting UTF-8 conversion: U_STRING_NOT_TERMINATED_WARNING",
			},
			altMessage:       "Password cannot be empty",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/945",
		},
		"BadPasswordValue": {
			username: userD,
			password: "password",
			updatePayload: bson.D{
				{"updateUser", userD},
				{"pwd", true},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'updateUser.pwd' is the wrong type 'bool', expected type 'string'",
			},
			altMessage: "BSON field 'pwd' is the wrong type 'bool', expected type 'string'",
		},
		"BadPasswordType": {
			username: userE,
			password: "password",
			updatePayload: bson.D{
				{"updateUser", userE},
				{"pwd", true},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'updateUser.pwd' is the wrong type 'bool', expected type 'string'",
			},
			altMessage: "BSON field 'pwd' is the wrong type 'bool', expected type 'string'",
		},
		"SamePassword": {
			username: userF,
			password: "donotchange",
			updatePayload: bson.D{
				{"updateUser", userF},
				{"pwd", "donotchange"},
			},
			expected: bson.D{
				{"users", bson.A{bson.D{
					{"_id", fmt.Sprintf("%s.%s", db.Name(), userF)},
					{"user", userF},
					{"db", db.Name()},
					{"roles", bson.A{}},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				}}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB/issues/5313",
		},
		"PasswordChange": {
			username: userG,
			password: "password",
			updatePayload: bson.D{
				{"updateUser", userG},
				{"pwd", "anewpassword"},
			},
			expected: bson.D{
				{"users", bson.A{bson.D{
					{"_id", fmt.Sprintf("%s.%s", db.Name(), userG)},
					{"user", userG},
					{"db", db.Name()},
					{"roles", bson.A{}},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				}}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB/issues/5313",
		},
		"PasswordChangeWithSCRAMMechanism": {
			username:   userH,
			password:   "password",
			mechanisms: bson.A{"SCRAM-SHA-256"},
			updatePayload: bson.D{
				{"updateUser", userH},
				{"pwd", "anewpassword"},
				{"mechanisms", bson.A{"SCRAM-SHA-256"}},
			},
			expected: bson.D{
				{"users", bson.A{bson.D{
					{"_id", fmt.Sprintf("%s.%s", db.Name(), userH)},
					{"user", userH},
					{"db", db.Name()},
					{"roles", bson.A{}},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				}}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/946",
		},
		"PasswordChangeWithBadAuthMechanism": {
			username: userI,
			password: "password",
			updatePayload: bson.D{
				{"updateUser", userI},
				{"pwd", "anewpassword"},
				{"mechanisms", bson.A{"BAD"}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Unknown auth mechanism 'BAD'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/946",
		},
		"PasswordChangeWithRoles": {
			username: userJ,
			password: "password",
			updatePayload: bson.D{
				{"updateUser", userJ},
				{"roles", bson.A{}},
				{"pwd", "anewpassword"},
			},
			expected: bson.D{
				{"users", bson.A{bson.D{
					{"_id", fmt.Sprintf("%s.%s", db.Name(), userJ)},
					{"user", userJ},
					{"db", db.Name()},
					{"roles", bson.A{}},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				}}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/944",
		},
		"WithComment": {
			username: userK,
			password: "password",
			updatePayload: bson.D{
				{"updateUser", userK},
				{"pwd", "anewpassword"},
				{"comment", "test string comment"},
			},
			expected: bson.D{
				{"users", bson.A{bson.D{
					{"_id", fmt.Sprintf("%s.%s", db.Name(), userK)},
					{"user", userK},
					{"db", db.Name()},
					{"roles", bson.A{}},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				}}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB/issues/5313",
		},
		"InvalidRoles": {
			username: userL,
			password: "password",
			updatePayload: bson.D{
				{"updateUser", userL},
				{"roles", "not-array"},
				{"pwd", "password"},
			},
			err: &mongo.CommandError{
				Code:    14,
				Name:    "TypeMismatch",
				Message: "BSON field 'updateUser.roles' is the wrong type 'string', expected type 'array'",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/944",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			var t testing.TB = tt
			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(t, tc.failsForFerretDB)
			}

			tt.Parallel()

			// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
			_ = db.RunCommand(ctx, bson.D{{"dropUser", tc.username}})

			createPayload := bson.D{
				{"createUser", tc.username},
				{"roles", bson.A{}},
				{"pwd", tc.password},
			}

			if tc.mechanisms != nil {
				createPayload = append(createPayload, bson.E{Key: "mechanisms", Value: tc.mechanisms})
			}

			err := db.RunCommand(ctx, createPayload).Err()
			require.NoError(t, err)

			var res bson.D
			err = db.RunCommand(ctx, tc.updatePayload).Decode(&res)
			if tc.err != nil {
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			require.NoError(t, err)
			integration.AssertEqualDocuments(t, bson.D{{"ok", float64(1)}}, res)

			err = db.RunCommand(ctx, bson.D{{"usersInfo", tc.username}}).Decode(&res)
			require.NoError(t, err)

			for i, v := range res {
				if v.Key != "users" {
					continue
				}

				var filteredUser bson.D

				for _, u := range v.Value.(bson.A)[0].(bson.D) {
					switch u.Key {
					case "userId":
						uuid, ok := u.Value.(primitive.Binary)
						assert.True(t, ok, "userId is not a primitive.Binary")
						assert.Equal(t, bson.TypeBinaryUUID, uuid.Subtype, "uuid subtype")
						assert.Equal(t, 16, len(uuid.Data), "UUID length")

					case "mechanisms":
						var mechanismsComparable bson.A

						for _, m := range u.Value.(bson.A) {
							if m != "SCRAM-SHA-256" {
								continue
							}

							mechanismsComparable = append(mechanismsComparable, m)
						}

						filteredUser = append(filteredUser, bson.E{Key: u.Key, Value: mechanismsComparable})
					default:
						filteredUser = append(filteredUser, u)
					}
				}

				res[i].Value = bson.A{filteredUser}
			}

			integration.AssertEqualDocuments(t, tc.expected, res)
		})
	}
}
