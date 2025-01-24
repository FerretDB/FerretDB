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

	"github.com/FerretDB/FerretDB/v2/integration"
	"github.com/FerretDB/FerretDB/v2/integration/setup"
)

func TestUpdateUserCommand(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, db := s.Ctx, s.Collection.Database()

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
			username: "missing_fields",
			password: "pass123654",
			updatePayload: bson.D{
				{"updateUser", "missing_fields"},
			},
			err: &mongo.CommandError{
				Code:    2,
				Name:    "BadValue",
				Message: "Must specify at least one field to update in updateUser",
			},
			altMessage: "updateUser and pwd are required fields",
		},
		"UserNotFound": {
			username: "do_not_use",
			password: "pass123654",
			updatePayload: bson.D{
				{"updateUser", "not_found"},
				{"pwd", "password"},
			},
			err: &mongo.CommandError{
				Code:    11,
				Name:    "UserNotFound",
				Message: fmt.Sprintf("User not_found@%s not found", db.Name()),
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
			username: "a_user_bad_password",
			password: "password",
			updatePayload: bson.D{
				{"updateUser", "a_user_bad_password"},
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
			username: "b_user_bad_password_value",
			password: "password",
			updatePayload: bson.D{
				{"updateUser", "b_user_bad_password_value"},
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
			username: "a_user_bad_password_type",
			password: "password",
			updatePayload: bson.D{
				{"updateUser", "a_user_bad_password_type"},
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
			username: "same_password_user",
			password: "donotchange",
			updatePayload: bson.D{
				{"updateUser", "same_password_user"},
				{"pwd", "donotchange"},
			},
			expected: bson.D{
				{"users", bson.A{bson.D{
					{"_id", fmt.Sprintf("%s.same_password_user", db.Name())},
					{"user", "same_password_user"},
					{"db", db.Name()},
					{"roles", bson.A{}},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				}}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/939",
		},
		"PasswordChange": {
			username: "a_user",
			password: "password",
			updatePayload: bson.D{
				{"updateUser", "a_user"},
				{"pwd", "anewpassword"},
			},
			expected: bson.D{
				{"users", bson.A{bson.D{
					{"_id", fmt.Sprintf("%s.a_user", db.Name())},
					{"user", "a_user"},
					{"db", db.Name()},
					{"roles", bson.A{}},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				}}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/939",
		},
		"PasswordChangeWithSCRAMMechanism": {
			username:   "a_user_with_scram_mechanism",
			password:   "password",
			mechanisms: bson.A{"SCRAM-SHA-256"},
			updatePayload: bson.D{
				{"updateUser", "a_user_with_scram_mechanism"},
				{"pwd", "anewpassword"},
				{"mechanisms", bson.A{"SCRAM-SHA-256"}},
			},
			expected: bson.D{
				{"users", bson.A{bson.D{
					{"_id", fmt.Sprintf("%s.a_user_with_scram_mechanism", db.Name())},
					{"user", "a_user_with_scram_mechanism"},
					{"db", db.Name()},
					{"roles", bson.A{}},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				}}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/946",
		},
		"PasswordChangeWithBadAuthMechanism": {
			username: "a_user_with_mechanism_bad",
			password: "password",
			updatePayload: bson.D{
				{"updateUser", "a_user_with_mechanism_bad"},
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
			username: "a_user_with_no_roles",
			password: "password",
			updatePayload: bson.D{
				{"updateUser", "a_user_with_no_roles"},
				{"roles", bson.A{}},
				{"pwd", "anewpassword"},
			},
			expected: bson.D{
				{"users", bson.A{bson.D{
					{"_id", fmt.Sprintf("%s.a_user_with_no_roles", db.Name())},
					{"user", "a_user_with_no_roles"},
					{"db", db.Name()},
					{"roles", bson.A{}},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				}}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/944",
		},
		"WithComment": {
			username: "another_user",
			password: "password",
			updatePayload: bson.D{
				{"updateUser", "another_user"},
				{"pwd", "anewpassword"},
				{"comment", "test string comment"},
			},
			expected: bson.D{
				{"users", bson.A{bson.D{
					{"_id", fmt.Sprintf("%s.another_user", db.Name())},
					{"user", "another_user"},
					{"db", db.Name()},
					{"roles", bson.A{}},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				}}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/939",
		},
		"InvalidRoles": {
			username: "invalid_roles",
			password: "password",
			updatePayload: bson.D{
				{"updateUser", "invalid_roles"},
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
