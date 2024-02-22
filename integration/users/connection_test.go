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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"

	"github.com/FerretDB/FerretDB/integration/setup"
	"github.com/FerretDB/FerretDB/internal/util/testutil/testtb"
)

func TestAuthentication(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx := s.Ctx
	collection := s.Collection
	db := collection.Database()

	testCases := map[string]struct { //nolint:vet // for readability
		username       string
		password       string
		updatePassword string   // if true, the password will be updated to this one after the user is created.
		mechanisms     []string // mechanisms to use for creating user authentication

		connectionMechanism string // if set, try to establish connection with this mechanism

		userNotFound  bool
		wrongPassword bool
		topologyError bool
		errorMessage  string
	}{
		"Success": {
			username:            "username", // when using the PLAIN mechanism we must use user "username"
			password:            "password",
			mechanisms:          []string{"PLAIN"},
			connectionMechanism: "PLAIN",
		},
		"ScramSHA1": {
			username:            "scramsha1",
			password:            "password",
			mechanisms:          []string{"SCRAM-SHA-1"},
			connectionMechanism: "SCRAM-SHA-1",
		},
		"ScramSHA256": {
			username:            "scramsha256",
			password:            "password",
			mechanisms:          []string{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
		},
		"MultipleScramSHA1": {
			username:            "scramsha1multi",
			password:            "password",
			mechanisms:          []string{"SCRAM-SHA-1", "SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-1",
		},
		"MultipleScramSHA256": {
			username:            "scramsha256multi",
			password:            "password",
			mechanisms:          []string{"SCRAM-SHA-1", "SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
		},
		"ScramSHA1Updated": {
			username:            "scramsha1updated",
			password:            "pass123",
			updatePassword:      "anotherpassword",
			mechanisms:          []string{"SCRAM-SHA-1"},
			connectionMechanism: "SCRAM-SHA-1",
		},
		"ScramSHA256Updated": {
			username:            "scramsha256updated",
			password:            "pass123",
			updatePassword:      "anotherpassword",
			mechanisms:          []string{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
		},
		"NotFoundUser": {
			username:            "notfound",
			password:            "something",
			mechanisms:          []string{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
			userNotFound:        true,
			errorMessage:        "Authentication failed",
			topologyError:       true,
		},
		"BadPassword": {
			username:            "baduser",
			password:            "something",
			mechanisms:          []string{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-256",
			wrongPassword:       true,
			topologyError:       true,
			errorMessage:        "Authentication failed",
		},
		"MechanismNotSet": {
			username:            "user_mechanism_not_set",
			password:            "password",
			mechanisms:          []string{"SCRAM-SHA-256"},
			connectionMechanism: "SCRAM-SHA-1",
			topologyError:       true,
			errorMessage:        "Unable to use SCRAM-SHA-1 based authentication for user without any SCRAM-SHA-1 credentials registered",
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(tt *testing.T) {
			tt.Parallel()

			var t testtb.TB = tt

			if !tc.userNotFound {
				var (
					// Use default mechanism for MongoDB and SCRAM-SHA-256 for FerretDB as SHA-1 won't be supported as it's deprecated.
					mechanisms bson.A

					hasPlain bool
				)

				if tc.mechanisms == nil {
					if !setup.IsMongoDB(t) {
						mechanisms = append(mechanisms, "SCRAM-SHA-1", "SCRAM-SHA-256")
					}
				} else {
					mechanisms = bson.A{}

					for _, mechanism := range tc.mechanisms {
						switch mechanism {
						case "PLAIN":
							hasPlain = true
							fallthrough
						case "SCRAM-SHA-1", "SCRAM-SHA-256":
							mechanisms = append(mechanisms, mechanism)
						default:
							t.Fatalf("unimplemented mechanism %s", mechanism)
						}
					}
				}

				if hasPlain {
					setup.SkipForMongoDB(t, "PLAIN mechanism is not supported by MongoDB")
				}

				createPayload := bson.D{
					{"createUser", tc.username},
					{"roles", bson.A{}},
					{"pwd", tc.password},
					{"mechanisms", mechanisms},
				}

				err := db.RunCommand(ctx, createPayload).Err()
				require.NoErrorf(t, err, "cannot create user")
			}

			if tc.updatePassword != "" {
				updatePayload := bson.D{
					{"updateUser", tc.username},
					{"pwd", tc.updatePassword},
				}

				err := db.RunCommand(ctx, updatePayload).Err()
				require.NoErrorf(t, err, "cannot update user")
			}

			password := tc.password
			if tc.updatePassword != "" {
				password = tc.updatePassword
			}
			if tc.wrongPassword {
				password = "wrongpassword"
			}

			connectionMechanism := tc.connectionMechanism

			credential := options.Credential{
				AuthMechanism: connectionMechanism,
				AuthSource:    db.Name(),
				Username:      tc.username,
				Password:      password,
			}

			opts := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)

			client, err := mongo.Connect(ctx, opts)
			require.NoError(t, err, "cannot connect to MongoDB")

			// Ping to force connection to be established and tested.
			err = client.Ping(ctx, nil)

			if tc.errorMessage != "" {
				require.Contains(t, err.Error(), tc.errorMessage, "expected error message")
			}

			if tc.topologyError {
				var ce topology.ConnectionError
				require.ErrorAs(t, err, &ce, "expected a connection error")
				return
			}

			require.NoError(t, err, "cannot ping MongoDB")

			connCollection := client.Database(db.Name()).Collection(collection.Name())

			require.NotNil(t, connCollection, "cannot get collection")

			r, err := connCollection.InsertOne(ctx, bson.D{{"ping", "pong"}})
			require.NoError(t, err, "cannot insert document")
			id := r.InsertedID.(primitive.ObjectID)
			require.NotEmpty(t, id)

			var result bson.D
			err = connCollection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&result)
			require.NoError(t, err, "cannot find document")
			assert.Equal(t, bson.D{{"_id", id}, {"ping", "pong"}}, result)

			require.NoError(t, client.Disconnect(context.Background()))
		})
	}
}

// TestAuthenticationEnableNewAuthNoUser tests that the backend authentication
// is used when there is no user in the database. This ensures that there is
// some form of authentication even if there is no user.
func TestAuthenticationEnableNewAuthNoUserExists(t *testing.T) {
	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx := s.Ctx
	collection := s.Collection
	db := collection.Database()

	if !setup.IsMongoDB(t) {
		// drop the user created in the setup
		err := db.Client().Database("admin").RunCommand(ctx, bson.D{
			{"dropUser", "username"},
		}).Err()
		require.NoError(t, err, "cannot drop user")
	}

	testCases := map[string]struct {
		username  string
		password  string
		mechanism string

		pingErr   string
		insertErr string
	}{
		"PLAINNonExistingUser": {
			username:  "plain-user",
			password:  "whatever",
			mechanism: "PLAIN",
			insertErr: `role "plain-user" does not exist`,
		},
		"PLAINBackendUser": {
			username:  "username",
			password:  "password",
			mechanism: "PLAIN",
		},
		"SHA256NonExistingUser": {
			username:  "sha256-user",
			password:  "whatever",
			mechanism: "SCRAM-SHA-256",
			pingErr:   "Authentication failed",
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			if tc.mechanism == "PLAIN" {
				setup.SkipForMongoDB(t, "PLAIN mechanism is not supported by MongoDB")
			}

			t.Parallel()

			credential := options.Credential{
				AuthMechanism: tc.mechanism,
				AuthSource:    db.Name(),
				Username:      tc.username,
				Password:      tc.password,
			}

			opts := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)

			client, err := mongo.Connect(ctx, opts)
			require.NoError(t, err, "cannot connect to MongoDB")

			t.Cleanup(func() {
				require.NoError(t, client.Disconnect(ctx))
			})

			err = client.Ping(ctx, nil)

			if tc.pingErr != "" {
				require.ErrorContains(t, err, tc.pingErr)
				return
			}

			require.NoError(t, err, "cannot ping MongoDB")

			connCollection := client.Database(db.Name()).Collection(collection.Name())
			_, err = connCollection.InsertOne(ctx, bson.D{{"ping", "pong"}})

			if tc.insertErr != "" {
				if setup.IsSQLite(t) {
					t.Skip("SQLite does not have backend authentication")
				}

				require.ErrorContains(t, err, tc.insertErr)

				return
			}

			require.NoError(t, err, "cannot insert document")
		})
	}
}

func TestAuthenticationEnableNewAuthPLAIN(t *testing.T) {
	setup.SkipForMongoDB(t, "PLAIN mechanism is not supported by MongoDB")

	t.Parallel()

	s := setup.SetupWithOpts(t, nil)
	ctx, cName, db := s.Ctx, s.Collection.Name(), s.Collection.Database()

	err := db.RunCommand(ctx, bson.D{
		{"createUser", "plain-user"},
		{"roles", bson.A{}},
		{"pwd", "correct"},
		{"mechanisms", bson.A{"PLAIN"}},
	}).Err()
	require.NoError(t, err, "cannot create user")

	testCases := map[string]struct {
		username  string
		password  string
		mechanism string

		err string
	}{
		"Success": {
			username:  "plain-user",
			password:  "correct",
			mechanism: "PLAIN",
		},
		"BadPassword": {
			username:  "plain-user",
			password:  "wrong",
			mechanism: "PLAIN",
			err:       "AuthenticationFailed",
		},
		"NonExistentUser": {
			username:  "not-found-user",
			password:  "something",
			mechanism: "PLAIN",
			err:       "AuthenticationFailed",
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			credential := options.Credential{
				AuthMechanism: tc.mechanism,
				AuthSource:    db.Name(),
				Username:      tc.username,
				Password:      tc.password,
			}

			opts := options.Client().ApplyURI(s.MongoDBURI).SetAuth(credential)

			client, err := mongo.Connect(ctx, opts)
			require.NoError(t, err, "cannot connect to MongoDB")

			t.Cleanup(func() {
				require.NoError(t, client.Disconnect(ctx))
			})

			c := client.Database(db.Name()).Collection(cName)
			_, err = c.InsertOne(ctx, bson.D{{"ping", "pong"}})

			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}

			require.NoError(t, err, "cannot insert document")
		})
	}
}
