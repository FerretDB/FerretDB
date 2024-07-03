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

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"github.com/FerretDB/FerretDB/internal/util/testutil"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"

	"github.com/FerretDB/FerretDB/integration"
	"github.com/FerretDB/FerretDB/integration/setup"
)

// createUser creates a bson.D command payload to create an user with the given username and password.
func createUser(username, password string) bson.D {
	return bson.D{
		{"createUser", username},
		{"roles", bson.A{}},
		{"pwd", password},
	}
}

func TestUsersinfo(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, collection := s.Ctx, s.Collection
	client := collection.Database().Client()

	dbToUsers := []struct {
		dbSuffix string
		payloads []bson.D
	}{
		{
			dbSuffix: "",
			payloads: []bson.D{
				createUser("one", "pwd1"),
				createUser("two", "pwd2"),
			},
		},
		{
			dbSuffix: "_example",
			payloads: []bson.D{
				createUser("a", "password1"),
				createUser("b", "password2"),
				createUser("c", "password3"),
			},
		},
		{
			dbSuffix: "_few",
			payloads: []bson.D{
				createUser("i", "password1"),
				createUser("j", "password2"),
			},
		},
		{
			dbSuffix: "_another",
			payloads: []bson.D{
				createUser("singleuser", "123456"),
			},
		},
		{
			dbSuffix: "allbackends",
			payloads: []bson.D{
				{
					{"createUser", "WithSCRAMSHA1"},
					{"roles", bson.A{}},
					{"pwd", "pwd1"},
					{"mechanisms", bson.A{"SCRAM-SHA-1"}},
				},
			},
		},
		{
			dbSuffix: "allbackends",
			payloads: []bson.D{
				{
					{"createUser", "WithSCRAMSHA256"},
					{"roles", bson.A{}},
					{"pwd", "pwd1"},
					{"mechanisms", bson.A{"SCRAM-SHA-256"}},
				},
			},
		},
	}

	dbPrefix := testutil.DatabaseName(t)

	// Create users in the databases.
	// Do not create users that require the PLAIN authentication mechanism when using MongoDB as
	// only its Enterprise version supports it.
	for _, inserted := range dbToUsers {
		dbName := testutil.DatabaseName(t) + inserted.dbSuffix
		db := client.Database(dbName)

		t.Cleanup(func() {
			db.RunCommand(ctx, bson.D{{"dropAllUsersFromDatabase", 1}})
		})

		for _, payload := range inserted.payloads {
			err := db.RunCommand(ctx, payload).Err()
			require.NoErrorf(t, err, "cannot create user on database %q: %q", dbName, payload)
		}
	}

	testCases := map[string]struct { //nolint:vet // for readability
		dbSuffix        string
		payload         bson.D
		err             *mongo.CommandError
		altMessage      string
		expected        bson.D
		hasUser         map[string]struct{}
		showCredentials []string // showCredentials list the credentials types expected to be returned
		failsForMongoDB string
	}{
		"NoUserFound": {
			dbSuffix: "no_users",
			payload: bson.D{
				{"usersInfo", int64(1)},
			},
			expected: bson.D{
				{"users", bson.A{}},
				{"ok", float64(1)},
			},
		},
		"Nil": {
			dbSuffix: "",
			payload: bson.D{
				{"usersInfo", nil},
			},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName must be either a string or an object",
				Name:    "BadValue",
			},
		},
		"UnknownField": {
			dbSuffix: "",
			payload: bson.D{
				// Note: if user is passed here, this test will fail only on MongoDB.
				{"usersInfo", bson.D{{"foo", "bar"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName contains an unknown field named: 'foo",
				Name:    "BadValue",
			},
			altMessage: "UserName must contain a field named: user",
		},
		"Default": {
			dbSuffix: "",
			payload: bson.D{
				{"usersInfo", "one"},
			},
			expected: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", "TestUsersinfo.one"},
						{"user", "one"},
						{"db", "TestUsersinfo"},
						{"roles", bson.A{}},
					},
				}},
				{"ok", float64(1)},
			},
		},
		"WithSCRAMSHA1": {
			dbSuffix: "allbackends",
			payload: bson.D{
				{"usersInfo", "WithSCRAMSHA1"},
				{"showCredentials", true},
			},
			showCredentials: []string{"SCRAM-SHA-1"},
			expected: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", "TestUsersinfo.WithSCRAMSHA1"},
						{"user", "scramsha1"},
						{"db", "TestUsersinfo"},
						{"roles", bson.A{}},
					},
				}},
				{"ok", float64(1)},
			},
		},
		"WithSCRAMSHA256": {
			dbSuffix: "allbackends",
			payload: bson.D{
				{"usersInfo", "WithSCRAMSHA256"},
				{"showCredentials", true},
			},
			showCredentials: []string{"SCRAM-SHA-256"},
			expected: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", "TestUsersinfo.WithSCRAMSHA256"},
						{"user", "scramsha256"},
						{"db", "TestUsersinfo"},
						{"roles", bson.A{}},
					},
				}},
				{"ok", float64(1)},
			},
		},
		"FromSameDatabase": {
			dbSuffix: "_example",
			payload: bson.D{{
				"usersInfo", bson.A{
					"a", "b",
				},
			}},
			expected: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", "TestUsersinfo_example.a"},
						{"user", "a"},
						{"db", "TestUsersinfo_example"},
						{"roles", bson.A{}},
					},
					bson.D{
						{"_id", "TestUsersinfo_example.b"},
						{"user", "b"},
						{"db", "TestUsersinfo_example"},
						{"roles", bson.A{}},
					},
				}},
				{"ok", float64(1)},
			},
		},
		"FromSameDatabaseWithMissingUser": {
			dbSuffix: "_example",
			payload: bson.D{{
				"usersInfo", bson.A{
					"a", "b", "missing",
				},
			}},
			expected: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", "TestUsersinfo_example.a"},
						{"user", "a"},
						{"db", "TestUsersinfo_example"},
						{"roles", bson.A{}},
					},
					bson.D{
						{"_id", "TestUsersinfo_example.b"},
						{"user", "b"},
						{"db", "TestUsersinfo_example"},
						{"roles", bson.A{}},
					},
				}},
				{"ok", float64(1)},
			},
		},
		"FromAnotherDatabase": {
			dbSuffix: "another_database",
			payload: bson.D{{
				"usersInfo", bson.D{
					{"user", "one"},
					{"db", "TestUsersinfo"},
				},
			}},
			expected: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", "TestUsersinfo.one"},
						{"user", "one"},
						{"db", "TestUsersinfo"},
						{"roles", bson.A{}},
					},
				}},
				{"ok", float64(1)},
			},
		},
		"BadType": {
			dbSuffix: "_example",
			payload: bson.D{{
				"usersInfo", true,
			}},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName must be either a string or an object",
				Name:    "BadValue",
			},
		},
		"BadTypeUsername": {
			dbSuffix: "another_database",
			payload: bson.D{{
				"usersInfo", bson.D{
					{"user", 123},
					{"db", "TestUsersinfo_example"},
				},
			}},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName must contain a string field named: user. But, has type int",
				Name:    "BadValue",
			},
			altMessage: "UserName must contain a string field named: user. But, has type int32",
		},
		"BadTypeDB": {
			dbSuffix: "another_database",
			payload: bson.D{{
				"usersInfo", bson.D{
					{"user", "one"},
					{"db", 123},
				},
			}},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName must contain a string field named: db. But, has type int",
				Name:    "BadValue",
			},
			altMessage: "UserName must contain a string field named: db. But, has type int32",
		},
		"FromOthersMultipleDatabases": {
			dbSuffix: "another_database",
			payload: bson.D{{
				"usersInfo", bson.A{
					bson.D{
						{"user", "one"},
						{"db", "TestUsersinfo"},
					},
					bson.D{
						{"user", "i"},
						{"db", "TestUsersinfo_few"},
					},
				},
			}},
			expected: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", "TestUsersinfo_few.i"},
						{"user", "i"},
						{"db", "TestUsersinfo_few"},
						{"roles", bson.A{}},
					},
					bson.D{
						{"_id", "TestUsersinfo.one"},
						{"user", "one"},
						{"db", "TestUsersinfo"},
						{"roles", bson.A{}},
					},
				}},
				{"ok", float64(1)},
			},
		},
		"FromMixed": {
			dbSuffix: "_few",
			payload: bson.D{{
				"usersInfo", bson.A{
					bson.D{
						{"user", "one"},
						{"db", "TestUsersinfo"},
					},
					"i",
				},
			}},
			expected: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", "TestUsersinfo_few.i"},
						{"user", "i"},
						{"db", "TestUsersinfo_few"},
						{"roles", bson.A{}},
					},
					bson.D{
						{"_id", "TestUsersinfo.one"},
						{"user", "one"},
						{"db", "TestUsersinfo"},
						{"roles", bson.A{}},
					},
				}},
				{"ok", float64(1)},
			},
		},
		"Many": {
			dbSuffix: "_example",
			payload: bson.D{
				{"usersInfo", int64(1)},
			},
			expected: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", "TestUsersinfo_example.a"},
						{"user", "a"},
						{"db", "TestUsersinfo_example"},
						{"roles", bson.A{}},
					},
					bson.D{
						{"_id", "TestUsersinfo_example.b"},
						{"user", "b"},
						{"db", "TestUsersinfo_example"},
						{"roles", bson.A{}},
					},
					bson.D{
						{"_id", "TestUsersinfo_example.c"},
						{"user", "c"},
						{"db", "TestUsersinfo_example"},
						{"roles", bson.A{}},
					},
				}},
				{"ok", float64(1)},
			},
		},
		"Multiple": {
			dbSuffix: "_example",
			payload: bson.D{
				{"usersInfo", bson.A{
					bson.D{{"user", "b"}, {"db", "TestUsersinfo_example"}},
					bson.D{{"user", "c"}, {"db", "TestUsersinfo_example"}},
				}},
			},
			expected: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", "TestUsersinfo_example.b"},
						{"user", "b"},
						{"db", "TestUsersinfo_example"},
						{"roles", bson.A{}},
					},
					bson.D{
						{"_id", "TestUsersinfo_example.c"},
						{"user", "c"},
						{"db", "TestUsersinfo_example"},
						{"roles", bson.A{}},
					},
				}},
				{"ok", float64(1)},
			},
		},
		"MissingDBFieldName": {
			dbSuffix: "_example",
			payload: bson.D{
				{"usersInfo", bson.A{
					bson.D{{"user", "missing_db"}},
				}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName must contain a field named: db",
				Name:    "BadValue",
			},
		},
		"MissingUserFieldName": {
			dbSuffix: "_example",
			payload: bson.D{
				{"usersInfo", bson.A{
					bson.D{{"db", "TestUsersinfo_example"}},
				}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName must contain a field named: user",
				Name:    "BadValue",
			},
		},
		"ForAllDBs": {
			dbSuffix: "_example",
			payload: bson.D{{
				"usersInfo", bson.D{
					{"forAllDBs", true},
				},
			}},
			hasUser: map[string]struct{}{
				"TestUsersinfo.one":                {},
				"TestUsersinfo.two":                {},
				"TestUsersinfo_example.a":          {},
				"TestUsersinfo_example.b":          {},
				"TestUsersinfo_example.c":          {},
				"TestUsersinfo_few.i":              {},
				"TestUsersinfo_few.j":              {},
				"TestUsersinfo_another.singleuser": {},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testtb.TB = tt

			if tc.failsForMongoDB != "" {
				t = setup.FailsForMongoDB(t, tc.failsForMongoDB)
			}

			var res bson.D
			dbName := dbPrefix + tc.dbSuffix
			err := client.Database(dbName).RunCommand(ctx, tc.payload).Decode(&res)
			if tc.err != nil {
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			require.NoError(t, err)

			actual := integration.ConvertDocument(t, res)
			actual.Remove("$clusterTime")
			actual.Remove("operationTime")

			actualUsers := must.NotFail(actual.Get("users")).(*types.Array)

			var (
				expected      *types.Document
				expectedUsers *types.Array
			)

			if tc.expected != nil {
				expected = integration.ConvertDocument(t, tc.expected)
				expectedUsers = must.NotFail(expected.Get("users")).(*types.Array)
				require.Equal(t, expectedUsers.Len(), actualUsers.Len(), "users length")
			}

			foundUsers := map[string]struct{}{}
			for i := 0; i < actualUsers.Len(); i++ {
				au, err := actualUsers.Get(i)
				require.NoError(t, err)
				actualUser := au.(*types.Document)

				require.True(t, (tc.hasUser == nil) != (tc.expected == nil))

				id, err := actualUser.Get("_id")
				require.NoError(t, err)

				// when `forAllDBs` is set true, it may contain more users from other databases,
				// so we check expected users were found rather than exact match
				foundUsers[id.(string)] = struct{}{}

				userIDV, err := actualUser.Get("userId")
				require.NoError(t, err)

				userID := userIDV.(types.Binary)
				assert.Equal(t, userID.Subtype.String(), types.BinaryUUID.String(), "uuid subtype")
				assert.Equal(t, 16, len(userID.B), "UUID length")

				if tc.showCredentials == nil {
					assert.False(t, actualUser.Has("credentials"))

					continue
				}

				credV, err := actualUser.Get("credentials")
				require.NoError(t, err)

				cred := credV.(*types.Document)

				for _, typ := range tc.showCredentials {
					switch typ {
					case "SCRAM-SHA-1":
						assertSCRAMSHA1Credentials(t, "SCRAM-SHA-1", cred)
					case "SCRAM-SHA-256":
						assertSCRAMSHA256Credentials(t, "SCRAM-SHA-256", cred)
					}
				}
			}

			if tc.hasUser != nil {
				assert.GreaterOrEqual(t, len(foundUsers), len(tc.hasUser), "users length min")
				for u := range tc.hasUser {
					_, ok := foundUsers[u]
					assert.True(t, ok, "user %q not found", u)
				}
			}

			if expected != nil {
				// Then, compare any remaining field in the document.
				actual.Remove("users")
				expected.Remove("users")
				testutil.AssertEqual(t, expected, actual)
			}
		})
	}
}
