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

	dbToUsers := []struct {
		dbSuffix       string
		payloads       []bson.D
		skipForMongoDB bool
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
			dbSuffix: "nomongo",
			payloads: []bson.D{
				{
					{"createUser", "WithPLAIN"},
					{"roles", bson.A{}},
					{"pwd", "pwd1"},
					{"mechanisms", bson.A{"PLAIN"}},
				},
			},
			skipForMongoDB: true,
		},
	}

	dbPrefix := testutil.DatabaseName(t)

	for _, inserted := range dbToUsers {
		if inserted.skipForMongoDB && setup.IsMongoDB(t) {
			continue
		}

		dbName := testutil.DatabaseName(t) + inserted.dbSuffix
		db := client.Database(dbName)

		for _, payload := range inserted.payloads {
			err := db.RunCommand(ctx, payload).Err()
			require.NoErrorf(t, err, "cannot create user on database %q: %q", dbName, payload)
		}
	}

	testCases := map[string]struct { //nolint:vet // for readability
		dbSuffix        string
		showCredentials bool // showCredentials should also be set on the payload
		payload         bson.D
		err             *mongo.CommandError
		altMessage      string
		expected        bson.D
		hasUser         map[string]struct{}
		skipForMongoDB  string // optional, skip test for MongoDB backend with a specific reason
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
		"WithPLAIN": {
			dbSuffix:        "nomongo",
			showCredentials: true,
			payload: bson.D{
				{"usersInfo", "WithPLAIN"},
				{"showCredentials", true},
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
			skipForMongoDB: "Only MongoDB Enterprise offers PLAIN",
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.skipForMongoDB != "" {
				setup.SkipForMongoDB(t, tc.skipForMongoDB)
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

				if tc.showCredentials {
					if !setup.IsMongoDB(t) {
						cred, ok := actualUser.Get("credentials")
						assert.Nil(t, ok, "credentials not found")
						assertPlainCredentials(t, "PLAIN", cred.(*types.Document))
					}
				} else {
					assert.False(t, actualUser.Has("credentials"))
				}

				foundUsers[must.NotFail(actualUser.Get("_id")).(string)] = struct{}{}

				uuid := must.NotFail(actualUser.Get("userId")).(types.Binary)
				assert.Equal(t, uuid.Subtype.String(), types.BinaryUUID.String(), "uuid subtype")
				assert.Equal(t, 16, len(uuid.B), "UUID length")
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
