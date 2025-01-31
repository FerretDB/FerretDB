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

// createUser creates a bson.D command payload to create an user with the given username and password.
func createUser(username, password string) bson.D {
	return bson.D{
		{"createUser", username},
		{"roles", bson.A{}},
		{"pwd", password},
	}
}

func TestUsersInfoCommand(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, collection := s.Ctx, s.Collection
	client := collection.Database().Client()

	dbPrefix := testutil.DatabaseName(t)

	dbNames := struct {
		A string
		B string
		C string
		D string
		E string
	}{
		A: dbPrefix,
		B: dbPrefix + "_example",
		C: dbPrefix + "_few",
		D: dbPrefix + "_another",
		E: dbPrefix + "allbackends",
	}

	dbToUsers := []struct {
		dbName   string
		payloads []bson.D
	}{
		{
			dbName: dbNames.A,
			payloads: []bson.D{
				createUser("one", "pwd1"),
				createUser("two", "pwd2"),
			},
		},
		{
			dbName: dbNames.B,
			payloads: []bson.D{
				createUser("a", "password1"),
				createUser("b", "password2"),
				createUser("c", "password3"),
			},
		},
		{
			dbName: dbNames.C,
			payloads: []bson.D{
				createUser("i", "password1"),
				createUser("j", "password2"),
			},
		},
		{
			dbName: dbNames.D,
			payloads: []bson.D{
				createUser("singleuser", "123456"),
			},
		},
		{
			dbName: dbNames.E,
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

	// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
	testUsernames := map[string]string{
		"one":             dbNames.A,
		"two":             dbNames.A,
		"a":               dbNames.B,
		"b":               dbNames.B,
		"c":               dbNames.B,
		"i":               dbNames.C,
		"j":               dbNames.C,
		"singleuser":      dbNames.D,
		"WithSCRAMSHA256": dbNames.E,
	}

	for username, dbName := range testUsernames {
		db := client.Database(dbName)
		_ = db.RunCommand(ctx, bson.D{{"dropUser", username}})
	}

	for _, inserted := range dbToUsers {
		db := client.Database(inserted.dbName)

		t.Cleanup(func() {
			db.RunCommand(ctx, bson.D{{"dropAllUsersFromDatabase", 1}})
		})

		for _, payload := range inserted.payloads {
			err := db.RunCommand(ctx, payload).Err()
			require.NoError(t, err)
		}
	}

	testCases := map[string]struct { //nolint:vet // for readability
		dbName  string
		payload bson.D

		expectedComparable bson.D // field keys without values are compared for `userId`, `salt`, `serverKey` and `storedKey` fields
		err                *mongo.CommandError
		altMessage         string
		failsForFerretDB   string
	}{
		"NoUserFound": {
			dbName: "no_users",
			payload: bson.D{
				{"usersInfo", int64(1)},
			},
			expectedComparable: bson.D{
				{"users", bson.A{}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"Nil": {
			dbName: dbNames.A,
			payload: bson.D{
				{"usersInfo", nil},
			},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName must be either a string or an object",
				Name:    "BadValue",
			},
			altMessage: "Unusupported value for usersInfo",
		},
		"UnknownField": {
			dbName: dbNames.A,
			payload: bson.D{
				// Note: if user is passed here, this test will fail only on MongoDB.
				{"usersInfo", bson.D{{"foo", "bar"}}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName contains an unknown field named: 'foo",
				Name:    "BadValue",
			},
			altMessage:       "UserName must contain a field named: user",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"Default": {
			dbName: dbNames.A,
			payload: bson.D{
				{"usersInfo", "one"},
			},
			expectedComparable: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", fmt.Sprintf("%s.one", dbNames.A)},
						{Key: "userId"},
						{"user", "one"},
						{"db", dbNames.A},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
				}},
				{"ok", float64(1)},
			},
			// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
			// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/961
			// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/962
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/939",
		},
		"WithSCRAMSHA256": {
			dbName: dbNames.E,
			payload: bson.D{
				{"usersInfo", "WithSCRAMSHA256"},
				{"showCredentials", true},
			},
			expectedComparable: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", fmt.Sprintf("%s.WithSCRAMSHA256", dbNames.E)},
						{Key: "userId"},
						{"user", "WithSCRAMSHA256"},
						{"db", dbNames.E},
						{"credentials", bson.D{
							{"SCRAM-SHA-256", bson.D{
								{"iterationCount", int32(15000)},
								{Key: "salt"},
								{Key: "storedKey"},
								{Key: "serverKey"},
							}},
						}},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-256"}},
					},
				}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/939",
		},
		"FromSameDatabase": {
			dbName: dbNames.B,
			payload: bson.D{{
				"usersInfo", bson.A{
					"a", "b",
				},
			}},
			expectedComparable: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", fmt.Sprintf("%s.a", dbNames.B)},
						{Key: "userId"},
						{"user", "a"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.b", dbNames.B)},
						{Key: "userId"},
						{"user", "b"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
				}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"FromSameDatabaseWithMissingUser": {
			dbName: dbNames.B,
			payload: bson.D{{
				"usersInfo", bson.A{
					"a", "b", "missing",
				},
			}},
			expectedComparable: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", fmt.Sprintf("%s.a", dbNames.B)},
						{Key: "userId"},
						{"user", "a"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.b", dbNames.B)},
						{Key: "userId"},
						{"user", "b"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
				}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"FromAnotherDatabase": {
			dbName: dbNames.D,
			payload: bson.D{{
				"usersInfo", bson.D{
					{"user", "one"},
					{"db", dbNames.A},
				},
			}},
			expectedComparable: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", fmt.Sprintf("%s.one", dbNames.A)},
						{Key: "userId"},
						{"user", "one"},
						{"db", dbNames.A},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
				}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/864",
		},
		"BadType": {
			dbName: dbNames.B,
			payload: bson.D{{
				"usersInfo", true,
			}},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName must be either a string or an object",
				Name:    "BadValue",
			},
			altMessage: "Unusupported value for usersInfo",
		},
		"BadTypeUsername": {
			dbName: dbNames.D,
			payload: bson.D{{
				"usersInfo", bson.D{
					{"user", 123},
					{"db", dbNames.B},
				},
			}},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName must contain a string field named: user. But, has type int",
				Name:    "BadValue",
			},
			altMessage: fmt.Sprintf("Unsupported value specified for db : %s", dbNames.B),
		},
		"BadTypeDB": {
			dbName: dbNames.D,
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
			altMessage:       "UserName must contain a string field named: db. But, has type int32",
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/964",
		},
		"FromOthersMultipleDatabases": {
			dbName: dbNames.C,
			payload: bson.D{{
				"usersInfo", bson.A{
					bson.D{
						{"user", "one"},
						{"db", dbNames.A},
					},
					bson.D{
						{"user", "i"},
						{"db", dbNames.C},
					},
				},
			}},
			expectedComparable: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", fmt.Sprintf("%s.i", dbNames.C)},
						{Key: "userId"},
						{"user", "i"},
						{"db", dbNames.C},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.one", dbNames.A)},
						{Key: "userId"},
						{"user", "one"},
						{"db", dbNames.A},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
				}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"FromMixed": {
			dbName: dbNames.C,
			payload: bson.D{{
				"usersInfo", bson.A{
					bson.D{
						{"user", "one"},
						{"db", dbNames.A},
					},
					"i",
				},
			}},
			expectedComparable: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", fmt.Sprintf("%s.i", dbNames.C)},
						{Key: "userId"},
						{"user", "i"},
						{"db", dbNames.C},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.one", dbNames.A)},
						{Key: "userId"},
						{"user", "one"},
						{"db", dbNames.A},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
				}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"SingleDatabaseInt": {
			dbName: dbNames.B,
			payload: bson.D{
				{"usersInfo", int32(1)},
			},
			expectedComparable: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", fmt.Sprintf("%s.a", dbNames.B)},
						{Key: "userId"},
						{"user", "a"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.b", dbNames.B)},
						{Key: "userId"},
						{"user", "b"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.c", dbNames.B)},
						{Key: "userId"},
						{"user", "c"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
				}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/939",
		},
		"SingleDatabaseLong": {
			dbName: "no_users",
			payload: bson.D{
				{"usersInfo", int64(1)},
			},
			expectedComparable: bson.D{
				{"users", bson.A{}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"SingleDatabaseFloat": {
			dbName: "no_users",
			payload: bson.D{
				{"usersInfo", float64(1)},
			},
			expectedComparable: bson.D{
				{"users", bson.A{}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"Multiple": {
			dbName: dbNames.B,
			payload: bson.D{
				{"usersInfo", bson.A{
					bson.D{{"user", "b"}, {"db", dbNames.B}},
					bson.D{{"user", "c"}, {"db", dbNames.B}},
				}},
			},
			expectedComparable: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", fmt.Sprintf("%s.b", dbNames.B)},
						{Key: "userId"},
						{"user", "b"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.c", dbNames.B)},
						{Key: "userId"},
						{"user", "c"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
				}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"EmptyUsers": {
			dbName: dbNames.B,
			payload: bson.D{{
				"usersInfo", bson.A{},
			}},
			err: &mongo.CommandError{
				Code:    2,
				Message: "$and/$or/$nor must be a nonempty array",
				Name:    "BadValue",
			},
			altMessage: "Unusupported value for usersInfo",
		},
		"MissingDBFieldName": {
			dbName: dbNames.B,
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
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"MissingUserFieldName": {
			dbName: dbNames.B,
			payload: bson.D{
				{"usersInfo", bson.A{
					bson.D{{"db", dbNames.B}},
				}},
			},
			err: &mongo.CommandError{
				Code:    2,
				Message: "UserName must contain a field named: user",
				Name:    "BadValue",
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
		"ForAllDBs": {
			dbName: dbNames.B,
			payload: bson.D{
				{"usersInfo", bson.D{
					{"forAllDBs", true},
				}},
			},
			expectedComparable: bson.D{
				{"users", bson.A{
					bson.D{
						{"_id", fmt.Sprintf("%s.WithSCRAMSHA256", dbNames.E)},
						{Key: "userId"},
						{"user", "WithSCRAMSHA256"},
						{"db", dbNames.E},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.a", dbNames.B)},
						{Key: "userId"},
						{"user", "a"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.b", dbNames.B)},
						{Key: "userId"},
						{"user", "b"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.c", dbNames.B)},
						{Key: "userId"},
						{"user", "c"},
						{"db", dbNames.B},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.i", dbNames.C)},
						{Key: "userId"},
						{"user", "i"},
						{"db", dbNames.C},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.j", dbNames.C)},
						{Key: "userId"},
						{"user", "j"},
						{"db", dbNames.C},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.one", dbNames.A)},
						{Key: "userId"},
						{"user", "one"},
						{"db", dbNames.A},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.singleuser", dbNames.D)},
						{Key: "userId"},
						{"user", "singleuser"},
						{"db", dbNames.D},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
					bson.D{
						{"_id", fmt.Sprintf("%s.two", dbNames.A)},
						{Key: "userId"},
						{"user", "two"},
						{"db", dbNames.A},
						{Key: "roles"},
						{"mechanisms", bson.A{"SCRAM-SHA-1", "SCRAM-SHA-256"}},
					},
				}},
				{"ok", float64(1)},
			},
			failsForFerretDB: "https://github.com/FerretDB/FerretDB-DocumentDB/issues/963",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			var t testing.TB = tt
			if tc.failsForFerretDB != "" {
				t = setup.FailsForFerretDB(tt, tc.failsForFerretDB)
			}

			tt.Parallel()

			var res bson.D
			err := client.Database(tc.dbName).RunCommand(ctx, tc.payload).Decode(&res)
			if tc.err != nil {
				integration.AssertEqualAltCommandError(t, *tc.err, tc.altMessage, err)
				return
			}

			require.NoError(t, err)

			var resComparable bson.D

			for _, v := range res {
				if v.Key != "users" {
					resComparable = append(resComparable, v)
					continue
				}

				var usersComparable bson.A

				for _, u := range v.Value.(bson.A) {
					var userComparable bson.D

					var userCreatedByOtherTests bool

					for _, field := range u.(bson.D) {
						switch field.Key {
						case "userId":
							uuid, ok := field.Value.(primitive.Binary)
							assert.True(t, ok, "userId is not a primitive.Binary")
							assert.Equal(t, bson.TypeBinaryUUID, uuid.Subtype, "uuid subtype")
							assert.Equal(t, 16, len(uuid.Data), "UUID length")
							userComparable = append(userComparable, bson.E{Key: field.Key})
						case "credentials":
							var credentials bson.D

							for _, c := range field.Value.(bson.D) {
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

							userComparable = append(userComparable, bson.E{Key: field.Key, Value: credentials})
						case "user":
							// use username to check if user was created by this test
							// TODO https://github.com/FerretDB/FerretDB-DocumentDB/issues/864
							_, ok := testUsernames[field.Value.(string)]
							if !ok {
								userCreatedByOtherTests = true
							}

							userComparable = append(userComparable, field)
						case "roles":
							userComparable = append(userComparable, bson.E{Key: field.Key})
						default:
							userComparable = append(userComparable, field)
						}
					}

					if userCreatedByOtherTests {
						continue
					}

					usersComparable = append(usersComparable, userComparable)
				}

				resComparable = append(resComparable, bson.E{Key: v.Key, Value: usersComparable})
			}

			integration.AssertEqualDocuments(t, tc.expectedComparable, resComparable)
		})
	}
}
