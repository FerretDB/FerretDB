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

	ctx, collection := setup.Setup(t)
	client := collection.Database().Client()

	// Map of users per database suffix
	dbToUsers := map[string][]bson.D{
		"": {
			createUser("one", "pwd1"),
			createUser("two", "pwd2"),
		},
		"_example": {
			createUser("a", "password1"),
			createUser("b", "password2"),
			createUser("c", "password3"),
		},
		"_few": {
			createUser("i", "password1"),
			createUser("j", "password2"),
		},
		"_another": {
			createUser("singleuser", "123456"),
		},
	}
	// allCreatedUsers is a list of all users created in the database that you want to verify
	// are returned by the usersInfo command.
	allCreatedUsers := map[string]struct{}{
		"TestUsersinfo.one":                {},
		"TestUsersinfo.two":                {},
		"TestUsersinfo_example.a":          {},
		"TestUsersinfo_example.b":          {},
		"TestUsersinfo_example.c":          {},
		"TestUsersinfo_few.i":              {},
		"TestUsersinfo_few.j":              {},
		"TestUsersinfo_another.singleuser": {},
	}

	dbPrefix := t.Name()

	// Create users into the database to test userInfo.
	for dbSuffix, payloads := range dbToUsers {
		dbName := t.Name() + dbSuffix
		db := client.Database(dbName)

		// Clear any residual database users before recreating them.
		assert.NoError(t, db.RunCommand(ctx, bson.D{
			{"dropAllUsersFromDatabase", 1},
		}).Err())

		for _, payload := range payloads {
			err := db.RunCommand(ctx, payload).Err()
			require.NoErrorf(t, err, "cannot create user on database %q: %q", dbName, payload)
		}
	}

	testCases := map[string]struct { //nolint:vet // for readability
		dbSuffix   string
		payload    bson.D
		err        *mongo.CommandError
		altMessage string
		expected   bson.D
		hasUser    map[string]struct{}
	}{
		"NoUserFound": {
			dbSuffix: "no_users",
			payload: bson.D{
				{"usersInfo", 1},
			},
			expected: bson.D{
				{"users", bson.A{}},
				{"ok", float64(1)},
			},
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
				{"usersInfo", 1},
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
		"ForAllDBs": {
			dbSuffix: "_example",
			payload: bson.D{{
				"usersInfo", bson.D{
					{"forAllDBs", true},
				},
			}},
			hasUser: allCreatedUsers,
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

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

				assert.False(t, tc.hasUser == nil && tc.expected == nil, "set at least one expectation for %q", name)

				foundUsers[must.NotFail(actualUser.Get("_id")).(string)] = struct{}{}

				if tc.expected != nil {
					eu, err := expectedUsers.Get(i)
					require.NoError(t, err)
					expectedUser := eu.(*types.Document)

					uuid := must.NotFail(actualUser.Get("userId")).(types.Binary)
					assert.Equal(t, uuid.Subtype.String(), types.BinaryUUID.String(), "uuid subtype")
					assert.Equal(t, 16, len(uuid.B), "UUID length")
					actualUser.Remove("userId")

					actualUser.Remove("mechanisms")
					actualUser.Remove("credentials")

					testutil.AssertEqual(t, expectedUser, actualUser)
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
